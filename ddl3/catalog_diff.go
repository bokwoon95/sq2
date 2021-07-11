package ddl3

import (
	"fmt"

	"github.com/bokwoon95/sq"
)

type CatalogDiff struct {
	Dialect          string
	SchemaDiffs      []SchemaDiff
	schemaDiffsCache map[string]int
}

func DiffCatalog(gotCatalog, wantCatalog Catalog) (CatalogDiff, error) {
	var catalogDiff CatalogDiff
	if gotCatalog.Dialect != wantCatalog.Dialect {
		return catalogDiff, fmt.Errorf("dialect mismatch")
	}
	catalogDiff.Dialect = wantCatalog.Dialect
	if !gotCatalog.GeneratedFromDB && wantCatalog.GeneratedFromDB {
		return catalogDiff, fmt.Errorf("GeneratedFromDB mismatch, did you mix gotCatalog and wantCatalog up?")
	}
	gotCatalog.RefreshSchemasCache()
	var err error
	for i, wantSchema := range wantCatalog.Schemas {
		err = diffSchema(catalogDiff.Dialect, &catalogDiff.SchemaDiffs, gotCatalog, wantSchema)
		if err != nil {
			return catalogDiff, fmt.Errorf("schema #%d %s: %w", i+1, wantSchema.SchemaName, err)
		}
	}
	return catalogDiff, nil
}

func diffSchema(dialect string, schemaDiffs *[]SchemaDiff, gotCatalog Catalog, wantSchema Schema) error {
	schemaDiff := SchemaDiff{
		SchemaName: wantSchema.SchemaName,
	}
	var gotSchema Schema
	schemaName := wantSchema.SchemaName
	if schemaName == "" {
		schemaName = gotCatalog.DefaultSchema
	}
	if schemaName != "" {
		if n := gotCatalog.CachedSchemaPosition(schemaName); n >= 0 {
			gotSchema = gotCatalog.Schemas[n]
			gotSchema.RefreshTableCache()
			gotSchema.RefreshViewsCache()
			gotSchema.RefreshFunctionsCache()
		} else {
			schemaDiff.CreateCommand = &CreateSchemaCommand{
				CreateIfNotExists: true,
				SchemaName:        wantSchema.SchemaName,
			}
		}
	}
	var err error
	for i, wantTable := range wantSchema.Tables {
		err = diffTable(dialect, &schemaDiff.TableDiffs, gotSchema, wantTable)
		if err != nil {
			return fmt.Errorf("table #%d %s: %w", i+1, wantTable.TableName, err)
		}
	}
	for _, wantView := range wantSchema.Views {
		if n := gotSchema.CachedViewPosition(wantView.ViewName); n < 0 {
			schemaDiff.ViewDiffs = append(schemaDiff.ViewDiffs, ViewDiff{
				ViewSchema:    wantSchema.SchemaName,
				ViewName:      wantView.ViewName,
				CreateCommand: &CreateViewCommand{View: wantView},
			})
		}
	}
	for _, wantFunction := range wantSchema.Functions {
		if n := gotSchema.CachedViewPosition(wantFunction.FunctionName); n < 0 {
			schemaDiff.FunctionDiffs = append(schemaDiff.FunctionDiffs, FunctionDiff{
				FunctionSchema: wantSchema.SchemaName,
				FunctionName:   wantFunction.FunctionName,
				CreateCommand:  &CreateFunctionCommand{Function: wantFunction},
			})
		}
	}
	if schemaDiff.CreateCommand != nil || schemaDiff.DropCommand != nil || schemaDiff.RenameCommand != nil || len(schemaDiff.TableDiffs) > 0 || len(schemaDiff.ViewDiffs) > 0 || len(schemaDiff.FunctionDiffs) > 0 {
		*schemaDiffs = append(*schemaDiffs, schemaDiff)
	}
	return nil
}

