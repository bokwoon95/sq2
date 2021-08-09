package ddl

import (
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

type Migration2 struct {
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
	DropConstraintCmds  []DropConstraintCommand
	DropColumnCmds      []DropColumnCommand
	DropFunctionCmds    []DropFunctionCommand
	DropExtensionCmds   []DropExtensionCommand
}

func Migrate2(mode MigrationMode, gotCatalog, wantCatalog Catalog) (Migration2, error) {
	m := Migration2{
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
	if mode&CreateMissing != 0 || mode&UpdateExisting != 0 {
		if m.Dialect == sq.DialectPostgres && mode&CreateMissing != 0 {
			for _, wantExtension := range wantCatalog.Extensions {
				if n := gotCatalog.CachedExtensionPosition(wantExtension); n >= 0 {
					continue
				}
				createExtensionCmd := CreateExtensionCommand{
					CreateIfNotExists: true,
					Extension:         wantExtension,
				}
				m.CreateExtensionCmds = append(m.CreateExtensionCmds, createExtensionCmd)
			}
		}
		for _, wantSchema := range wantCatalog.Schemas {
			if wantSchema.Ignore {
				continue
			}
			if wantSchema.SchemaName == "" {
				wantSchema.SchemaName = m.CurrentSchema
			}
			gotSchema := Schema{
				SchemaName: wantSchema.SchemaName,
			}
			if n := gotCatalog.CachedSchemaPosition(wantSchema.SchemaName); n >= 0 {
				gotSchema = gotCatalog.Schemas[n]
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
		_ = dropTableCmd
		_ = dropViewCmd
		_ = dropMatViewCmd
		_ = dropExtensionCmd
		for _, gotExtension := range gotCatalog.Extensions {
			if strings.HasPrefix(gotExtension, "plpgsql") {
				continue
			}
			if n := wantCatalog.CachedExtensionPosition(gotExtension); n >= 0 {
				continue
			}
			dropExtensionCmd.Extensions = append(dropExtensionCmd.Extensions, gotExtension)
		}
		for _, gotSchema := range gotCatalog.Schemas {
			wantSchema := Schema{
				SchemaName: gotSchema.SchemaName,
			}
			_ = wantSchema
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
