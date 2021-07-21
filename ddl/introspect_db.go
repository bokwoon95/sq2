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

func postgresColumnRowMapper(catalog *Catalog, rows *sql.Rows) error {
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
		&column.Identity, // BY DEFAULT | ALWAYS
		&column.IsNotNull,
		&column.GeneratedExpr,
		&column.GeneratedExprStored,
		&column.CollationName,
		&column.ColumnDefault,
	)
	if err != nil {
		return fmt.Errorf("scanning column %s.%s: %w", column.TableName, column.ColumnName, err)
	}
	if strings.EqualFold(column.Identity, "BY DEFAULT") {
		column.Identity = BY_DEFAULT_AS_IDENTITY
	} else if strings.EqualFold(column.Identity, "ALWAYS") {
		column.Identity = ALWAYS_AS_IDENTITY
	} else {
		column.Identity = ""
	}
	if i := strings.IndexByte(column.ColumnType, ' '); i >= 0 {
		column.ColumnType = strings.ToUpper(columnType2)
	} else {
		column.ColumnType = strings.ToUpper(column.ColumnType)
	}
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
	err := runQuery(ctx, db, catalog, "sql/postgres-columns.sql", postgresColumnRowMapper)
	if err != nil {
		return err
	}
	return nil
}
