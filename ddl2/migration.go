package ddl2

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/bokwoon95/sq"
)

type MigrationMode int

const (
	CreateMissing  MigrationMode = 0b1
	UpdateExisting MigrationMode = 0b10
	DropExtraneous MigrationMode = 0b100
	DropCascade    MigrationMode = 0b1000
)

type Migration struct {
	Dialect                     string
	CurrentSchema               string
	SchemaCommands              []Command
	ExtensionCommands           []Command
	EnumCommands                []Command
	IndependentFunctionCommands []Command
	TableCommands               []Command
	ViewCommands                []Command
	IndexCommands               []Command
	FunctionCommands            []Command
	TriggerCommands             []Command
	ForeignKeyCommands          []Command
	RenameCommands              []Command
	DropCommands                []Command
}

func AutoMigrate(dialect string, db sq.DB, migrationMode MigrationMode, wantCatalogOpts ...CatalogOption) error {
	gotCatalog, err := NewCatalog(dialect, WithDB(db))
	if err != nil {
		return fmt.Errorf("introspecting db: %w", err)
	}
	wantCatalog, err := NewCatalog(dialect, wantCatalogOpts...)
	if err != nil {
		return fmt.Errorf("building catalog: %w", err)
	}
	migration, err := Migrate(migrationMode, gotCatalog, wantCatalog)
	if err != nil {
		return fmt.Errorf("building migration: %w", err)
	}
	err = migration.Exec(db)
	if err != nil {
		return fmt.Errorf("executing migration: %w", err)
	}
	return nil
}

func Migrate(mode MigrationMode, gotCatalog, wantCatalog Catalog) (Migration, error) {
	m := Migration{
		Dialect:       gotCatalog.Dialect,
		CurrentSchema: gotCatalog.CurrentSchema,
	}
	if gotCatalog.Dialect == "" && wantCatalog.Dialect == "" {
		return m, fmt.Errorf("dialect missing")
	}
	if gotCatalog.Dialect != "" && wantCatalog.Dialect != "" && gotCatalog.Dialect != wantCatalog.Dialect {
		return m, fmt.Errorf("dialect mismatch: gotCatalog is %s, wantCatalog is %s", gotCatalog.Dialect, wantCatalog.Dialect)
	}
	if m.Dialect == "" {
		m.Dialect = wantCatalog.Dialect
	}
	var err error
	if mode&CreateMissing != 0 {
		if m.Dialect == sq.DialectPostgres {
			for _, wantExtension := range wantCatalog.Extensions {
				if n := gotCatalog.CachedExtensionPosition(wantExtension); n >= 0 {
					continue
				}
				createExtensionCmd := &CreateExtensionCommand{
					CreateIfNotExists: true,
					Extension:         wantExtension,
				}
				m.ExtensionCommands = append(m.ExtensionCommands, createExtensionCmd)
			}
		}
		for _, wantSchema := range wantCatalog.Schemas {
			if wantSchema.Ignore {
				continue
			}
			err = migrateSchema(&m, mode, gotCatalog, wantSchema)
			if err != nil {
				return m, err
			}
		}
	}
	if mode&DropExtraneous != 0 {
		err = dropExtraneousObjects(&m, mode, gotCatalog, wantCatalog)
		if err != nil {
			return m, err
		}
	}
	return m, nil
}

