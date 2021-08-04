package ddl2

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"text/template"

	"github.com/bokwoon95/sq"
)

type IntrospectSettings struct {
	IncludeSystemObjects bool
	WithSchemas          []string
	WithoutSchemas       []string
	WithTables           []string
	WithoutTables        []string
}

type Introspector interface {
	// SELECT sqlite_version();
	// SELECT split_part(current_setting('server_version'), ' ', 1);
	// SELECT version();
	GetVersion(context.Context, *IntrospectSettings) (version string, err error)
	GetTables(context.Context, *IntrospectSettings) ([]Table, error)
	GetColumns(context.Context, *IntrospectSettings) ([]Column, error)
	GetConstraints(context.Context, *IntrospectSettings) ([]Constraint, error)
	GetIndexes(context.Context, *IntrospectSettings) ([]Index, error)
	GetTriggers(context.Context, *IntrospectSettings) ([]Trigger, error)
	GetViews(context.Context, *IntrospectSettings) ([]View, error)
	GetDefaultSchema(context.Context) (defaultSchema string, err error)
	GetSchemas(context.Context, *IntrospectSettings) ([]Schema, error)
	GetExtensions(context.Context, *IntrospectSettings) (extensions [][2]string, err error)
	GetFunctions(context.Context, *IntrospectSettings) ([]Function, error)
}

type DatabaseIntrospector struct {
	dialect         string
	db              sq.DB
	mu              *sync.RWMutex
	defaultSettings *IntrospectSettings
	templates       map[string]*template.Template
}

func NewDatabaseIntrospector(dialect string, db sq.DB, defaultSettings *IntrospectSettings) (*DatabaseIntrospector, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	if dialect != sq.DialectSQLite && dialect != sq.DialectPostgres && dialect != sq.DialectMySQL {
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
	_, err := db.ExecContext(context.Background(), "SELECT 1")
	if err != nil {
		return nil, fmt.Errorf("liveliness check failed: %w", err)
	}
	dbi := &DatabaseIntrospector{
		dialect:         dialect,
		db:              db,
		mu:              &sync.RWMutex{},
		defaultSettings: defaultSettings,
		templates:       make(map[string]*template.Template),
	}
	return dbi, nil
}

func (dbi *DatabaseIntrospector) queryContext(ctx context.Context, fsys fs.FS, name string, settings *IntrospectSettings) (*sql.Rows, error) {
	var err error
	dbi.mu.RLock()
	tmpl := dbi.templates[name]
	dbi.mu.RUnlock()
	if tmpl == nil {
		tmpl, err = template.ParseFS(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", name, err)
		}
		dbi.mu.Lock()
		dbi.templates[name] = tmpl
		dbi.mu.Unlock()
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	if settings == nil {
		settings = dbi.defaultSettings
	}
	if settings == nil {
		settings = &IntrospectSettings{}
	}
	err = tmpl.Execute(buf, *settings)
	if err != nil {
		return nil, fmt.Errorf("executing %s: %w", name, err)
	}
	return dbi.db.QueryContext(ctx, buf.String())
}

func (dbi *DatabaseIntrospector) GetTables(ctx context.Context, settings *IntrospectSettings) ([]Table, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.queryContext(ctx, sqlDir, "sqlite_tables.sql", settings)
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, sqlDir, "postgres_tables.sql", settings)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.queryContext(ctx, sqlDir, "mysql_tables.sql", settings)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dbi.dialect)
	}
	var tbls []Table
	for rows.Next() {
		var tbl Table
		switch dbi.dialect {
		case sq.DialectSQLite:
			err = rows.Scan(&tbl.TableName, &tbl.SQL)
			if err != nil {
				return nil, fmt.Errorf("scanning Table: %w", err)
			}
		case sq.DialectPostgres, sq.DialectMySQL:
			err = rows.Scan(&tbl.TableSchema, &tbl.TableName)
			if err != nil {
				return nil, fmt.Errorf("scanning Table: %w", err)
			}
		}
		tbls = append(tbls, tbl)
	}
	return tbls, nil
}

