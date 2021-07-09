package ddl3

import (
	"fmt"

	"github.com/bokwoon95/sq"
)

type CatalogDiff struct {
	SchemaDiffs      []SchemaDiff
	schemaDiffsCache map[string]int // 8 bytes
}

// pass down dialect and default schema
func DiffCatalog(gotCatalog, wantCatalog Catalog) (CatalogDiff, error) {
	var catalogDiff CatalogDiff
	if gotCatalog.Dialect != wantCatalog.Dialect {
		return catalogDiff, fmt.Errorf("dialect mismatch")
	}
	if !gotCatalog.GeneratedFromDB && wantCatalog.GeneratedFromDB {
		return catalogDiff, fmt.Errorf("GeneratedFromDB mismatch, did you mix gotCatalog and wantCatalog up?")
	}
	gotCatalog.RefreshSchemasCache()
	for _, wantSchema := range wantCatalog.Schemas {
		schemaDiff := SchemaDiff{
			SchemaName: wantSchema.SchemaName,
		}
		var gotSchema Schema
		if n := gotCatalog.CachedSchemaPosition(wantSchema.SchemaName); n < 0 {
			schemaDiff.CreateCommand.Valid = true
			schemaDiff.CreateCommand.CreateIfNotExists = true
			schemaDiff.CreateCommand.SchemaName = schemaDiff.SchemaName
		} else {
			gotSchema = gotCatalog.Schemas[n]
			gotSchema.RefreshTableCache()
			gotSchema.RefreshViewsCache()
			gotSchema.RefreshFunctionsCache()
		}
		for _, wantTable := range wantSchema.Tables {
			tableDiff := TableDiff{
				TableSchema: schemaDiff.SchemaName,
				TableName:   wantTable.TableName,
			}
			var gotTable Table
			if n := gotSchema.CachedTablePosition(wantTable.TableName); n < 0 {
				tableDiff.CreateCommand.Valid = true
				tableDiff.CreateCommand.CreateIfNotExists = true
				tableDiff.CreateCommand.Table = wantTable
			} else {
				gotTable = gotSchema.Tables[n]
				gotTable.RefreshColumnsCache()
				gotTable.RefreshConstraintsCache()
				gotTable.RefreshIndexesCache()
				gotTable.RefreshTriggersCache()
			}
			for _, wantColumn := range wantTable.Columns {
				if tableDiff.CreateCommand.Valid {
					break
				}
				columnDiff := ColumnDiff{
					TableSchema: tableDiff.TableSchema,
					TableName:   tableDiff.TableName,
					ColumnName:  wantColumn.ColumnName,
				}
				var gotColumn Column
				if n := gotTable.CachedColumnPosition(wantColumn.ColumnName); n >= 0 {
					gotColumn = gotTable.Columns[n]
				} else {
					columnDiff.AddCommand.Valid = true
					columnDiff.AddCommand.AddIfNotExists = true
					columnDiff.AddCommand.Column = wantColumn
					if wantCatalog.Dialect == sq.DialectSQLite {
					}
				}
				_ = gotColumn
			}
			for _, wantConstraint := range wantTable.Constraints {
				if n := gotTable.CachedConstraintPosition(wantConstraint.ConstraintName); n < 0 {
				}
			}
			for _, wantIndex := range wantTable.Indexes {
				if n := gotTable.CachedIndexPosition(wantIndex.IndexName); n < 0 {
				}
			}
			for _, wantTrigger := range wantTable.Triggers {
				if n := gotTable.CachedTriggerPosition(wantTrigger.TriggerName); n < 0 {
				}
			}
			if tableDiff.CreateCommand.Valid || tableDiff.DropCommand.Valid || tableDiff.RenameCommand.Valid || len(tableDiff.ColumnDiffs) > 0 || len(tableDiff.ConstraintDiffs) > 0 || len(tableDiff.IndexDiffs) > 0 || len(tableDiff.TriggerDiffs) > 0 {
				schemaDiff.TableDiffs = append(schemaDiff.TableDiffs, tableDiff)
			}
		}
		for _, wantView := range wantSchema.Views {
			if n := gotSchema.CachedViewPosition(wantView.ViewName); n < 0 {
			}
		}
		for _, wantFunction := range wantSchema.Functions {
			if ns := gotSchema.CachedFunctionPositions(wantFunction.FunctionName); len(ns) == 0 {
			}
		}
		if schemaDiff.CreateCommand.Valid || schemaDiff.DropCommand.Valid || schemaDiff.RenameCommand.Valid || len(schemaDiff.TableDiffs) > 0 || len(schemaDiff.ViewDiffs) > 0 || len(schemaDiff.FunctionDiffs) > 0 {
			catalogDiff.SchemaDiffs = append(catalogDiff.SchemaDiffs, schemaDiff)
		}
	}
	return catalogDiff, nil
}