func migrateSchema(m *Migration, mode MigrationMode, gotCatalog Catalog, wantSchema Schema) error {
	if mode&CreateMissing == 0 {
		return nil
	}
	var err error
	var gotSchema Schema
	if wantSchema.SchemaName == "" {
		wantSchema.SchemaName = m.CurrentSchema
	}
	if n := gotCatalog.CachedSchemaPosition(wantSchema.SchemaName); n >= 0 {
		gotSchema = gotCatalog.Schemas[n]
	} else {
		gotSchema.SchemaName = wantSchema.SchemaName
		if gotSchema.SchemaName != "" {
			createSchemaCmd := &CreateSchemaCommand{
				CreateIfNotExists: true,
				SchemaName:        wantSchema.SchemaName,
			}
			m.SchemaCommands = append(m.SchemaCommands, createSchemaCmd)
		}
	}
	for _, wantTable := range wantSchema.Tables {
		if wantTable.Ignore {
			continue
		}
		err = migrateTable(m, mode, gotSchema, wantTable)
		if err != nil {
			return err
		}
	}
	for _, wantView := range wantSchema.Views {
		if wantView.Ignore {
			continue
		}
		if n := gotSchema.CachedViewPosition(wantView.ViewName); n >= 0 {
			continue
		}
		createViewCmd := &CreateViewCommand{View: wantView}
		if m.Dialect == sq.DialectMySQL || (m.Dialect == sq.DialectPostgres && !wantView.IsMaterialized) {
			createViewCmd.CreateOrReplace = true
		}
		if m.Dialect == sq.DialectSQLite || (m.Dialect == sq.DialectPostgres && wantView.IsMaterialized) {
			createViewCmd.CreateIfNotExists = true
		}
		m.ViewCommands = append(m.ViewCommands, createViewCmd)
		if wantView.IsMaterialized && m.Dialect == sq.DialectPostgres {
			for _, wantIndex := range wantView.Indexes {
				if wantIndex.Ignore {
					continue
				}
				createIndexCmd := &CreateIndexCommand{CreateIfNotExists: true, Index: wantIndex}
				m.IndexCommands = append(m.IndexCommands, createIndexCmd)
			}
			for _, wantTrigger := range wantView.Triggers {
				if wantTrigger.Ignore {
					continue
				}
				createTriggerCmd := &CreateTriggerCommand{Trigger: wantTrigger}
				m.TriggerCommands = append(m.TriggerCommands, createTriggerCmd)
			}
		}
	}
	if m.Dialect == sq.DialectPostgres || m.Dialect == sq.DialectMySQL {
		for _, wantFunction := range wantSchema.Functions {
			if wantFunction.Ignore {
				continue
			}
			if positions := gotSchema.CachedFunctionPositions(wantFunction.FunctionName); len(positions) > 0 {
				continue
			}
			createFunctionCmd := &CreateFunctionCommand{Function: wantFunction}
			if wantFunction.IsIndependent {
				m.IndependentFunctionCommands = append(m.IndependentFunctionCommands, createFunctionCmd)
			} else {
				m.FunctionCommands = append(m.FunctionCommands, createFunctionCmd)
			}
		}
	}
	return nil
}

