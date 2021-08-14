package ddl

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/bokwoon95/sq"
)

type MigrationMode int64

// note: do not serialize these values! the underlying value of the constant
// may change and if you depend on the value, it will break your application.
// Only ever refer to these constants by name and never by value.
const (
	CreateMissing MigrationMode = 1 << iota
	UpdateExisting
	DropExtraneous
	DropCascade
)

type Migration struct {
	Dialect             string
	CurrentSchema       string
	CreateSchemaCmds    []CreateSchemaCommand
	CreateExtensionCmds []CreateExtensionCommand
	CreateFunctionCmds  []CreateFunctionCommand
	CreateTableCmds     []CreateTableCommand
	AlterTableCmds      []AlterTableCommand // add & alter columns | add & alter constraints | add indexes
	CreateViewCmds      []CreateViewCommand
	CreateIndexCmds     []CreateIndexCommand
	CreateTriggerCmds   []CreateTriggerCommand
	AddForeignKeyCmds   []AlterTableCommand
	DropViewCmds        []DropViewCommand
	DropTableCmds       []DropTableCommand
	DropTriggerCmds     []DropTriggerCommand
	DropIndexCmds       []DropIndexCommand
	AlterTableDropCmds  []AlterTableCommand
	DropFunctionCmds    []DropFunctionCommand
	DropExtensionCmds   []DropExtensionCommand
}

func AutoMigrate(dialect string, db sq.DB, migrationMode MigrationMode, databaseMetadataOpts ...DatabaseMetadataOption) error {
	gotDBMetadata, err := NewDatabaseMetadata(dialect, WithDB(db, nil))
	if err != nil {
		return fmt.Errorf("introspecting db: %w", err)
	}
	wantDBMetadata, err := NewDatabaseMetadata(dialect, databaseMetadataOpts...)
	if err != nil {
		return fmt.Errorf("building db metadata: %w", err)
	}
	migration, err := Migrate(migrationMode, gotDBMetadata, wantDBMetadata)
	if err != nil {
		return fmt.Errorf("building migration: %w", err)
	}
	err = migration.Exec(db)
	if err != nil {
		return fmt.Errorf("executing migration: %w", err)
	}
	return nil
}

