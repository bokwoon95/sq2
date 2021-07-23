package ddl

import (
	"bytes"
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
	if column.ColumnDefault != "" {
		column.ColumnDefault = toExpr(dialect, column.ColumnDefault)
	}
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
	if len(column.GeneratedExpr) > 2 {
		last := len(column.GeneratedExpr) - 1
		if column.GeneratedExpr[0] == '(' && column.GeneratedExpr[last] == ')' {
			column.GeneratedExpr = column.GeneratedExpr[1:last]
		}
	}
	switch dialect {
	case sq.DialectPostgres:
		if strings.EqualFold(column.ColumnType, "TIMESTAMP WITH TIME ZONE") {
			column.ColumnType = "TIMESTAMPTZ"
		} else if strings.EqualFold(column.ColumnType, "USER-DEFINED") {
			column.ColumnType = columnType2
		} else if strings.EqualFold(column.ColumnType, "ARRAY") {
			column.ColumnType = "[]" + columnType2[1:]
		} else {
			column.ColumnType = strings.ToUpper(column.ColumnType)
		}
	case sq.DialectMySQL:
		if column.GeneratedExpr != "" {
			column.GeneratedExpr = strings.ReplaceAll(column.GeneratedExpr, `\'`, `'`)
		}
		column.ColumnType = strings.ToUpper(columnType2)
	}
}

func mapTables(catalog *Catalog, rows *sql.Rows) error {
	var tbl Table
	err := rows.Scan(
		&tbl.TableSchema,
		&tbl.TableName,
		&tbl.SQL,
	)
	if err != nil {
		return fmt.Errorf("scanning table %s: %w", tbl.TableName, err)
	}
	if catalog.Dialect == sq.DialectSQLite && tbl.SQL != "" {
		tokens, remainder := popIdentifierTokens(catalog.Dialect, tbl.SQL, 3)
		if len(tokens) == 3 &&
			strings.EqualFold(tokens[0], "CREATE") &&
			strings.EqualFold(tokens[1], "VIRTUAL") &&
			strings.EqualFold(tokens[2], "TABLE") {
			var token string
			var foundUsing bool
			for remainder != "" && !foundUsing {
				token, remainder = popIdentifierToken(catalog.Dialect, remainder)
				if strings.EqualFold(token, "USING") {
					foundUsing = true
					break
				}
			}
			if foundUsing {
				tbl.VirtualTable, remainder = popIdentifierToken(catalog.Dialect, remainder)
				i := strings.IndexByte(remainder, '(')
				j := strings.LastIndexByte(remainder, ')')
				if i >= 0 && j >= i {
					tbl.VirtualTableArgs = splitArgs(remainder[i+1 : j])
					for k, arg := range tbl.VirtualTableArgs {
						tbl.VirtualTableArgs[k] = strings.TrimSpace(arg)
					}
				}
			}
		}
	}
	// TODO: also find the check constraints
	// TODO: also find the autoincrement status
	// TODO: also find GENERATED ALWAYS columns
	var schema *Schema
	if n := catalog.CachedSchemaPosition(tbl.TableSchema); n >= 0 {
		schema = catalog.Schemas[n]
	} else {
		schema = &Schema{SchemaName: tbl.TableSchema}
		catalog.AppendSchema(schema)
	}
	schema.AppendTable(&tbl)
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
		return fmt.Errorf("scanning table %s column %s: %w", column.TableName, column.ColumnName, err)
	}
	normalizeColumn(catalog.Dialect, &column, columnType2)
	var schema *Schema
	if n := catalog.CachedSchemaPosition(column.TableSchema); n >= 0 {
		schema = catalog.Schemas[n]
	} else {
		schema = &Schema{SchemaName: column.TableSchema}
		catalog.AppendSchema(schema)
	}
	var tbl *Table
	if n := schema.CachedTablePosition(column.TableName); n >= 0 {
		tbl = schema.Tables[n]
	} else {
		tbl = &Table{
			TableSchema: column.TableSchema,
			TableName:   column.TableName,
		}
		schema.AppendTable(tbl)
	}
	tbl.AppendColumn(&column)
	return nil
}

