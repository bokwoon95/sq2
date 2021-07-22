package ddl

import (
	"context"
	"fmt"
	"io"

	"github.com/bokwoon95/sq"
)

type MigrationCommands struct {
	Dialect                   string
	SchemaCommands            []Command
	ExtensionCommands         []Command
	EnumCommands              []Command
	FunctionCommands          []Command
	TableCommands             []Command
	ViewCommands              []Command
	IndexCommands             []Command
	DependentFunctionCommands []Command
	TriggerCommands           []Command
	ForeignKeyCommands        []Command
}

func (c *Catalog) Commands() *MigrationCommands {
	m := &MigrationCommands{Dialect: c.Dialect}
	for _, schema := range c.Schemas {
		if schema.SchemaName != "" {
			createSchemaCmd := &CreateSchemaCommand{
				CreateIfNotExists: true,
				SchemaName:        schema.SchemaName,
			}
			m.SchemaCommands = append(m.SchemaCommands, createSchemaCmd)
		}
		for _, table := range schema.Tables {
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
			var indexCmds []Command
			for _, index := range table.Indexes {
				createIndexCmd := &CreateIndexCommand{Index: index}
				if c.Dialect == sq.DialectMySQL {
					createTableCmd.CreateIndexCommands = append(createTableCmd.CreateIndexCommands, *createIndexCmd)
				} else {
					createIndexCmd.CreateIfNotExists = true
					indexCmds = append(indexCmds, createIndexCmd)
				}
			}
			for _, trigger := range table.Triggers {
				createTriggerCmd := &CreateTriggerCommand{Trigger: trigger}
				m.TriggerCommands = append(m.TriggerCommands, createTriggerCmd)
			}
			m.TableCommands = append(m.TableCommands, createTableCmd)
			m.TableCommands = append(m.TableCommands, indexCmds...)
		}
		for _, view := range schema.Views {
			createViewCmd := &CreateViewCommand{View: view}
			if c.Dialect == sq.DialectMySQL || (c.Dialect == sq.DialectPostgres && !view.IsMaterialized) {
				createViewCmd.CreateOrReplace = true
			}
			if c.Dialect == sq.DialectSQLite || (c.Dialect == sq.DialectPostgres && view.IsMaterialized) {
				createViewCmd.CreateIfNotExists = true
			}
			m.ViewCommands = append(m.ViewCommands, createViewCmd)
		}
		for _, function := range schema.Functions {
			createFunctionCmd := &CreateFunctionCommand{Function: function}
			m.FunctionCommands = append(m.FunctionCommands, createFunctionCmd)
		}
	}
	return m
}

func (m *MigrationCommands) WriteSQL(w io.Writer) error {
	var written bool
	for _, cmds := range [][]Command{
		m.SchemaCommands,
		m.FunctionCommands,
		m.TableCommands,
		m.ViewCommands,
		m.TriggerCommands,
		m.ForeignKeyCommands,
	} {
		for _, cmd := range cmds {
			query, args, _, err := sq.ToSQL(m.Dialect, cmd)
			if err != nil {
				return fmt.Errorf("command: %s: %w", query, err)
			}
			if len(args) > 0 {
				query, err = sq.Sprintf(m.Dialect, query, args)
				if err != nil {
					return fmt.Errorf("command: %s: %w", query, err)
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

func (m *MigrationCommands) Exec(db sq.DB) error {
	return m.ExecContext(context.Background(), db)
}

func (m *MigrationCommands) ExecContext(ctx context.Context, db sq.DB) error {
	for _, cmds := range [][]Command{
		m.SchemaCommands,
		m.FunctionCommands,
		m.TableCommands,
		m.ViewCommands,
		m.TriggerCommands,
		m.ForeignKeyCommands,
	} {
		for _, cmd := range cmds {
			query, args, _, err := sq.ToSQL(m.Dialect, cmd)
			if err != nil {
				return fmt.Errorf("command: %s: %w", query, err)
			}
			_, err = db.ExecContext(ctx, query, args...)
			if err != nil {
				return fmt.Errorf("command: %s: %w", query, err)
			}
		}
	}
	return nil
}