func Migrate(mode MigrationMode, gotDBMetadata, wantDBMetadata DatabaseMetadata) (*Migration, error) {
	m := &Migration{
		Dialect:       gotDBMetadata.Dialect,
		CurrentSchema: gotDBMetadata.CurrentSchema,
	}
	if gotDBMetadata.Dialect == "" && wantDBMetadata.Dialect == "" {
		return m, fmt.Errorf("dialect missing")
	}
	if gotDBMetadata.Dialect != "" && wantDBMetadata.Dialect != "" && gotDBMetadata.Dialect != wantDBMetadata.Dialect {
		return m, fmt.Errorf("dialect mismatch: got %s, want %s", gotDBMetadata.Dialect, wantDBMetadata.Dialect)
	}
	if m.Dialect == "" {
		m.Dialect = wantDBMetadata.Dialect
	}
	if mode&CreateMissing != 0 || mode&UpdateExisting != 0 {
		if m.Dialect == sq.DialectPostgres && mode&CreateMissing != 0 {
			for _, wantExtension := range wantDBMetadata.Extensions {
				if n := gotDBMetadata.CachedExtensionPosition(wantExtension); n >= 0 {
					continue
				}
				createExtensionCmd := CreateExtensionCommand{
					CreateIfNotExists: true,
					Extension:         wantExtension,
				}
				m.CreateExtensionCmds = append(m.CreateExtensionCmds, createExtensionCmd)
			}
		}
		for _, wantSchema := range wantDBMetadata.Schemas {
			if wantSchema.Ignore {
				continue
			}
			if wantSchema.SchemaName == "" {
				wantSchema.SchemaName = m.CurrentSchema
			}
			gotSchema := Schema{
				SchemaName: wantSchema.SchemaName,
			}
			if n := gotDBMetadata.CachedSchemaPosition(wantSchema.SchemaName); n >= 0 {
				gotSchema = gotDBMetadata.Schemas[n]
			} else if mode&CreateMissing != 0 {
				if wantSchema.SchemaName != "" {
					createSchemaCmd := CreateSchemaCommand{
						CreateIfNotExists: true,
						SchemaName:        wantSchema.SchemaName,
					}
					m.CreateSchemaCmds = append(m.CreateSchemaCmds, createSchemaCmd)
				}
			}
			for _, wantTable := range wantSchema.Tables {
				if wantTable.Ignore {
					continue
				}
				tableExists := false
				createTableCmd := CreateTableCommand{
					CreateIfNotExists:  true,
					IncludeConstraints: true,
				}
				alterTableCmd := AlterTableCommand{
					TableSchema:   wantTable.TableSchema,
					TableName:     wantTable.TableName,
					AlterIfExists: m.Dialect == sq.DialectPostgres,
				}
				addForeignKeyCmd := alterTableCmd
				gotTable := Table{
					TableSchema: wantTable.TableSchema,
					TableName:   wantTable.TableName,
				}
				if n := gotSchema.CachedTablePosition(wantTable.TableName); n >= 0 {
					tableExists = true
					gotTable = gotSchema.Tables[n]
				} else if mode&CreateMissing != 0 {
					createTableCmd.Table = wantTable
				}
				if tableExists {
					for _, wantColumn := range wantTable.Columns {
						if wantColumn.Ignore {
							continue
						}
						gotColumn := Column{
							TableSchema: wantColumn.TableSchema,
							TableName:   wantColumn.TableName,
							ColumnName:  wantColumn.ColumnName,
						}
						if n := gotTable.CachedColumnPosition(wantColumn.ColumnName); n >= 0 {
							if mode&UpdateExisting != 0 {
								gotColumn = gotTable.Columns[n]
								alterColumnCmd, isDifferent := diffColumn(m.Dialect, gotColumn, wantColumn)
								if isDifferent {
									if m.Dialect == sq.DialectMySQL {
										alterColumnCmd.Column = wantColumn
									}
									alterTableCmd.AlterColumnCmds = append(alterTableCmd.AlterColumnCmds, alterColumnCmd)
								}
							}
						} else if mode&CreateMissing != 0 {
							addColumnCmd := AddColumnCommand{
								AddIfNotExists: m.Dialect == sq.DialectPostgres,
								Column:         wantColumn,
							}
							alterTableCmd.AddColumnCmds = append(alterTableCmd.AddColumnCmds, addColumnCmd)
						}
					}
				}
				if m.Dialect != sq.DialectSQLite {
					for _, wantConstraint := range wantTable.Constraints {
						if wantConstraint.Ignore {
							continue
						}
						if n := gotTable.CachedConstraintPosition(wantConstraint.ConstraintName); n >= 0 {
							continue
						}
						addConstraintCmd := AddConstraintCommand{
							Constraint: wantConstraint,
						}
						if addConstraintCmd.Constraint.ConstraintType == FOREIGN_KEY {
							addForeignKeyCmd.AddConstraintCmds = append(addForeignKeyCmd.AddConstraintCmds, addConstraintCmd)
						} else if tableExists || !createTableCmd.IncludeConstraints {
							alterTableCmd.AddConstraintCmds = append(alterTableCmd.AddConstraintCmds, addConstraintCmd)
						}
					}
				}
				for _, wantIndex := range wantTable.Indexes {
					if wantIndex.Ignore {
						continue
					}
					if n := gotTable.CachedIndexPosition(wantIndex.IndexName); n >= 0 {
						continue
					}
					createIndexCmd := CreateIndexCommand{
						CreateIfNotExists: m.Dialect != sq.DialectMySQL,
						Index:             wantIndex,
					}
					if m.Dialect == sq.DialectMySQL {
						if !tableExists {
							createTableCmd.CreateIndexCmds = append(createTableCmd.CreateIndexCmds, createIndexCmd)
						} else {
							alterTableCmd.CreateIndexCmds = append(alterTableCmd.CreateIndexCmds, createIndexCmd)
						}
					} else {
						m.CreateIndexCmds = append(m.CreateIndexCmds, createIndexCmd)
					}
				}
				for _, wantTrigger := range wantTable.Triggers {
					if wantTrigger.Ignore {
						continue
					}
					if n := gotTable.CachedTriggerPosition(wantTrigger.TriggerName); n >= 0 {
						continue
					}
					createTriggerCmd := CreateTriggerCommand{
						Trigger: wantTrigger,
					}
					m.CreateTriggerCmds = append(m.CreateTriggerCmds, createTriggerCmd)
				}
				if !tableExists {
					m.CreateTableCmds = append(m.CreateTableCmds, createTableCmd)
				}
				if m.Dialect == sq.DialectSQLite {
					m.AlterTableCmds = append(m.AlterTableCmds, decomposeAlterTableCommandSQLite2(alterTableCmd)...)
				} else if len(alterTableCmd.AddColumnCmds) > 0 ||
					len(alterTableCmd.AlterColumnCmds) > 0 ||
					len(alterTableCmd.AddConstraintCmds) > 0 ||
					len(alterTableCmd.AlterConstraintCmds) > 0 ||
					len(alterTableCmd.CreateIndexCmds) > 0 {
					m.AlterTableCmds = append(m.AlterTableCmds, alterTableCmd)
				}
				if len(addForeignKeyCmd.AddConstraintCmds) > 0 {
					m.AddForeignKeyCmds = append(m.AddForeignKeyCmds, addForeignKeyCmd)
				}
			}
			for _, wantView := range wantSchema.Views {
				if wantView.Ignore {
					continue
				}
				if n := gotSchema.CachedViewPosition(wantView.ViewName); n >= 0 {
					continue
				}
				createViewCmd := CreateViewCommand{
					CreateOrReplace:   m.Dialect == sq.DialectMySQL || (m.Dialect == sq.DialectPostgres && !wantView.IsMaterialized),
					CreateIfNotExists: m.Dialect == sq.DialectSQLite || (m.Dialect == sq.DialectPostgres && wantView.IsMaterialized),
					View:              wantView,
				}
				m.CreateViewCmds = append(m.CreateViewCmds, createViewCmd)
				if createViewCmd.View.IsMaterialized && m.Dialect == sq.DialectPostgres {
					for _, wantIndex := range wantView.Indexes {
						if wantIndex.Ignore {
							continue
						}
						createIndexCmd := CreateIndexCommand{
							CreateIfNotExists: true,
							Index:             wantIndex,
						}
						m.CreateIndexCmds = append(m.CreateIndexCmds, createIndexCmd)
					}
					for _, wantTrigger := range wantView.Triggers {
						if wantTrigger.Ignore {
							continue
						}
						createTriggerCmd := CreateTriggerCommand{
							Trigger: wantTrigger,
						}
						m.CreateTriggerCmds = append(m.CreateTriggerCmds, createTriggerCmd)
					}
				}
			}
			if m.Dialect != sq.DialectSQLite {
				for _, wantFunction := range wantSchema.Functions {
					if wantFunction.Ignore {
						continue
					}
					if positions := gotSchema.CachedFunctionPositions(wantFunction.FunctionName); len(positions) > 0 {
						continue
					}
					createFunctionCmd := CreateFunctionCommand{
						Function: wantFunction,
					}
					m.CreateFunctionCmds = append(m.CreateFunctionCmds, createFunctionCmd)
				}
			}
		}
	}
	if mode&DropExtraneous != 0 {
		dropTableCmd := DropTableCommand{
			DropIfExists: true,
			DropCascade:  mode&DropCascade != 0,
		}
		dropViewCmd := DropViewCommand{
			DropIfExists: true,
			DropCascade:  mode&DropCascade != 0,
		}
		dropMatViewCmd := DropViewCommand{
			DropIfExists:   true,
			IsMaterialized: true,
			DropCascade:    mode&DropCascade != 0,
		}
		dropExtensionCmd := DropExtensionCommand{
			DropIfExists: true,
			DropCascade:  mode&DropCascade != 0,
		}
		var alterTableCmds []AlterTableCommand
		// drop extensions
		for _, gotExtension := range gotDBMetadata.Extensions {
			if strings.HasPrefix(gotExtension, "plpgsql") {
				// we never want to drop the plpgsql extension since postgres
				// enables it by default
				continue
			}
			if n := wantDBMetadata.CachedExtensionPosition(gotExtension); n >= 0 {
				continue
			}
			dropExtensionCmd.Extensions = append(dropExtensionCmd.Extensions, gotExtension)
		}
		if len(dropExtensionCmd.Extensions) > 0 {
			m.DropExtensionCmds = append(m.DropExtensionCmds, dropExtensionCmd)
		}
		for _, gotSchema := range gotDBMetadata.Schemas {
			if gotSchema.Ignore {
				continue
			}
			wantSchema := Schema{
				SchemaName: gotSchema.SchemaName,
			}
			if n := wantDBMetadata.CachedSchemaPosition(gotSchema.SchemaName); n >= 0 {
				wantSchema = wantDBMetadata.Schemas[n]
			}
			// drop tables
			for _, gotTable := range gotSchema.Tables {
				if gotTable.Ignore {
					continue
				}
				wantTable := Table{
					TableSchema: gotTable.TableSchema,
					TableName:   gotTable.TableName,
				}
				if n := wantSchema.CachedTablePosition(gotTable.TableName); n >= 0 {
					wantTable = wantSchema.Tables[n]
				} else {
					dropTableCmd.TableSchemas = append(dropTableCmd.TableSchemas, gotTable.TableSchema)
					dropTableCmd.TableNames = append(dropTableCmd.TableNames, gotTable.TableName)
					continue
				}
				alterTableCmd := AlterTableCommand{
					TableSchema: gotTable.TableSchema,
					TableName:   gotTable.TableName,
				}
				// drop columns
				for _, gotColumn := range gotTable.Columns {
					if n := wantTable.CachedColumnPosition(gotColumn.ColumnName); n >= 0 {
						continue
					}
					dropColumnCmd := DropColumnCommand{
						DropIfExists: m.Dialect == sq.DialectPostgres,
						ColumnName:   gotColumn.ColumnName,
						DropCascade:  m.Dialect == sq.DialectPostgres,
					}
					alterTableCmd.DropColumnCmds = append(alterTableCmd.DropColumnCmds, dropColumnCmd)
				}
				// drop constraints
				if m.Dialect != sq.DialectSQLite {
					for _, gotConstraint := range gotTable.Constraints {
						if n := wantTable.CachedConstraintPosition(gotConstraint.ConstraintName); n >= 0 {
							continue
						}
						dropConstraintCmd := DropConstraintCommand{
							DropIfExists:   m.Dialect == sq.DialectPostgres,
							ConstraintName: gotConstraint.ConstraintName,
							DropCascade:    m.Dialect == sq.DialectPostgres,
						}
						alterTableCmd.DropConstraintCmds = append(alterTableCmd.DropConstraintCmds, dropConstraintCmd)
					}
				}
				// drop indexes
				for _, gotIndex := range gotTable.Indexes {
					if n := wantTable.CachedIndexPosition(gotIndex.IndexName); n >= 0 {
						continue
					}
					dropIndexCmd := DropIndexCommand{
						DropIfExists: m.Dialect == sq.DialectPostgres || m.Dialect == sq.DialectSQLite,
						TableSchema:  gotIndex.TableSchema,
						TableName:    gotIndex.TableName,
						IndexName:    gotIndex.IndexName,
						DropCascade:  m.Dialect == sq.DialectPostgres,
					}
					if m.Dialect == sq.DialectMySQL {
						alterTableCmd.DropIndexCmds = append(alterTableCmd.DropIndexCmds, dropIndexCmd)
					} else {
						m.DropIndexCmds = append(m.DropIndexCmds, dropIndexCmd)
					}
				}
				// drop triggers
				for _, gotTrigger := range gotTable.Triggers {
					if n := wantTable.CachedTriggerPosition(gotTrigger.TriggerName); n >= 0 {
						continue
					}
					dropTriggerCmd := DropTriggerCommand{
						DropIfExists: true,
						TableSchema:  gotTrigger.TableSchema,
						TableName:    gotTrigger.TableName,
						TriggerName:  gotTrigger.TriggerName,
						DropCascade:  m.Dialect == sq.DialectPostgres,
					}
					m.DropTriggerCmds = append(m.DropTriggerCmds, dropTriggerCmd)
				}
				if len(alterTableCmd.DropColumnCmds) > 0 ||
					len(alterTableCmd.DropConstraintCmds) > 0 ||
					len(alterTableCmd.DropIndexCmds) > 0 {
					alterTableCmds = append(alterTableCmds, alterTableCmd)
				}
			}
			// drop views
			for _, gotView := range gotSchema.Views {
				if n := wantSchema.CachedViewPosition(gotView.ViewName); n >= 0 {
					continue
				}
				if gotView.IsMaterialized {
					dropMatViewCmd.ViewSchemas = append(dropMatViewCmd.ViewSchemas, gotView.ViewSchema)
					dropMatViewCmd.ViewNames = append(dropMatViewCmd.ViewNames, gotView.ViewName)
				} else {
					dropViewCmd.ViewSchemas = append(dropViewCmd.ViewSchemas, gotView.ViewSchema)
					dropViewCmd.ViewNames = append(dropViewCmd.ViewNames, gotView.ViewName)
				}
			}
			// drop functions
			for _, gotFunction := range gotSchema.Functions {
				if positions := wantSchema.CachedFunctionPositions(gotFunction.FunctionName); len(positions) > 0 {
					continue
				}
				dropFunctionCmd := DropFunctionCommand{
					DropIfExists: true,
					Function:     gotFunction,
					DropCascade:  true,
				}
				m.DropFunctionCmds = append(m.DropFunctionCmds, dropFunctionCmd)
			}
		}
		if m.Dialect == sq.DialectSQLite {
			m.DropViewCmds = append(m.DropViewCmds, decomposeDropViewCommandSQLite2(dropViewCmd)...)
			m.DropTableCmds = append(m.DropTableCmds, decomposeDropTableCommandSQLite2(dropTableCmd)...)
			for _, alterTableCmd := range alterTableCmds {
				m.AlterTableDropCmds = append(m.AlterTableDropCmds, decomposeAlterTableCommandSQLite2(alterTableCmd)...)
			}
		} else {
			for _, alterTableCmd := range alterTableCmds {
				m.AlterTableDropCmds = append(m.AlterTableDropCmds, alterTableCmd)
			}
			if len(dropViewCmd.ViewNames) > 0 {
				m.DropViewCmds = append(m.DropViewCmds, dropViewCmd)
			}
			if len(dropMatViewCmd.ViewNames) > 0 {
				m.DropViewCmds = append(m.DropViewCmds, dropMatViewCmd)
			}
			if len(dropTableCmd.TableNames) > 0 {
				m.DropTableCmds = append(m.DropTableCmds, dropTableCmd)
			}
		}
	}
	return m, nil
}