func mapConstraints(catalog *Catalog, rows *sql.Rows) error {
	var constraint Constraint
	var rawColumns, rawExprs, rawReferencesColumns, rawOperators string
	err := rows.Scan(
		&constraint.TableSchema,
		&constraint.TableName,
		&constraint.ConstraintName,
		&constraint.ConstraintType,
		&rawColumns,
		&rawExprs,
		&constraint.ReferencesSchema,
		&constraint.ReferencesTable,
		&rawReferencesColumns,
		&constraint.UpdateRule,
		&constraint.DeleteRule,
		&constraint.MatchOption,
		&constraint.CheckExpr,
		&rawOperators,
		&constraint.IndexType,
		&constraint.Predicate,
		&constraint.IsDeferrable,
		&constraint.IsInitiallyDeferred,
	)
	if err != nil {
		return fmt.Errorf("scanning table %s constraint %s: %w", constraint.TableName, constraint.ConstraintName, err)
	}
	if rawColumns != "" {
		constraint.Columns = strings.Split(rawColumns, ",")
	}
	if rawExprs != "" {
		constraint.Exprs = strings.Split(rawExprs, ",")
	}
	if rawReferencesColumns != "" {
		constraint.ReferencesColumns = strings.Split(rawReferencesColumns, ",")
	}
	if rawOperators != "" {
		constraint.Operators = strings.Split(rawOperators, ",")
	}
	if last := len(constraint.CheckExpr) - 1; len(constraint.CheckExpr) > 2 && constraint.CheckExpr[0] == '(' && constraint.CheckExpr[last] == ')' {
		constraint.CheckExpr = constraint.CheckExpr[1:last]
	}
	var schema *Schema
	if n := catalog.CachedSchemaPosition(constraint.TableSchema); n >= 0 {
		schema = catalog.Schemas[n]
	} else {
		schema = &Schema{SchemaName: constraint.TableSchema}
		catalog.AppendSchema(schema)
	}
	if n := schema.CachedTablePosition(constraint.TableName); n >= 0 {
		schema.Tables[n].AppendConstraint(&constraint)
	}
	return nil
}

func mapIndexes(catalog *Catalog, rows *sql.Rows) error {
	var index Index
	var numKeyColumns int
	var rawColumns, rawExprs string
	err := rows.Scan(
		&index.TableSchema,
		&index.TableName,
		&index.IndexName,
		&index.IndexType,
		&index.IsUnique,
		&numKeyColumns,
		&rawColumns,
		&rawExprs,
		&index.Predicate,
		&index.SQL,
	)
	if err != nil {
		return fmt.Errorf("scanning table %s index %s: %w", index.TableName, index.IndexName, err)
	}
	if rawColumns != "" {
		index.Columns = strings.Split(rawColumns, ",")
	}
	if catalog.Dialect == sq.DialectPostgres {
		index.Columns, index.IncludeColumns = index.Columns[:numKeyColumns], index.Columns[numKeyColumns:]
	}
	index.Exprs = make([]string, len(index.Columns))
	if catalog.Dialect == sq.DialectSQLite && index.SQL != "" {
		start := strings.IndexByte(index.SQL, '(')
		end := strings.LastIndexByte(index.SQL, ')')
		if start >= 0 && end > start && end < len(index.SQL) {
			args := splitArgs(index.SQL[start+1 : end])
			for i, column := range index.Columns {
				args[i] = strings.TrimSpace(args[i])
				if column != "" {
					if i >= len(args) || args[i] != column {
						return fmt.Errorf("column mismatch: sqlite reported table %s column #%d to be %s, I got %s instead. My splitArgs function is faulty.", index.TableName, i+1, column, args[i])
					}
					continue
				}
				index.Exprs[i] = args[i]
			}
		}
		if token, remainder := popIdentifierToken(sq.DialectSQLite, index.SQL[end+1:]); strings.EqualFold(token, "WHERE") {
			index.Predicate = strings.TrimSpace(remainder)
		}
	}
	if rawExprs != "" {
		if catalog.Dialect == sq.DialectMySQL {
			rawExprs = strings.TrimSpace(strings.ReplaceAll(rawExprs, `\'`, `'`))
		}
		exprs := splitArgs(rawExprs)
		for i, column := range index.Columns {
			if column == "" && len(exprs) > 0 {
				index.Exprs[i] = "(" + strings.TrimSpace(exprs[0]) + ")"
				if catalog.Dialect == sq.DialectPostgres {
					exprs = exprs[1:]
				}
			}
			if catalog.Dialect == sq.DialectMySQL {
				exprs = exprs[1:]
			}
		}
	}
	var schema *Schema
	if n := catalog.CachedSchemaPosition(index.TableSchema); n >= 0 {
		schema = catalog.Schemas[n]
	} else {
		schema = &Schema{SchemaName: index.TableSchema}
		catalog.AppendSchema(schema)
	}
	if n := schema.CachedTablePosition(index.TableName); n >= 0 {
		schema.Tables[n].AppendIndex(&index)
	} else if n := schema.CachedViewPosition(index.TableName); n >= 0 {
		schema.Views[n].AppendIndex(&index)
	}
	return nil
}