func DiffSchema(dialect string, schemaDiffs *[]SchemaDiff, gotCatalog Catalog, wantSchema Schema) error {
	schemaDiff := SchemaDiff{
		SchemaName: wantSchema.SchemaName,
	}
	var gotSchema Schema
	if n := gotCatalog.CachedSchemaPosition(wantSchema.SchemaName); n >= 0 {
		gotSchema = gotCatalog.Schemas[n]
		gotSchema.RefreshTableCache()
		gotSchema.RefreshViewsCache()
		gotSchema.RefreshFunctionsCache()
	} else {
		schemaDiff.CreateCommand.Valid = true
		schemaDiff.CreateCommand.CreateIfNotExists = true
		schemaDiff.CreateCommand.SchemaName = schemaDiff.SchemaName
	}
	var err error
	for i, wantTable := range wantSchema.Tables {
		err = DiffTable(dialect, &schemaDiff.TableDiffs, gotSchema, wantTable)
		if err != nil {
			return fmt.Errorf("table #%d %s: %w", i+1, wantTable.TableName, err)
		}
	}
	for _, wantView := range wantSchema.Views {
		if n := gotSchema.CachedViewPosition(wantView.ViewName); n < 0 {
			schemaDiff.ViewDiffs = append(schemaDiff.ViewDiffs, ViewDiff{
				ViewSchema:    wantSchema.SchemaName,
				ViewName:      wantView.ViewName,
				CreateCommand: CreateViewCommand{Valid: true, View: wantView},
			})
		}
	}
	for _, wantFunction := range wantSchema.Functions {
		if n := gotSchema.CachedViewPosition(wantFunction.FunctionName); n < 0 {
			schemaDiff.FunctionDiffs = append(schemaDiff.FunctionDiffs, FunctionDiff{
				FunctionSchema: wantSchema.SchemaName,
				FunctionName:   wantFunction.FunctionName,
				CreateCommand:  CreateFunctionCommand{Valid: true, Function: wantFunction},
			})
		}
	}
	if schemaDiff.CreateCommand.Valid || schemaDiff.DropCommand.Valid || schemaDiff.RenameCommand.Valid || len(schemaDiff.TableDiffs) > 0 || len(schemaDiff.ViewDiffs) > 0 || len(schemaDiff.FunctionDiffs) > 0 {
		*schemaDiffs = append(*schemaDiffs, schemaDiff)
	}
	return nil
}

func DiffTable(dialect string, tableDiffs *[]TableDiff, gotSchema Schema, wantTable Table) error {
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
		tableDiff.CreateCommand.Valid = true
		tableDiff.CreateCommand.CreateIfNotExists = true
		tableDiff.CreateCommand.Table = wantTable
	}
	var err error
	for i, wantColumn := range wantTable.Columns {
		if tableDiff.CreateCommand.Valid {
			break
		}
		err = DiffColumn(dialect, &tableDiff.ColumnDiffs, gotTable, wantColumn)
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
				AddCommand: AddConstraintCommand{
					Valid:              true,
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
				CreateCommand: CreateIndexCommand{
					Valid:             true,
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
				CreateCommand: CreateTriggerCommand{
					Valid:   true,
					Trigger: wantTrigger,
				},
			})
		}
	}
	if tableDiff.CreateCommand.Valid || tableDiff.DropCommand.Valid || tableDiff.RenameCommand.Valid || len(tableDiff.ColumnDiffs) > 0 || len(tableDiff.ConstraintDiffs) > 0 || len(tableDiff.IndexDiffs) > 0 || len(tableDiff.TriggerDiffs) > 0 {
		*tableDiffs = append(*tableDiffs, tableDiff)
	}
	return nil
}

func DiffColumn(dialect string, columnDiffs *[]ColumnDiff, gotTable Table, wantColumn Column) error {
	columnDiff := ColumnDiff{
		TableSchema: wantColumn.TableSchema,
		TableName:   wantColumn.TableName,
		ColumnName:  wantColumn.ColumnName,
	}
	var gotColumn Column
	if n := gotTable.CachedColumnPosition(wantColumn.ColumnName); n >= 0 {
		gotColumn = gotTable.Columns[n]
	} else {
		columnDiff.AddCommand.Valid = true
		columnDiff.AddCommand.AlterTableIfExists = true
		columnDiff.AddCommand.TableSchema = columnDiff.TableSchema
		columnDiff.AddCommand.TableName = columnDiff.TableName
		columnDiff.AddCommand.AddIfNotExists = true
		columnDiff.AddCommand.Column = wantColumn
		return nil
	}
	alterCmd := AlterColumnCommand{
		Valid:              true,
		AlterTableIfExists: true,
		AlterIfExists:      true,
		Column:             wantColumn,
	}
	if hasEquivalentColumnTypes(gotColumn, wantColumn) {
		alterCmd.Column.ColumnType = ""
	}
	if gotColumn.Identity == wantColumn.Identity {
		alterCmd.Column.Identity = ""
	} else if gotColumn.Identity != "" && wantColumn.Identity == "" {
		alterCmd.DropIdentity = true
	}
	if gotColumn.Autoincrement == wantColumn.Autoincrement {
		alterCmd.Column.Autoincrement = false
	} else if gotColumn.Autoincrement && !wantColumn.Autoincrement {
		// TODO: I think MySQL doesn't allow for dropping autoincrement without
		// dropping primary key, need to investigate further. SQLite doesn't
		// even allow dropping any constraint in the first place.
		alterCmd.DropAutoincrement = true
	}
	if gotColumn.IsNotNull == wantColumn.IsNotNull {
		alterCmd.Column.IsNotNull = false
	} else if gotColumn.IsNotNull && !wantColumn.IsNotNull {
		alterCmd.DropNotNull = true
	}
	if columnDiff.AddCommand.Valid || columnDiff.AlterCommand.Valid || columnDiff.DropCommand.Valid || columnDiff.RenameCommand.Valid || columnDiff.ReplaceCommand.Valid {
		*columnDiffs = append(*columnDiffs, columnDiff)
	}
	return nil
}