func diffColumn(dialect string, gotColumn, wantColumn Column) (alterColumnCmd AlterColumnCommand, isDifferent bool) {
	// do we SET DATA TYPE?
	if !strings.EqualFold(gotColumn.ColumnType, wantColumn.ColumnType) {
		isDifferent = true
		alterColumnCmd.Column.ColumnType = wantColumn.ColumnType
	}
	// do we DROP NOT NULL?
	if gotColumn.IsNotNull && !wantColumn.IsNotNull {
		isDifferent = true
		alterColumnCmd.DropNotNull = true
	} else
	// do we SET NOT NULL?
	if !gotColumn.IsNotNull && wantColumn.IsNotNull {
		isDifferent = true
		alterColumnCmd.Column.IsNotNull = true
	}
	// do we DROP DEFAULT?
	if gotColumn.ColumnDefault != "" && wantColumn.ColumnDefault == "" {
		isDifferent = true
		alterColumnCmd.DropDefault = true
	} else
	// do we SET DEFAULT?
	if gotColumn.ColumnDefault == "" && wantColumn.ColumnDefault != "" {
		isDifferent = true
		alterColumnCmd.Column.ColumnDefault = wantColumn.ColumnDefault
	}
	if dialect == sq.DialectPostgres {
		// do we DROP IDENTITY?
		if gotColumn.Identity != "" && wantColumn.Identity == "" {
			isDifferent = true
			alterColumnCmd.DropIdentity = true
			alterColumnCmd.DropIdentityIfExists = true
		} else
		// do we ADD GENERATED AS $IDENTITY?
		if gotColumn.Identity == "" && wantColumn.Identity != "" {
			isDifferent = true
			alterColumnCmd.Column.Identity = wantColumn.Identity
		}
	}
	return alterColumnCmd, isDifferent
}