func migrateTable(m *Migration, mode MigrationMode, gotSchema Schema, wantTable Table) error {
	var gotTable Table
	var createTableCmd *CreateTableCommand
	var alterTableCmd *AlterTableCommand
	var fkeyCmd *AlterTableCommand
	if n := gotSchema.CachedTablePosition(wantTable.TableName); n >= 0 {
		gotTable = gotSchema.Tables[n]
		for _, wantColumn := range wantTable.Columns {
			if wantColumn.Ignore {
				continue
			}
			addColumnCmd, alterColumnCmd := migrateColumn(m.Dialect, mode, gotTable, wantColumn)
			if addColumnCmd != nil || alterColumnCmd != nil {
				if alterTableCmd == nil {
					alterTableCmd = &AlterTableCommand{TableSchema: wantTable.TableSchema, TableName: wantTable.TableName}
				}
				if addColumnCmd != nil {
					alterTableCmd.AddColumnCommands = append(alterTableCmd.AddColumnCommands, *addColumnCmd)
				}
				if alterColumnCmd != nil {
					alterTableCmd.AlterColumnCommands = append(alterTableCmd.AlterColumnCommands, *alterColumnCmd)
				}
			}
		}
	} else {
		gotTable.TableSchema = wantTable.TableSchema
		gotTable.TableName = wantTable.TableName
		createTableCmd = &CreateTableCommand{
			CreateIfNotExists:  true,
			IncludeConstraints: true,
			Table:              wantTable,
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
			addConstraintCmd := AddConstraintCommand{Constraint: wantConstraint}
			switch wantConstraint.ConstraintType {
			case FOREIGN_KEY:
				if fkeyCmd == nil {
					fkeyCmd = &AlterTableCommand{TableSchema: wantTable.TableSchema, TableName: wantTable.TableName}
				}
				fkeyCmd.AddConstraintCommands = append(fkeyCmd.AddConstraintCommands, addConstraintCmd)
			default:
				if createTableCmd == nil || createTableCmd.IncludeConstraints {
					continue
				}
				if alterTableCmd == nil {
					alterTableCmd = &AlterTableCommand{TableSchema: wantTable.TableSchema, TableName: wantTable.TableName}
				}
				alterTableCmd.AddConstraintCommands = append(alterTableCmd.AddConstraintCommands, addConstraintCmd)
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
		createIndexCmd := CreateIndexCommand{Index: wantIndex}
		switch m.Dialect {
		case sq.DialectMySQL:
			if createTableCmd != nil {
				createTableCmd.CreateIndexCommands = append(createTableCmd.CreateIndexCommands, createIndexCmd)
			} else {
				if alterTableCmd == nil {
					alterTableCmd = &AlterTableCommand{TableSchema: wantTable.TableSchema, TableName: wantTable.TableName}
				}
				alterTableCmd.CreateIndexCommands = append(alterTableCmd.CreateIndexCommands, createIndexCmd)
			}
		default:
			createIndexCmd.CreateIfNotExists = true
			m.IndexCommands = append(m.IndexCommands, &createIndexCmd)
		}
	}
	for _, wantTrigger := range wantTable.Triggers {
		if wantTrigger.Ignore {
			continue
		}
		if n := gotTable.CachedTriggerPosition(wantTrigger.TriggerName); n >= 0 {
			continue
		}
		createTriggerCmd := CreateTriggerCommand{Trigger: wantTrigger}
		m.TriggerCommands = append(m.TriggerCommands, createTriggerCmd)
	}
	if createTableCmd != nil {
		m.TableCommands = append(m.TableCommands, createTableCmd)
	}
	if alterTableCmd != nil {
		switch m.Dialect {
		case sq.DialectSQLite:
			alterTableCmds := decomposeAlterTableCommandSQLite(alterTableCmd)
			m.TableCommands = append(m.TableCommands, alterTableCmds...)
		case sq.DialectPostgres:
			alterTableCmd.AlterIfExists = true
			m.TableCommands = append(m.TableCommands, alterTableCmd)
		case sq.DialectMySQL:
			m.TableCommands = append(m.TableCommands, alterTableCmd)
		}
	}
	if fkeyCmd != nil {
		switch m.Dialect {
		case sq.DialectPostgres:
			fkeyCmd.AlterIfExists = true
			m.ForeignKeyCommands = append(m.ForeignKeyCommands, fkeyCmd)
		case sq.DialectMySQL:
			m.ForeignKeyCommands = append(m.ForeignKeyCommands, fkeyCmd)
		}
	}
	return nil
}

func migrateColumn(dialect string, mode MigrationMode, gotTable Table, wantColumn Column) (*AddColumnCommand, *AlterColumnCommand) {
	n := gotTable.CachedColumnPosition(wantColumn.ColumnName)
	if n < 0 {
		addColumnCmd := &AddColumnCommand{Column: wantColumn}
		if dialect == sq.DialectPostgres {
			addColumnCmd.AddIfNotExists = true
		}
		return addColumnCmd, nil
	}
	if dialect == sq.DialectSQLite || mode&UpdateExisting == 0 {
		return nil, nil
	}
	var columnModified bool
	gotColumn := gotTable.Columns[n]
	alterColumnCmd := &AlterColumnCommand{Column: Column{
		TableSchema: wantColumn.TableSchema,
		TableName:   wantColumn.TableName,
		ColumnName:  wantColumn.ColumnName,
	}}
	if !datatypeEq(dialect, gotColumn.ColumnType, wantColumn.ColumnType) {
		columnModified = true
		alterColumnCmd.Column.ColumnType = wantColumn.ColumnType
	}
	if gotColumn.IsNotNull && !wantColumn.IsNotNull {
		columnModified = true
		alterColumnCmd.DropNotNull = true
	} else if !gotColumn.IsNotNull && wantColumn.IsNotNull {
		columnModified = true
		alterColumnCmd.Column.IsNotNull = true
	}
	if gotColumn.ColumnDefault != "" && wantColumn.ColumnDefault == "" {
		columnModified = true
		alterColumnCmd.DropDefault = true
	} else if gotColumn.ColumnDefault == "" && wantColumn.ColumnDefault != "" {
		columnModified = true
		alterColumnCmd.Column.ColumnDefault = wantColumn.ColumnDefault
	}
	if dialect == sq.DialectPostgres {
		if gotColumn.Identity != "" && wantColumn.Identity == "" {
			columnModified = true
			alterColumnCmd.DropIdentity = true
			alterColumnCmd.DropIdentityIfExists = true
		} else if gotColumn.Identity == "" && wantColumn.Identity != "" {
			columnModified = true
			alterColumnCmd.Column.Identity = wantColumn.Identity
		}
	}
	if columnModified {
		return nil, alterColumnCmd
	}
	return nil, nil
}

// TODO: implement this function
func datatypeEq(dialect, typeA, typeB string) bool {
	return strings.EqualFold(typeA, typeB)
}

func dropExtraneousObjects(m *Migration, mode MigrationMode, gotCatalog, wantCatalog Catalog) error {
	dropTableCmd := DropTableCommand{
		DropIfExists: true,
		DropCascade:  mode&DropCascade != 0,
	}
	dropViewCmd := DropViewCommand{
		DropIfExists: true,
		DropCascade:  mode&DropCascade != 0,
	}
	dropMatViewCmd := DropViewCommand{
		IsMaterialized: true,
		DropCascade:    mode&DropCascade != 0,
	}
	var alterTableCmds []AlterTableCommand // drop columns, drop constraints, drop indexes (mysql only)
	var dropIndexCmds []Command
	var dropTriggerCmds []Command
	for _, gotSchema := range gotCatalog.Schemas {
		wantSchema := Schema{SchemaName: gotSchema.SchemaName}
		if n := wantCatalog.CachedSchemaPosition(gotSchema.SchemaName); n >= 0 {
			wantSchema = wantCatalog.Schemas[n]
		}
		// drop tables
		for _, gotTable := range gotSchema.Tables {
			wantTable := Table{TableSchema: gotTable.TableSchema, TableName: gotTable.TableName}
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
				if n := wantTable.CachedColumnPosition(gotColumn.ColumnName); n < 0 {
					dropColumnCmd := DropColumnCommand{ColumnName: gotColumn.ColumnName}
					if m.Dialect == sq.DialectPostgres {
						dropColumnCmd.DropIfExists = true
						dropColumnCmd.DropCascade = true
					}
					alterTableCmd.DropColumnCommands = append(alterTableCmd.DropColumnCommands, dropColumnCmd)
				}
			}
			// drop constraints
			if m.Dialect != sq.DialectSQLite {
				for _, gotConstraint := range gotTable.Constraints {
					if n := wantTable.CachedConstraintPosition(gotConstraint.ConstraintName); n < 0 {
						dropConstraintCmd := DropConstraintCommand{ConstraintName: gotConstraint.ConstraintName}
						if m.Dialect == sq.DialectPostgres {
							dropConstraintCmd.DropIfExists = true
							dropConstraintCmd.DropCascade = true
						}
						alterTableCmd.DropConstraintCommands = append(alterTableCmd.DropConstraintCommands, dropConstraintCmd)
					}
				}
			}
			// drop indexes
			for _, gotIndex := range gotTable.Indexes {
				if n := wantTable.CachedIndexPosition(gotIndex.IndexName); n < 0 {
					dropIndexCmd := DropIndexCommand{
						TableSchema: gotIndex.TableSchema,
						TableName:   gotIndex.TableName,
						IndexName:   gotIndex.IndexName,
					}
					switch m.Dialect {
					case sq.DialectSQLite:
						dropIndexCmd.DropIfExists = true
						dropIndexCmd.DropCascade = true
						dropIndexCmds = append(dropIndexCmds, &dropIndexCmd)
					case sq.DialectPostgres:
						dropIndexCmd.DropIfExists = true
						dropIndexCmds = append(dropIndexCmds, &dropIndexCmd)
					case sq.DialectMySQL:
						alterTableCmd.DropIndexCommands = append(alterTableCmd.DropIndexCommands, dropIndexCmd)
					}
				}
			}
			// drop triggers
			for _, gotTrigger := range gotTable.Triggers {
				if n := wantTable.CachedTriggerPosition(gotTrigger.TriggerName); n < 0 {
					dropTriggerCmd := DropTriggerCommand{
						DropIfExists: true,
						TableSchema:  gotTrigger.TableSchema,
						TableName:    gotTrigger.TableName,
						TriggerName:  gotTrigger.TriggerName,
					}
					if m.Dialect == sq.DialectPostgres {
						dropTriggerCmd.DropCascade = true
					}
					dropTriggerCmds = append(dropTriggerCmds, &dropTriggerCmd)
				}
			}
			if len(alterTableCmd.DropColumnCommands) > 0 || len(alterTableCmd.DropConstraintCommands) > 0 || len(alterTableCmd.DropIndexCommands) > 0 {
				alterTableCmds = append(alterTableCmds, alterTableCmd)
			}
		}
		// drop views
		for _, gotView := range gotSchema.Views {
			if n := wantSchema.CachedViewPosition(gotView.ViewName); n < 0 {
				if gotView.IsMaterialized {
					dropMatViewCmd.ViewSchemas = append(dropMatViewCmd.ViewSchemas, gotView.ViewSchema)
					dropMatViewCmd.ViewNames = append(dropMatViewCmd.ViewNames, gotView.ViewName)
				} else {
					dropViewCmd.ViewSchemas = append(dropViewCmd.ViewSchemas, gotView.ViewSchema)
					dropViewCmd.ViewNames = append(dropViewCmd.ViewNames, gotView.ViewName)
				}
			}
		}
	}
	m.DropCommands = append(m.DropCommands, dropTriggerCmds...)
	m.DropCommands = append(m.DropCommands, dropIndexCmds...)
	switch m.Dialect {
	case sq.DialectSQLite:
		for _, alterTableCmd := range alterTableCmds {
			m.DropCommands = append(m.DropCommands, decomposeAlterTableCommandSQLite(&alterTableCmd)...)
		}
		m.DropCommands = append(m.DropCommands, decomposeDropViewCommandSQLite(&dropViewCmd)...)
		m.DropCommands = append(m.DropCommands, decomposeDropTableCommandSQLite(&dropTableCmd)...)
	default:
		for _, alterTableCmd := range alterTableCmds {
			m.DropCommands = append(m.DropCommands, &alterTableCmd)
		}
		if len(dropViewCmd.ViewNames) > 0 {
			m.DropCommands = append(m.DropCommands, &dropViewCmd)
		}
		if len(dropMatViewCmd.ViewNames) > 0 {
			m.DropCommands = append(m.DropCommands, &dropMatViewCmd)
		}
		if len(dropTableCmd.TableNames) > 0 {
			m.DropCommands = append(m.DropCommands, &dropTableCmd)
		}
	}
	return nil
}

// NOTE: this function is too simple to warrant both a DropFunctions and
// DropFunctionsContext slot. Ask the user to do it themselves. Deprecate this
// function.
func DropFunctions(dialect string, db sq.DB, functions []Function, dropCascade bool) error {
	var cmd DropFunctionCommand
	for _, function := range functions {
		cmd.DropIfExists = true
		cmd.Function = function
		cmd.DropCascade = dropCascade
		query, args, _, err := sq.ToSQL(dialect, cmd)
		if err != nil {
			return fmt.Errorf("building command (%s): %w", query, err)
		}
		_, err = db.ExecContext(context.Background(), query, args...)
		if err != nil {
			return fmt.Errorf("executing command (%s): %w", query, err)
		}
	}
	return nil
}

func (m *Migration) WriteSQL(w io.Writer) error {
	var written bool
	for _, cmds := range [][]Command{
		m.SchemaCommands,
		m.IndependentFunctionCommands,
		m.TableCommands,
		m.ViewCommands,
		m.IndexCommands,
		m.FunctionCommands,
		m.TriggerCommands,
		m.ForeignKeyCommands,
		m.RenameCommands,
		m.DropCommands,
	} {
		for _, cmd := range cmds {
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
			if last := len(query) - 1; query[last] != ';' {
				io.WriteString(w, ";")
			}
		}
	}
	return nil
}

func (m *Migration) Exec(db sq.DB) error {
	return m.ExecContext(context.Background(), db)
}

func (m *Migration) ExecContext(ctx context.Context, db sq.DB) error {
	for _, cmds := range [][]Command{
		m.SchemaCommands,
		m.ExtensionCommands,
		m.EnumCommands,
		m.IndependentFunctionCommands,
		m.TableCommands,
		m.ViewCommands,
		m.IndexCommands,
		m.TriggerCommands,
		m.ForeignKeyCommands,
		m.RenameCommands,
		m.DropCommands,
	} {
		for i, cmd := range cmds {
			_ = i
			query, args, _, err := sq.ToSQL(m.Dialect, cmd)
			if err != nil {
				return fmt.Errorf("building command (%s): %w", query, err)
			}
			_, err = db.ExecContext(ctx, query, args...)
			if err != nil {
				return fmt.Errorf("executing command (%s): %w", query, err)
			}
		}
	}
	return nil
}