func diffTable(dialect string, tableDiffs *[]TableDiff, gotSchema Schema, wantTable Table) error {
	tableDiff := TableDiff{
		TableSchema: wantTable.TableSchema,
		TableName:   wantTable.TableName,
	}
	var gotTable Table
	if n := gotSchema.CachedTablePosition(wantTable.TableName); n >= 0 {
		gotTable = gotSchema.Tables[n]
		gotTable.RefreshColumnsCache()
		gotTable.RefreshConstraintsCache()
		gotTable.RefreshIndexesCache()
		gotTable.RefreshTriggersCache()
	} else {
		tableDiff.CreateCommand = &CreateTableCommand{
			CreateIfNotExists:  true,
			IncludeConstraints: true,
			Table:              wantTable,
		}
	}
	var err error
	for i, wantColumn := range wantTable.Columns {
		if tableDiff.CreateCommand != nil {
			break
		}
		err = diffColumn(dialect, &tableDiff.ColumnDiffs, gotTable, wantColumn)
		if err != nil {
			return fmt.Errorf("column #%d %s: %w", i+1, wantColumn.ColumnName, err)
		}
	}
	for _, wantConstraint := range wantTable.Constraints {
		if n := gotTable.CachedConstraintPosition(wantConstraint.ConstraintName); n < 0 {
			tableDiff.ConstraintDiffs = append(tableDiff.ConstraintDiffs, ConstraintDiff{
				TableSchema:    wantTable.TableSchema,
				TableName:      wantTable.TableName,
				ConstraintName: wantConstraint.ConstraintName,
				ConstraintType: wantConstraint.ConstraintType,
				AddCommand: &AddConstraintCommand{
					AlterTableIfExists: true,
					TableSchema:        wantTable.TableSchema,
					TableName:          wantTable.TableName,
					AddIfNotExists:     true,
					Constraint:         wantConstraint,
				},
			})
		}
	}
	for _, wantIndex := range wantTable.Indexes {
		if n := gotTable.CachedIndexPosition(wantIndex.IndexName); n < 0 {
			tableDiff.IndexDiffs = append(tableDiff.IndexDiffs, IndexDiff{
				TableSchema: wantTable.TableSchema,
				TableName:   wantTable.TableName,
				IndexName:   wantIndex.IndexName,
				IndexType:   wantIndex.IndexType,
				CreateCommand: &CreateIndexCommand{
					CreateIfNotExists: true,
					Index:             wantIndex,
				},
			})
		}
	}
	for _, wantTrigger := range wantTable.Triggers {
		if n := gotTable.CachedTriggerPosition(wantTrigger.TriggerName); n < 0 {
			tableDiff.TriggerDiffs = append(tableDiff.TriggerDiffs, TriggerDiff{
				TableSchema: wantTable.TableSchema,
				TableName:   wantTable.TableName,
				TriggerName: wantTrigger.TriggerName,
				CreateCommand: &CreateTriggerCommand{
					Trigger: wantTrigger,
				},
			})
		}
	}
	if tableDiff.CreateCommand != nil || tableDiff.DropCommand != nil || tableDiff.RenameCommand != nil || len(tableDiff.ColumnDiffs) > 0 || len(tableDiff.ConstraintDiffs) > 0 || len(tableDiff.IndexDiffs) > 0 || len(tableDiff.TriggerDiffs) > 0 {
		*tableDiffs = append(*tableDiffs, tableDiff)
	}
	return nil
}

func diffColumn(dialect string, columnDiffs *[]ColumnDiff, gotTable Table, wantColumn Column) error {
	columnDiff := ColumnDiff{
		TableSchema: wantColumn.TableSchema,
		TableName:   wantColumn.TableName,
		ColumnName:  wantColumn.ColumnName,
	}
	var gotColumn Column
	if n := gotTable.CachedColumnPosition(wantColumn.ColumnName); n >= 0 {
		gotColumn = gotTable.Columns[n]
	} else {
		columnDiff.AddCommand = &AddColumnCommand{
			AlterTableIfExists: true,
			TableSchema:        wantColumn.TableSchema,
			TableName:          wantColumn.TableName,
			AddIfNotExists:     true,
			Column:             wantColumn,
		}
		return nil
	}
	alterCmd := &AlterColumnCommand{
		AlterTableIfExists: true,
		AlterIfExists:      true,
		Column:             wantColumn,
	}
	err := diffColumnType(&alterCmd.Column, gotColumn, wantColumn)
	if err != nil {
		return fmt.Errorf("diffing column type: %w", err)
	}
	if gotColumn.Identity == wantColumn.Identity {
		alterCmd.Column.Identity = ""
	} else if gotColumn.Identity != "" && wantColumn.Identity == "" {
		alterCmd.DropIdentity = true
	}
	if gotColumn.IsNotNull == wantColumn.IsNotNull {
		alterCmd.Column.IsNotNull = false
	} else if gotColumn.IsNotNull && !wantColumn.IsNotNull {
		alterCmd.DropNotNull = true
	}
	// NOTE: we are not touching generated exprs because their equality simply
	// cannot be defined without parsing the SQL. For the same reason, we are
	// not touching default expressions.
	if gotColumn.CollationName == wantColumn.CollationName {
		// Collation cannot be dropped, so we can only overwrite it or leave it alone
		alterCmd.Column.CollationName = ""
	}
	if gotColumn.ColumnDefault != "" && wantColumn.ColumnDefault == "" {
		alterCmd.Column.ColumnDefault = ""
	}
	if columnDiff.AddCommand != nil || columnDiff.AlterCommand != nil || columnDiff.DropCommand != nil || columnDiff.RenameCommand != nil || columnDiff.ReplaceCommand != nil {
		*columnDiffs = append(*columnDiffs, columnDiff)
	}
	return nil
}

