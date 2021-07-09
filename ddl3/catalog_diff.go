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