func (dbi *DatabaseIntrospector) GetColumns(ctx context.Context, settings *IntrospectSettings) ([]Column, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.queryContext(ctx, sqlDir, "sqlite_tables.sql", settings)
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, sqlDir, "postgres_tables.sql", settings)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.queryContext(ctx, sqlDir, "mysql_tables.sql", settings)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dbi.dialect)
	}
	defer rows.Close()
	var columns []Column
	for rows.Next() {
		var column Column
		var columnType2 string
		switch dbi.dialect {
		case sq.DialectSQLite:
			err = rows.Scan(
				&column.TableName,
				&column.ColumnName,
				&column.ColumnType,
				&column.IsNotNull,
				&column.ColumnDefault,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Column: %w", err)
			}
		case sq.DialectPostgres:
			err = rows.Scan(
				&column.TableSchema,
				&column.TableName,
				&column.ColumnName,
				&column.ColumnType,
				&columnType2,
				&column.NumericPrecision,
				&column.NumericScale,
				&column.Identity,
				&column.IsNotNull,
				&column.GeneratedExpr,
				&column.GeneratedExprStored,
				&column.CollationName,
				&column.ColumnDefault,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Column: %w", err)
			}
		case sq.DialectMySQL:
			err = rows.Scan(
				&column.TableSchema,
				&column.TableName,
				&column.ColumnName,
				&column.ColumnType,
				&columnType2,
				&column.NumericPrecision,
				&column.NumericScale,
				&column.IsAutoincrement,
				&column.IsNotNull,
				&column.OnUpdateCurrentTimestamp,
				&column.GeneratedExpr,
				&column.GeneratedExprStored,
				&column.CollationName,
				&column.ColumnDefault,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Column: %w", err)
			}
		}
		normalizeColumn(dbi.dialect, &column, columnType2)
		columns = append(columns, column)
	}
	err = rows.Close()
	if err != nil {
		return nil, fmt.Errorf("rows.Close: %w", err)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return columns, nil
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

func (dbi *DatabaseIntrospector) GetConstraints(ctx context.Context, settings *IntrospectSettings) ([]Constraint, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.queryContext(ctx, sqlDir, "sqlite_constraints.sql", settings)
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, sqlDir, "postgres_constraints.sql", settings)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.queryContext(ctx, sqlDir, "mysql_constraints.sql", settings)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dbi.dialect)
	}
	defer rows.Close()
	var constraints []Constraint
	for rows.Next() {
		var constraint Constraint
		var rawColumns, rawExprs, rawReferencesColumns, rawOperators string
		switch dbi.dialect {
		case sq.DialectSQLite:
			err = rows.Scan(
				&constraint.TableName,
				&constraint.ConstraintType,
				&rawColumns,
				&constraint.ReferencesTable,
				&rawReferencesColumns,
				&constraint.UpdateRule,
				&constraint.DeleteRule,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Constraint: %w", err)
			}
		case sq.DialectPostgres:
			err = rows.Scan(
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
				return nil, fmt.Errorf("scanning Constraint: %w", err)
			}
		case sq.DialectMySQL:
			err = rows.Scan(
				&constraint.TableSchema,
				&constraint.TableName,
				&constraint.ConstraintName,
				&constraint.ConstraintType,
				&rawColumns,
				&constraint.ReferencesSchema,
				&constraint.ReferencesTable,
				&rawReferencesColumns,
				&constraint.UpdateRule,
				&constraint.DeleteRule,
				&constraint.MatchOption,
				&constraint.CheckExpr,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Constraint: %w", err)
			}
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
			constraint.ExclusionOperators = strings.Split(rawOperators, ",")
		}
		if last := len(constraint.CheckExpr) - 1; len(constraint.CheckExpr) > 2 && constraint.CheckExpr[0] == '(' && constraint.CheckExpr[last] == ')' {
			constraint.CheckExpr = constraint.CheckExpr[1:last]
		}
		constraints = append(constraints, constraint)
	}
	err = rows.Close()
	if err != nil {
		return nil, fmt.Errorf("rows.Close: %w", err)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return constraints, nil
}