func diffColumnType(column *Column, gotColumn, wantColumn Column) error {
	if (*column).ColumnType == "" {
		*column = wantColumn
	}
	if gotColumn.ColumnType == wantColumn.ColumnType {
		column.ColumnType = ""
	}
	return nil
}

func (catalogDiff CatalogDiff) Commands() CommandSet {
	cmdset := CommandSet{Dialect: catalogDiff.Dialect}
	for _, schemaDiff := range catalogDiff.SchemaDiffs {
		if schemaDiff.CreateCommand != nil {
			cmdset.SchemaCommands = append(cmdset.SchemaCommands, schemaDiff.CreateCommand)
		}
		for _, tableDiff := range schemaDiff.TableDiffs {
			if tableDiff.CreateCommand != nil {
				cmdset.TableCommands = append(cmdset.TableCommands, tableDiff.CreateCommand)
			}
			for _, columnDiff := range tableDiff.ColumnDiffs {
				if columnDiff.AddCommand != nil {
					cmdset.TableCommands = append(cmdset.TableCommands, columnDiff.AddCommand)
				}
				// TODO: implement AlterColumnCommand
				// if columnDiff.AlterCommand != nil {
				// 	tableCmds = append(tableCmds, columnDiff.AlterCommand)
				// }
			}
			constraintsAlreadyIncluded := tableDiff.CreateCommand != nil && tableDiff.CreateCommand.IncludeConstraints
			for _, constraintDiff := range tableDiff.ConstraintDiffs {
				if constraintDiff.ConstraintType == PRIMARY_KEY && catalogDiff.Dialect == sq.DialectSQLite {
					continue
				}
				if constraintDiff.ConstraintType == FOREIGN_KEY && catalogDiff.Dialect != sq.DialectSQLite {
					if constraintDiff.AddCommand != nil {
						cmdset.ForeignKeyCommands = append(cmdset.ForeignKeyCommands, constraintDiff.AddCommand)
					}
					continue
				}
				if !constraintsAlreadyIncluded && constraintDiff.AddCommand != nil {
					cmdset.TableCommands = append(cmdset.TableCommands, constraintDiff.AddCommand)
				}
			}
			for _, indexDiff := range tableDiff.IndexDiffs {
				if indexDiff.CreateCommand != nil {
					cmdset.TableCommands = append(cmdset.TableCommands, indexDiff.CreateCommand)
				}
			}
			for _, triggerDiff := range tableDiff.TriggerDiffs {
				if triggerDiff.CreateCommand != nil {
					cmdset.TableCommands = append(cmdset.TableCommands, triggerDiff.CreateCommand)
				}
			}
		}
		for _, functionDiff := range schemaDiff.FunctionDiffs {
			if functionDiff.CreateCommand != nil {
				if functionDiff.CreateCommand.Function.ContainsTable {
					cmdset.TableFunctionCommands = append(cmdset.TableFunctionCommands, functionDiff.CreateCommand)
				} else {
					cmdset.FunctionCommands = append(cmdset.FunctionCommands, functionDiff.CreateCommand)
				}
			}
		}
		for _, viewDiff := range schemaDiff.ViewDiffs {
			if viewDiff.CreateCommand != nil {
				cmdset.ViewCommands = append(cmdset.ViewCommands, viewDiff.CreateCommand)
			}
			for _, triggerDiff := range viewDiff.TriggerDiffs {
				if triggerDiff.CreateCommand != nil {
					cmdset.ViewCommands = append(cmdset.ViewCommands, triggerDiff.CreateCommand)
				}
			}
		}
	}
	return cmdset
}