func (m *Migration) WriteSQL(w io.Writer) error {
	var err error
	var written bool
	writeCmd := func(cmd Command, isMySQLFunction bool) error {
		query, args, _, err := sq.ToSQL(m.Dialect, cmd)
		if err != nil {
			return fmt.Errorf("building command (%s): %w", query, err)
		}
		if len(args) > 0 {
			query, err = sq.Sprintf(m.Dialect, query, args)
			if err != nil {
				return fmt.Errorf("building command (%s): %w", query, err)
			}
		}
		if !written {
			written = true
		} else {
			io.WriteString(w, "\n\n")
		}
		query = strings.TrimSpace(query)
		io.WriteString(w, query)
		if isMySQLFunction {
			io.WriteString(w, "; -- ;;")
		} else {
			if last := len(query) - 1; query[last] != ';' {
				io.WriteString(w, ";")
			}
		}
		return nil
	}
	for _, cmd := range m.CreateSchemaCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.CreateExtensionCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	if len(m.CreateFunctionCmds) > 0 {
		if m.Dialect == sq.DialectMySQL {
			io.WriteString(w, "\n\n-- DELIMITER ;;")
		}
		for _, cmd := range m.CreateFunctionCmds {
			if cmd.Ignore {
				continue
			}
			err = writeCmd(cmd, m.Dialect == sq.DialectMySQL)
			if err != nil {
				return err
			}
		}
		if m.Dialect == sq.DialectMySQL {
			io.WriteString(w, "\n\n-- DELIMITER ;")
		}
	}
	for _, cmd := range m.CreateTableCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.AlterTableCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.CreateViewCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.CreateIndexCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	if len(m.CreateTriggerCmds) > 0 {
		if m.Dialect == sq.DialectMySQL {
			io.WriteString(w, "\n\n-- DELIMITER ;;")
		}
		for _, cmd := range m.CreateTriggerCmds {
			if cmd.Ignore {
				continue
			}
			err = writeCmd(cmd, m.Dialect == sq.DialectMySQL)
			if err != nil {
				return err
			}
		}
		if m.Dialect == sq.DialectMySQL {
			io.WriteString(w, "\n\n-- DELIMITER ;")
		}
	}
	for _, cmd := range m.AddForeignKeyCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.DropViewCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.DropTableCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.DropTriggerCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.DropIndexCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.AlterTableDropCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.DropFunctionCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.DropExtensionCmds {
		if cmd.Ignore {
			continue
		}
		err = writeCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Migration) Exec(db sq.DB) error {
	return m.ExecContext(context.Background(), db)
}

func (m *Migration) ExecContext(ctx context.Context, db sq.DB) error {
	var err error
	execCmd := func(cmd Command) error {
		query, args, _, err := sq.ToSQL(m.Dialect, cmd)
		if err != nil {
			return fmt.Errorf("building command (%s): %w", query, err)
		}
		_, err = db.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("executing command (%s): %w", query, err)
		}
		return nil
	}
	for _, cmd := range m.CreateSchemaCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.CreateExtensionCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.CreateFunctionCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.CreateTableCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.AlterTableCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.CreateViewCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.CreateIndexCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.CreateTriggerCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.AddForeignKeyCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.DropViewCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.DropTableCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.DropTriggerCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.DropIndexCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.AlterTableDropCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.DropFunctionCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	for _, cmd := range m.DropExtensionCmds {
		if cmd.Ignore {
			continue
		}
		err = execCmd(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}
