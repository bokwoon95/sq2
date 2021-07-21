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

func runQuery(ctx context.Context, db sq.DB, catalog *Catalog, filename string, rowmapper func(*Catalog, *sql.Rows) error) error {
	b, err := fs.ReadFile(sqlFS, filename)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filename, err)
	}
	rows, err := db.QueryContext(ctx, string(b))
	if err != nil {
		return fmt.Errorf("executing %s: %w", filename, err)
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
	return rows.Err()
}

func normalizeColumn(dialect string, column *Column, columnType2 string) {
	switch dialect {
	case sq.DialectPostgres:
		if (strings.EqualFold(column.ColumnType, "NUMERIC") || strings.EqualFold(column.ColumnType, "DECIMAL")) && (column.Precision > 0 || column.Scale > 0) {
			column.ColumnType = fmt.Sprintf("%s(%d,%d)", column.ColumnType, column.Precision, column.Scale)
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

func columnRowMapper(catalog *Catalog, rows *sql.Rows) error {
	var column Column
	var columnType2 string
	err := rows.Scan(
		&column.TableSchema,
		&column.TableName,
		&column.ColumnName,
		&column.ColumnType,
		&columnType2,
		&column.Precision,
		&column.Scale,
		&column.Autoincrement,
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
	err := runQuery(ctx, db, catalog, "sql/postgres-column.sql", columnRowMapper)
	if err != nil {
		return err
	}
	return nil
}
