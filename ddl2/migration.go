package ddl2

import (
	"context"
	"fmt"
	"io"

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

func AutoMigrate(dialect string, db sq.DB, migrationMode MigrationMode, opts ...CatalogOption) error {
	gotCatalog, err := NewCatalog(dialect, WithDB(db))
	if err != nil {
		return fmt.Errorf("introspecting db: %w", err)
	}
	wantCatalog, err := NewCatalog(dialect, opts...)
	if err != nil {
		return fmt.Errorf("building catalog: %w", err)
	}
	m, err := Migrate(migrationMode, gotCatalog, wantCatalog)
	if err != nil {
		return fmt.Errorf("building migration: %w", err)
	}
	err = m.Exec(db)
	if err != nil {
		return fmt.Errorf("executing migration: %w", err)
	}
	return nil
}

func Migrate(mode MigrationMode, gotCatalog, wantCatalog Catalog) (Migration, error) {
	m := Migration{Dialect: gotCatalog.Dialect}
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
		if len(wantCatalog.Extensions) > 0 {
			if m.Dialect != sq.DialectPostgres {
				return m, fmt.Errorf("{%w} dialect=%s feature=extensions", ErrUnsupportedFeature, m.Dialect)
			}
			for _, wantExtension := range wantCatalog.Extensions {
				if n := gotCatalog.CachedExtensionPosition(wantExtension); n < 0 {
					createExtensionCmd := &CreateExtensionCommand{
						CreateIfNotExists: true,
						Extension:         wantExtension,
					}
					m.ExtensionCommands = append(m.ExtensionCommands, createExtensionCmd)
				}
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
	if n := gotCatalog.CachedSchemaPosition(wantSchema.SchemaName); n >= 0 {
		gotSchema = gotCatalog.Schemas[n]
	} else {
		gotSchema.SchemaName = wantSchema.SchemaName
		createSchemaCmd := &CreateSchemaCommand{
			CreateIfNotExists: true,
			SchemaName:        wantSchema.SchemaName,
		}
		m.SchemaCommands = append(m.SchemaCommands, createSchemaCmd)
	}
	for _, wantTable := range wantSchema.Tables {
		err = migrateTable(m, mode, gotSchema, wantTable)
		if err != nil {
			return err
		}
	}
	for _, wantView := range wantSchema.Views {
		err = migrateView(m, mode, gotSchema, wantView)
		if err != nil {
			return err
		}
	}
	if len(wantSchema.Functions) > 0 {
		if m.Dialect != sq.DialectPostgres && m.Dialect != sq.DialectMySQL {
			return fmt.Errorf("{%w} dialect=%s feature=functions", ErrUnsupportedFeature, m.Dialect)
		}
		for _, wantFunction := range wantSchema.Functions {
			err = migrateFunction(m, mode, gotSchema, wantFunction)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func migrateTable(m *Migration, mode MigrationMode, gotSchema Schema, wantTable Table) error {
	var err error
	var gotTable Table
	var createTableCmd *CreateTableCommand
	var alterTableCmd *AlterTableCommand
	var fkeyCmd *AlterTableCommand
	if n := gotSchema.CachedTablePosition(wantTable.TableName); n >= 0 {
		gotTable = gotSchema.Tables[n]
		alterTableCmd = &AlterTableCommand{
			TableSchema: wantTable.TableSchema,
			TableName:   wantTable.TableName,
		}
		for _, wantColumn := range wantTable.Columns {
			err = migrateColumn(m.Dialect, alterTableCmd, mode, gotTable, wantColumn)
			if err != nil {
				return err
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
	if len(wantTable.Constraints) > 0 && createTableCmd == nil {
		for _, wantConstraint := range wantTable.Constraints {
			switch wantConstraint.ConstraintType {
			case FOREIGN_KEY:
				if fkeyCmd == nil {
					fkeyCmd = &AlterTableCommand{TableSchema: wantTable.TableSchema, TableName: wantTable.TableName}
				}
				err = migrateConstraint(m.Dialect, fkeyCmd, mode, gotTable, wantConstraint)
				if err != nil {
					return err
				}
			default:
				err = migrateConstraint(m.Dialect, alterTableCmd, mode, gotTable, wantConstraint)
				if err != nil {
					return err
				}
			}
		}
	}
	for _, wantIndex := range wantTable.Indexes {
		err = migrateTableIndex(m.Dialect, alterTableCmd, mode, gotTable, wantIndex)
		if err != nil {
			return err
		}
	}
	for _, wantTrigger := range wantTable.Triggers {
		err = migrateTableTrigger(m, mode, gotTable, wantTrigger)
		if err != nil {
			return err
		}
	}
	if m.Dialect == sq.DialectPostgres {
		if alterTableCmd != nil {
			alterTableCmd.AlterIfExists = true
		}
		if fkeyCmd != nil {
			fkeyCmd.AlterIfExists = true
		}
	}
	if createTableCmd != nil {
		m.TableCommands = append(m.TableCommands, createTableCmd)
	}
	if alterTableCmd != nil {
		switch m.Dialect {
		case sq.DialectSQLite:
			alterTableCmds, err := decomposeAlterTableCommandSQLite(alterTableCmd)
			if err != nil {
				return err
			}
			_ = alterTableCmds
		case sq.DialectPostgres:
		case sq.DialectMySQL:
		}
	}
	if fkeyCmd != nil {
	}
	return nil
}

func migrateView(m *Migration, mode MigrationMode, gotSchema Schema, wantView View) error {
	var err error
	var gotView View
	if n := gotSchema.CachedViewPosition(wantView.ViewName); n < 0 {
		gotView.ViewSchema = wantView.ViewSchema
		gotView.ViewName = wantView.ViewName
		createViewCmd := &CreateViewCommand{View: wantView}
		if m.Dialect == sq.DialectSQLite || (m.Dialect == sq.DialectPostgres && wantView.IsMaterialized) {
			createViewCmd.CreateOrReplace = true
		}
		if m.Dialect == sq.DialectMySQL || (m.Dialect == sq.DialectPostgres && !wantView.IsMaterialized) {
			createViewCmd.CreateIfNotExists = true
		}
		m.ViewCommands = append(m.ViewCommands, createViewCmd)
		if wantView.IsMaterialized {
			if m.Dialect != sq.DialectPostgres {
				return fmt.Errorf("{%w} dialect=%s feature={materialized views}", ErrUnsupportedFeature, m.Dialect)
			}
			for _, wantIndex := range wantView.Indexes {
				err = migrateViewIndex(m, mode, gotView, wantIndex)
				if err != nil {
					return err
				}
			}
			for _, wantTrigger := range wantView.Triggers {
				err = migrateViewTrigger(m, mode, gotView, wantTrigger)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func migrateFunction(m *Migration, mode MigrationMode, gotSchema Schema, wantFunction Function) error {
	if positions := gotSchema.CachedFunctionPositions(wantFunction.FunctionName); len(positions) == 0 {
		createFunctionCmd := &CreateFunctionCommand{Function: wantFunction}
		if wantFunction.IsIndependent {
			m.IndependentFunctionCommands = append(m.IndependentFunctionCommands, createFunctionCmd)
		} else {
			m.FunctionCommands = append(m.FunctionCommands, createFunctionCmd)
		}
	}
	return nil
}

func migrateColumn(dialect string, alterTableCmd *AlterTableCommand, mode MigrationMode, gotTable Table, wantColumn Column) error {
	return nil
}

func migrateConstraint(dialect string, alterTableCmd *AlterTableCommand, mode MigrationMode, gotTable Table, wantConstraint Constraint) error {
	if n := gotTable.CachedConstraintPosition(wantConstraint.ConstraintName); n < 0 {
	}
	return nil
}

func migrateTableIndex(dialect string, alterTableCmd *AlterTableCommand, mode MigrationMode, gotTable Table, wantIndex Index) error {
	if n := gotTable.CachedIndexPosition(wantIndex.IndexName); n < 0 {
	}
	return nil
}

func migrateTableTrigger(m *Migration, mode MigrationMode, gotTable Table, wantTrigger Trigger) error {
	return nil
}

func migrateViewIndex(m *Migration, mode MigrationMode, gotView View, wantIndex Index) error {
	return nil
}

func migrateViewTrigger(m *Migration, mode MigrationMode, gotView View, wantTrigger Trigger) error {
	return nil
}

func dropExtraneousObjects(m *Migration, mode MigrationMode, gotCatalog, wantCatalog Catalog) error {
	dropTableCmd := DropTableCommand{
		DropIfExists: true,
		DropCascade:  mode&DropCascade != 0,
	}
	for _, gotSchema := range gotCatalog.Schemas {
		n1 := wantCatalog.CachedSchemaPosition(gotSchema.SchemaName)
		if n1 < 0 {
			break
		}
		wantSchema := wantCatalog.Schemas[n1]
		// drop tables
		for _, gotTable := range gotSchema.Tables {
			n2 := wantSchema.CachedTablePosition(gotTable.TableName)
			if n2 < 0 {
				dropTableCmd.TableSchemas = append(dropTableCmd.TableSchemas, gotTable.TableSchema)
				dropTableCmd.TableNames = append(dropTableCmd.TableSchemas, gotTable.TableName)
				continue
			}
			wantTable := wantSchema.Tables[n2]
			// drop columns
			for _, gotColumn := range gotTable.Columns {
				n3 := wantTable.CachedColumnPosition(gotColumn.ColumnName)
				if n3 < 0 {
				}
			}
			// drop constraints
			// drop indexes
			// drop triggers
		}
		// drop views
		for _, gotView := range gotSchema.Views {
			viewPosition := wantSchema.CachedViewPosition(gotView.ViewName)
			if viewPosition < 0 {
				continue
			}
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
			return fmt.Errorf("building command %s: %w", query, err)
		}
		_, err = db.ExecContext(context.Background(), query, args...)
		if err != nil {
			return fmt.Errorf("executing command %s: %w", query, err)
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
				return fmt.Errorf("building command %s: %w", query, err)
			}
			if len(args) > 0 {
				query, err = sq.Sprintf(m.Dialect, query, args)
				if err != nil {
					return fmt.Errorf("building command %s: %w", query, err)
				}
			}
			if !written {
				written = true
			} else {
				io.WriteString(w, "\n\n")
			}
			io.WriteString(w, query)
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
	} {
		for _, cmd := range cmds {
			query, args, _, err := sq.ToSQL(m.Dialect, cmd)
			if err != nil {
				return fmt.Errorf("building command %s: %w", query, err)
			}
			_, err = db.ExecContext(ctx, query, args...)
			if err != nil {
				return fmt.Errorf("executing command %s: %w", query, err)
			}
		}
	}
	return nil
}

func (c *Catalog) Commands() *Migration {
	m := &Migration{Dialect: c.Dialect}
	for _, schema := range c.Schemas {
		if schema.SchemaName != "" {
			createSchemaCmd := &CreateSchemaCommand{
				CreateIfNotExists: true,
				SchemaName:        schema.SchemaName,
			}
			m.SchemaCommands = append(m.SchemaCommands, createSchemaCmd)
		}
		for _, table := range schema.Tables {
			if table.Ignore {
				continue
			}
			createTableCmd := &CreateTableCommand{
				CreateIfNotExists:  true,
				IncludeConstraints: true,
				Table:              table,
			}
			alterTableCmd := &AlterTableCommand{
				TableSchema: table.TableSchema,
				TableName:   table.TableName,
			}
			if c.Dialect == sq.DialectPostgres {
				alterTableCmd.AlterIfExists = true
			}
			var hasForeignKey bool
			for _, constraint := range table.Constraints {
				if constraint.ConstraintType == FOREIGN_KEY && c.Dialect != sq.DialectSQLite {
					hasForeignKey = true
					alterTableCmd.AddConstraintCommands = append(alterTableCmd.AddConstraintCommands, AddConstraintCommand{Constraint: constraint})
				}
			}
			if hasForeignKey {
				m.ForeignKeyCommands = append(m.ForeignKeyCommands, alterTableCmd)
			}
			for _, index := range table.Indexes {
				if index.Ignore {
					continue
				}
				createIndexCmd := &CreateIndexCommand{Index: index}
				if c.Dialect == sq.DialectMySQL {
					createTableCmd.CreateIndexCommands = append(createTableCmd.CreateIndexCommands, *createIndexCmd)
				} else {
					createIndexCmd.CreateIfNotExists = true
					m.IndexCommands = append(m.IndexCommands, createIndexCmd)
				}
			}
			for _, trigger := range table.Triggers {
				if trigger.Ignore {
					continue
				}
				createTriggerCmd := &CreateTriggerCommand{Trigger: trigger}
				m.TriggerCommands = append(m.TriggerCommands, createTriggerCmd)
			}
			m.TableCommands = append(m.TableCommands, createTableCmd)
		}
		for _, view := range schema.Views {
			if view.Ignore {
				continue
			}
			createViewCmd := &CreateViewCommand{View: view}
			if c.Dialect == sq.DialectMySQL || (c.Dialect == sq.DialectPostgres && !view.IsMaterialized) {
				createViewCmd.CreateOrReplace = true
			}
			if c.Dialect == sq.DialectSQLite || (c.Dialect == sq.DialectPostgres && view.IsMaterialized) {
				createViewCmd.CreateIfNotExists = true
			}
			if c.Dialect == sq.DialectPostgres {
				for _, index := range view.Indexes {
					if index.Ignore {
						continue
					}
					createIndexCmd := &CreateIndexCommand{
						CreateIfNotExists: true,
						Index:             index,
					}
					m.IndexCommands = append(m.IndexCommands, createIndexCmd)
				}
				for _, trigger := range view.Triggers {
					if trigger.Ignore {
						continue
					}
					createTriggerCmd := &CreateTriggerCommand{Trigger: trigger}
					m.IndexCommands = append(m.IndexCommands, createTriggerCmd)
				}
			}
			m.ViewCommands = append(m.ViewCommands, createViewCmd)
		}
		for _, function := range schema.Functions {
			if function.Ignore {
				continue
			}
			createFunctionCmd := &CreateFunctionCommand{Function: function}
			if function.IsIndependent {
				m.IndependentFunctionCommands = append(m.IndependentFunctionCommands, createFunctionCmd)
			} else {
				m.FunctionCommands = append(m.FunctionCommands, createFunctionCmd)
			}
		}
	}
	return m
}