func mapTriggers(catalog *Catalog, rows *sql.Rows) error {
	var trigger Trigger
	var actionTiming, eventManipulation string
	err := rows.Scan(
		&trigger.TableSchema,
		&trigger.TableName,
		&trigger.TriggerName,
		&trigger.SQL,
		&actionTiming,
		&eventManipulation,
	)
	if err != nil {
		return fmt.Errorf("scanning table %s trigger %s: %w", trigger.TableName, trigger.TriggerName, err)
	}
	if catalog.Dialect == sq.DialectMySQL {
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		buf.WriteString("CREATE TRIGGER " + sq.QuoteIdentifier(catalog.Dialect, trigger.TriggerName) + " " + actionTiming + " " + eventManipulation + " ON ")
		if trigger.TableSchema != "" {
			buf.WriteString(sq.QuoteIdentifier(catalog.Dialect, trigger.TableSchema) + ".")
		}
		buf.WriteString(sq.QuoteIdentifier(catalog.Dialect, trigger.TableName) + " FOR EACH ROW " + trigger.SQL)
		trigger.SQL = buf.String()
	}
	var schema *Schema
	if n := catalog.CachedSchemaPosition(trigger.TableSchema); n >= 0 {
		schema = catalog.Schemas[n]
	} else {
		schema = &Schema{SchemaName: trigger.TableSchema}
		catalog.AppendSchema(schema)
	}
	if n := schema.CachedTablePosition(trigger.TableName); n >= 0 {
		schema.Tables[n].AppendTrigger(&trigger)
	}
	return nil
}

func mapViews(catalog *Catalog, rows *sql.Rows) error {
	var view View
	err := rows.Scan(
		&view.ViewSchema,
		&view.ViewName,
		&view.IsMaterialized,
		&view.SQL,
	)
	if err != nil {
		return fmt.Errorf("scanning view %s: %w", view.ViewName, err)
	}
	view.SQL = strings.TrimSpace(view.SQL)
	if last := len(view.SQL) - 1; view.SQL[last] == ';' {
		view.SQL = view.SQL[:last]
	}
	if catalog.Dialect == sq.DialectSQLite {
		var token string
		remainder := view.SQL
		for remainder != "" {
			token, remainder = popIdentifierToken(catalog.Dialect, remainder)
			if strings.EqualFold(token, "AS") {
				break
			}
		}
		view.SQL = strings.TrimSpace(remainder)
	}
	var schema *Schema
	if n := catalog.CachedSchemaPosition(view.ViewSchema); n >= 0 {
		schema = catalog.Schemas[n]
	} else {
		schema = &Schema{SchemaName: view.ViewSchema}
		catalog.AppendSchema(schema)
	}
	schema.AppendView(&view)
	return nil
}

func introspectPostgres(ctx context.Context, db sq.DB, catalog *Catalog) error {
	err := introspectQuery(ctx, db, catalog, "sql/postgres_columns.sql", nil, mapColumns)
	if err != nil {
		return err
	}
	// NOTE: introspecting views must come before indexes and triggers so that
	// we can figure out if a index or trigger belongs to a table or
	// materialized view
	err = introspectQuery(ctx, db, catalog, "sql/postgres_views.sql", nil, mapViews)
	if err != nil {
		return err
	}
	err = introspectQuery(ctx, db, catalog, "sql/postgres_constraints.sql", nil, mapConstraints)
	if err != nil {
		return err
	}
	err = introspectQuery(ctx, db, catalog, "sql/postgres_indexes.sql", nil, mapIndexes)
	if err != nil {
		return err
	}
	err = introspectQuery(ctx, db, catalog, "sql/postgres_triggers.sql", nil, mapTriggers)
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
	err = introspectQuery(ctx, db, catalog, "sql/sqlite_constraints.sql", argslist, mapConstraints)
	if err != nil {
		return err
	}
	err = introspectQuery(ctx, db, catalog, "sql/sqlite_indexes.sql", argslist, mapIndexes)
	if err != nil {
		return err
	}
	err = introspectQuery(ctx, db, catalog, "sql/sqlite_triggers.sql", nil, mapTriggers)
	if err != nil {
		return err
	}
	err = introspectQuery(ctx, db, catalog, "sql/sqlite_views.sql", nil, mapViews)
	if err != nil {
		return err
	}
	for _, schema := range catalog.Schemas {
		for _, tbl := range schema.Tables {
			tbl.RefreshColumnCache()
			for _, constraint := range tbl.Constraints {
				if len(constraint.Columns) != 1 {
					continue
				}
				n := tbl.CachedColumnPosition(constraint.Columns[0])
				if n < 0 {
					continue
				}
				switch constraint.ConstraintType {
				case PRIMARY_KEY:
					tbl.Columns[n].IsPrimaryKey = true
				case UNIQUE:
					tbl.Columns[n].IsUnique = true
				}
			}
		}
	}
	return nil
}

func introspectMySQL(ctx context.Context, db sq.DB, catalog *Catalog) error {
	err := introspectQuery(ctx, db, catalog, "sql/mysql_columns.sql", nil, mapColumns)
	if err != nil {
		return err
	}
	err = introspectQuery(ctx, db, catalog, "sql/mysql_constraints.sql", nil, mapConstraints)
	if err != nil {
		return err
	}
	err = introspectQuery(ctx, db, catalog, "sql/mysql_indexes.sql", nil, mapIndexes)
	if err != nil {
		return err
	}
	err = introspectQuery(ctx, db, catalog, "sql/mysql_triggers.sql", nil, mapTriggers)
	if err != nil {
		return err
	}
	err = introspectQuery(ctx, db, catalog, "sql/mysql_views.sql", nil, mapViews)
	if err != nil {
		return err
	}
	return nil
}
