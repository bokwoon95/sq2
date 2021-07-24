package ddl

import (
	"context"
	"fmt"
	"io"

	"github.com/bokwoon95/sq"
)

type MigrationOption int

const (
	CreateMissing  MigrationOption = 0b1
	UpdateExisting MigrationOption = 0b10
	DropExtraneous MigrationOption = 0b100
	DropCascade    MigrationOption = 0b1000
)

type MigrationCommands struct {
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
}

func AutoMigrate(dialect string, db sq.DB, migrationOption MigrationOption, CatalogOptions ...CatalogOption) error {
	return nil
}

func Migrate(migrationOption MigrationOption, wantCatalog, gotCatalog Catalog) (MigrationCommands, error) {
	var m MigrationCommands
	return m, nil
}

func (m *MigrationCommands) WriteSQL(w io.Writer) error {
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
			if table.Ignore {
				continue
			}
			createTableCmd := &CreateTableCommand{
				CreateIfNotExists:  true,
				IncludeConstraints: true,
				Table:              *table,
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
					alterTableCmd.AddConstraintCommands = append(alterTableCmd.AddConstraintCommands, AddConstraintCommand{Constraint: *constraint})
				}
			}
			if hasForeignKey {
				m.ForeignKeyCommands = append(m.ForeignKeyCommands, alterTableCmd)
			}
			for _, index := range table.Indexes {
				if index.Ignore {
					continue
				}
				createIndexCmd := &CreateIndexCommand{Index: *index}
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
				createTriggerCmd := &CreateTriggerCommand{Trigger: *trigger}
				m.TriggerCommands = append(m.TriggerCommands, createTriggerCmd)
			}
			m.TableCommands = append(m.TableCommands, createTableCmd)
		}
		for _, view := range schema.Views {
			if view.Ignore {
				continue
			}
			createViewCmd := &CreateViewCommand{View: *view}
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
						Index:             *index,
					}
					m.IndexCommands = append(m.IndexCommands, createIndexCmd)
				}
				for _, trigger := range view.Triggers {
					if trigger.Ignore {
						continue
					}
					createTriggerCmd := &CreateTriggerCommand{Trigger: *trigger}
					m.IndexCommands = append(m.IndexCommands, createTriggerCmd)
				}
			}
			m.ViewCommands = append(m.ViewCommands, createViewCmd)
		}
		for _, function := range schema.Functions {
			if function.Ignore {
				continue
			}
			createFunctionCmd := &CreateFunctionCommand{Function: *function}
			if function.IsIndependent {
				m.IndependentFunctionCommands = append(m.IndependentFunctionCommands, createFunctionCmd)
			} else {
				m.FunctionCommands = append(m.FunctionCommands, createFunctionCmd)
			}
		}
	}
	return m
}
