package ddl

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/bokwoon95/sq"
)

//go:embed sql
var sqlFS embed.FS

func introspectQuery(ctx context.Context, db sq.DB, catalog *Catalog, queryfile string, argslist [][]interface{}, rowmapper func(*Catalog, *sql.Rows) error) error {
	b, err := fs.ReadFile(sqlFS, queryfile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", queryfile, err)
	}
	stmt, err := db.PrepareContext(ctx, string(b))
	if err != nil {
		return fmt.Errorf("preparing %s: %w", queryfile, err)
	}
	defer stmt.Close()
	if len(argslist) == 0 {
		argslist = append(argslist, nil)
	}
	for _, args := range argslist {
		err = func() error {
			rows, err := stmt.QueryContext(ctx, args...)
			if err != nil {
				return fmt.Errorf("executing %s: %w", queryfile, err)
			}
			defer rows.Close()
			for rows.Next() {
				err = rowmapper(catalog, rows)
				if err != nil {
					return err
				}
			}
			err = rows.Close()
			if err != nil {
				return err
			}
			err = rows.Err()
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func normalizeColumn(dialect string, column *Column, columnType2 string) {
	if column.ColumnDefault != "" && isExpression(column.ColumnDefault) && dialect != sq.DialectPostgres {
		column.ColumnDefault = "(" + column.ColumnDefault + ")"
	}
	switch dialect {
	case sq.DialectPostgres:
		if (strings.EqualFold(column.ColumnType, "NUMERIC") || strings.EqualFold(column.ColumnType, "DECIMAL")) && (column.NumericPrecision > 0 || column.NumericScale > 0) {
			column.ColumnType = fmt.Sprintf("%s(%d,%d)", column.ColumnType, column.NumericPrecision, column.NumericScale)
		}
		if strings.EqualFold(column.Identity, "BY DEFAULT") {
			column.Identity = BY_DEFAULT_AS_IDENTITY
		} else if strings.EqualFold(column.Identity, "ALWAYS") {
			column.Identity = ALWAYS_AS_IDENTITY
		} else {
			column.Identity = ""
		}
		if strings.EqualFold(column.ColumnType, "TIMESTAMP WITH TIME ZONE") {
			column.ColumnType = "TIMESTAMPTZ"
		} else if strings.EqualFold(column.ColumnType, "USER-DEFINED") {
			column.ColumnType = columnType2
		} else if strings.EqualFold(column.ColumnType, "ARRAY") {
			column.ColumnType = "[]" + columnType2[1:]
		} else {
			column.ColumnType = strings.ToUpper(column.ColumnType)
		}
		if len(column.GeneratedExpr) > 2 {
			last := len(column.GeneratedExpr) - 1
			if column.GeneratedExpr[0] == '(' && column.GeneratedExpr[last] == ')' {
				column.GeneratedExpr = column.GeneratedExpr[1:last]
			}
		}
	}
}

func mapTables(catalog *Catalog, rows *sql.Rows) error {
	var tbl Table
	err := rows.Scan(
		&tbl.TableSchema,
		&tbl.TableName,
	)
	if err != nil {
		return fmt.Errorf("scanning table %s: %w", tbl.TableName, err)
	}
	var schema Schema
	if n := catalog.CachedSchemaPosition(tbl.TableSchema); n >= 0 {
		schema = catalog.Schemas[n]
		defer func() { catalog.Schemas[n] = schema }()
	} else {
		schema.SchemaName = tbl.TableSchema
		defer func() { catalog.AppendSchema(schema) }()
	}
	if n := schema.CachedTablePosition(tbl.TableName); n >= 0 {
		schema.Tables[n] = tbl
	} else {
		schema.AppendTable(tbl)
	}
	return nil
}

func mapColumns(catalog *Catalog, rows *sql.Rows) error {
	var column Column
	var columnType2 string
	err := rows.Scan(
		&column.TableSchema,
		&column.TableName,
		&column.ColumnName,
		&column.ColumnType,
		&columnType2,
		&column.NumericPrecision,
		&column.NumericScale,
		&column.IsAutoincrement,
		&column.Identity,
		&column.IsNotNull,
		&column.OnUpdateCurrentTimestamp,
		&column.GeneratedExpr,
		&column.GeneratedExprStored,
		&column.CollationName,
		&column.ColumnDefault,
	)
	if err != nil {
		return fmt.Errorf("scanning column %s.%s: %w", column.TableName, column.ColumnName, err)
	}
	normalizeColumn(catalog.Dialect, &column, columnType2)
	var schema Schema
	if n := catalog.CachedSchemaPosition(column.TableSchema); n >= 0 {
		schema = catalog.Schemas[n]
		defer func() { catalog.Schemas[n] = schema }()
	} else {
		schema.SchemaName = column.TableSchema
		defer func() { catalog.AppendSchema(schema) }()
	}
	var tbl Table
	if n := schema.CachedTablePosition(column.TableName); n >= 0 {
		tbl = schema.Tables[n]
		defer func() { schema.Tables[n] = tbl }()
	} else {
		tbl.TableSchema = column.TableSchema
		tbl.TableName = column.TableName
		defer func() { schema.AppendTable(tbl) }()
	}
	if n := tbl.CachedColumnPosition(column.ColumnName); n >= 0 {
		tbl.Columns[n] = column
	} else {
		tbl.AppendColumn(column)
	}
	return nil
}

func introspectPostgres(ctx context.Context, db sq.DB, catalog *Catalog) error {
	err := introspectQuery(ctx, db, catalog, "sql/postgres_columns.sql", nil, mapColumns)
	if err != nil {
		return err
	}
	return nil
}

func introspectSQLite(ctx context.Context, db sq.DB, catalog *Catalog) error {
	err := introspectQuery(ctx, db, catalog, "sql/sqlite_tables.sql", nil, mapTables)
	if err != nil {
		return err
	}
	var argslist [][]interface{}
	for _, schema := range catalog.Schemas {
		for _, tbl := range schema.Tables {
			argslist = append(argslist, []interface{}{tbl.TableName})
		}
	}
	err = introspectQuery(ctx, db, catalog, "sql/sqlite_columns.sql", argslist, mapColumns)
	if err != nil {
		return err
	}
	return nil
}
