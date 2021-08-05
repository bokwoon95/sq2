package ddl2

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/bokwoon95/sq"
)

var ErrUnsupportedFeature = errors.New("dialect does not support this feature")

type IntrospectSettings struct {
	IncludeSystemObjects bool
	WithSchemas          []string
	WithoutSchemas       []string
	WithTables           []string
	WithoutTables        []string
}

type Introspector interface {
	GetVersionNums(context.Context) (versionNums []int, err error)
	GetCatalogName(context.Context) (catalogName string, err error)
	GetCurrentSchema(context.Context) (currentSchema string, err error)
	GetExtensions(context.Context) (extensions []string, err error)
	GetTables(context.Context, *IntrospectSettings) ([]Table, error)
	GetColumns(context.Context, *IntrospectSettings) ([]Column, error)
	GetConstraints(context.Context, *IntrospectSettings) ([]Constraint, error)
	GetIndexes(context.Context, *IntrospectSettings) ([]Index, error)
	GetTriggers(context.Context, *IntrospectSettings) ([]Trigger, error)
	GetViews(context.Context, *IntrospectSettings) ([]View, error)
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

func (dbi *DatabaseIntrospector) GetVersionNums(ctx context.Context) (versionNums []int, err error) {
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.db.QueryContext(ctx, "SELECT sqlite_version()")
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.db.QueryContext(ctx, "SELECT current_settings('server_version')")
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.db.QueryContext(ctx, "SELECT version()")
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dbi.dialect)
	}
	defer rows.Close()
	var versionString string
	for rows.Next() {
		err = rows.Scan(&versionString)
		if err != nil {
			return nil, fmt.Errorf("scanning versionString: %w", err)
		}
		switch dbi.dialect {
		case sq.DialectPostgres:
			if i := strings.IndexByte(versionString, ' '); i >= 0 {
				versionString = versionString[:i]
			}
		}
		break
	}
	rawVersionNums := strings.Split(versionString, ".")
	versionNums = make([]int, len(rawVersionNums))
	for i, rawVersionNum := range rawVersionNums {
		versionNum, err := strconv.Atoi(rawVersionNum)
		if err != nil {
			return versionNums, fmt.Errorf("version %s: cannot convert %s to integer: %w", versionString, rawVersionNum, err)
		}
		versionNums[i] = versionNum
	}
	err = rows.Close()
	if err != nil {
		return nil, fmt.Errorf("rows.Close: %w", err)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return versionNums, nil
}

func (dbi *DatabaseIntrospector) GetCatalogName(ctx context.Context) (catalogName string, err error) {
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		return "", nil
	case sq.DialectPostgres:
		rows, err = dbi.db.QueryContext(ctx, "SELECT current_database()")
		if err != nil {
			return "", err
		}
	case sq.DialectMySQL:
		rows, err = dbi.db.QueryContext(ctx, "SELECT database()")
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported dialect: %s", dbi.dialect)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&catalogName)
		if err != nil {
			return "", fmt.Errorf("scanning catalogName: %w", err)
		}
		break
	}
	err = rows.Close()
	if err != nil {
		return "", fmt.Errorf("rows.Close: %w", err)
	}
	err = rows.Err()
	if err != nil {
		return "", fmt.Errorf("rows.Err: %w", err)
	}
	return catalogName, nil
}

func (dbi *DatabaseIntrospector) GetCurrentSchema(ctx context.Context) (currentSchema string, err error) {
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		return "", nil
	case sq.DialectPostgres:
		rows, err = dbi.db.QueryContext(ctx, "SELECT current_schema()")
		if err != nil {
			return "", err
		}
	case sq.DialectMySQL:
		rows, err = dbi.db.QueryContext(ctx, "SELECT database()")
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported dialect: %s", dbi.dialect)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&currentSchema)
		if err != nil {
			return "", fmt.Errorf("scanning currentSchema: %w", err)
		}
		break
	}
	err = rows.Close()
	if err != nil {
		return "", fmt.Errorf("rows.Close: %w", err)
	}
	err = rows.Err()
	if err != nil {
		return "", fmt.Errorf("rows.Err: %w", err)
	}
	return currentSchema, nil
}

func (dbi *DatabaseIntrospector) GetExtensions(ctx context.Context) (extensions []string, err error) {
	if dbi.dialect != sq.DialectPostgres {
		return nil, fmt.Errorf("%w dialect=%s, feature=extensions", ErrUnsupportedFeature, dbi.dialect)
	}
	rows, err := dbi.db.QueryContext(ctx, "SELECT extname, extversion FROM pg_catalog.pg_extension")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var extname, extversion string
		err = rows.Scan(&extname, &extversion)
		if err != nil {
			return nil, fmt.Errorf("scanning extension: %w", err)
		}
		extensions = append(extensions, extname+"@"+extversion)
	}
	return extensions, nil
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
	defer rows.Close()
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
	err = rows.Close()
	if err != nil {
		return nil, fmt.Errorf("rows.Close: %w", err)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
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
	var columnType2 string
	for rows.Next() {
		var column Column
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
			if strings.EqualFold(column.ColumnType, "USER-DEFINED") {
				column.ColumnType = columnType2
			} else if strings.EqualFold(column.ColumnType, "ARRAY") {
				column.ColumnType = "[]" + columnType2[1:]
			} else if (strings.EqualFold(column.ColumnType, "NUMERIC") || strings.EqualFold(column.ColumnType, "DECIMAL")) && (column.NumericPrecision > 0 || column.NumericScale > 0) {
				column.ColumnType = fmt.Sprintf("%s(%d,%d)", strings.ToUpper(column.ColumnType), column.NumericPrecision, column.NumericScale)
			} else {
				column.ColumnType = strings.ToUpper(column.ColumnType)
			}
			// remove surrounding brackets
			if len(column.GeneratedExpr) > 2 {
				last := len(column.GeneratedExpr) - 1
				if column.GeneratedExpr[0] == '(' && column.GeneratedExpr[last] == ')' {
					column.GeneratedExpr = column.GeneratedExpr[1:last]
				}
			}
			if strings.EqualFold(column.Identity, "BY DEFAULT") {
				column.Identity = BY_DEFAULT_AS_IDENTITY
			} else if strings.EqualFold(column.Identity, "ALWAYS") {
				column.Identity = ALWAYS_AS_IDENTITY
			}
		case sq.DialectMySQL:
			err = rows.Scan(
				&column.TableSchema,
				&column.TableName,
				&column.ColumnName,
				&column.ColumnType,
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
			column.ColumnType = strings.ToUpper(column.ColumnType)
			if column.GeneratedExpr != "" {
				column.GeneratedExpr = strings.ReplaceAll(column.GeneratedExpr, `\'`, `'`)
			}
		}
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
	var rawColumns, rawExprs, rawReferencesColumns, rawOperators string
	for rows.Next() {
		var constraint Constraint
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
			if constraint.CheckExpr != "" {
				constraint.CheckExpr = strings.TrimPrefix(constraint.CheckExpr, "CHECK ")
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
		// remove surrounding brackets
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

func (dbi *DatabaseIntrospector) GetIndexes(ctx context.Context, settings *IntrospectSettings) ([]Index, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.queryContext(ctx, sqlDir, "sqlite_indexes.sql", settings)
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, sqlDir, "postgres_indexes.sql", settings)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.queryContext(ctx, sqlDir, "mysql_indexes.sql", settings)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dbi.dialect)
	}
	defer rows.Close()
	var indexes []Index
	var numKeyColumns int
	var rawColumns, rawExprs string
	for rows.Next() {
		var index Index
		switch dbi.dialect {
		case sq.DialectSQLite:
			err = rows.Scan(
				&index.TableName,
				&index.IndexName,
				&index.IsUnique,
				&rawColumns,
				&index.SQL,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Index: %w", err)
			}
			if rawColumns != "" {
				index.Columns = strings.Split(rawColumns, ",")
			}
			index.Exprs = make([]string, len(index.Columns))
			if index.SQL != "" {
				start := strings.IndexByte(index.SQL, '(')
				end := strings.LastIndexByte(index.SQL, ')')
				if start >= 0 && end > start && end < len(index.SQL) {
					args := splitArgs(index.SQL[start+1 : end])
					for i, column := range index.Columns {
						args[i] = strings.TrimSpace(args[i])
						if column != "" {
							if i >= len(args) || args[i] != column {
								return nil, fmt.Errorf("column mismatch: sqlite reported table %s column #%d to be %s, I got %s instead. This means the splitArgs function is faulty and must be escalated", index.TableName, i+1, column, args[i])
							}
							continue
						}
						index.Exprs[i] = args[i]
					}
				}
				if token, remainder, _ := popIdentifierToken(sq.DialectSQLite, index.SQL[end+1:]); strings.EqualFold(token, "WHERE") {
					index.Predicate = strings.TrimSpace(remainder)
				}
			}
		case sq.DialectPostgres:
			err = rows.Scan(
				&index.TableSchema,
				&index.TableName,
				&index.IndexName,
				&index.IndexType,
				&index.IsUnique,
				&numKeyColumns,
				&rawColumns,
				&rawExprs,
				&index.Predicate,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Index: %w", err)
			}
			if rawColumns != "" {
				index.Columns = strings.Split(rawColumns, ",")
			}
			index.Exprs = make([]string, len(index.Columns))
			exprs := splitArgs(rawExprs)
			for i, column := range index.Columns {
				if column == "" && len(exprs) > 0 {
					index.Exprs[i] = "(" + strings.TrimSpace(exprs[0]) + ")"
					exprs = exprs[1:]
				}
			}
			index.Columns, index.IncludeColumns = index.Columns[:numKeyColumns], index.Columns[numKeyColumns:]
		case sq.DialectMySQL:
			err = rows.Scan(
				&index.TableSchema,
				&index.TableName,
				&index.IndexName,
				&index.IndexType,
				&index.IsUnique,
				&rawColumns,
				&rawExprs,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Index: %w", err)
			}
			if rawColumns != "" {
				index.Columns = strings.Split(rawColumns, ",")
			}
			index.Exprs = make([]string, len(index.Columns))
			exprs := splitArgs(strings.TrimSpace(strings.ReplaceAll(rawExprs, `\'`, `'`)))
			for i, column := range index.Columns {
				if column == "" && len(exprs) > 0 {
					index.Exprs[i] = "(" + strings.TrimSpace(exprs[0]) + ")"
				}
				exprs = exprs[1:]
			}
		}
		indexes = append(indexes, index)
	}
	err = rows.Close()
	if err != nil {
		return nil, fmt.Errorf("rows.Close: %w", err)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return indexes, nil
}

func (dbi *DatabaseIntrospector) GetTriggers(ctx context.Context, settings *IntrospectSettings) ([]Trigger, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.queryContext(ctx, sqlDir, "sqlite_triggers.sql", settings)
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, sqlDir, "postgres_triggers.sql", settings)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.queryContext(ctx, sqlDir, "mysql_triggers.sql", settings)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dbi.dialect)
	}
	defer rows.Close()
	var triggers []Trigger
	var actionTiming, eventManipulation string
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	for rows.Next() {
		var trigger Trigger
		switch dbi.dialect {
		case sq.DialectSQLite:
			err = rows.Scan(
				&trigger.TableName,
				&trigger.TriggerName,
				&trigger.SQL,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Trigger: %w", err)
			}
		case sq.DialectPostgres:
			err = rows.Scan(
				&trigger.TableSchema,
				&trigger.TableName,
				&trigger.TriggerName,
				&trigger.SQL,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Trigger: %w", err)
			}
		case sq.DialectMySQL:
			err = rows.Scan(
				&trigger.TableSchema,
				&trigger.TableName,
				&trigger.TriggerName,
				&trigger.SQL,
				&actionTiming,
				&eventManipulation,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Trigger: %w", err)
			}
			buf.Reset()
			buf.WriteString("CREATE TRIGGER " + sq.QuoteIdentifier(dbi.dialect, trigger.TriggerName) + " " + actionTiming + " " + eventManipulation + " ON ")
			if trigger.TableSchema != "" {
				buf.WriteString(sq.QuoteIdentifier(dbi.dialect, trigger.TableSchema) + ".")
			}
			buf.WriteString(sq.QuoteIdentifier(dbi.dialect, trigger.TableName) + " FOR EACH ROW " + trigger.SQL)
			trigger.SQL = buf.String()
		}
		triggers = append(triggers, trigger)
	}
	err = rows.Close()
	if err != nil {
		return nil, fmt.Errorf("rows.Close: %w", err)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return triggers, nil
}

func (dbi *DatabaseIntrospector) GetViews(ctx context.Context, settings *IntrospectSettings) ([]View, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.queryContext(ctx, sqlDir, "sqlite_views.sql", settings)
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, sqlDir, "postgres_views.sql", settings)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.queryContext(ctx, sqlDir, "mysql_views.sql", settings)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dbi.dialect)
	}
	defer rows.Close()
	var views []View
	for rows.Next() {
		var view View
		switch dbi.dialect {
		case sq.DialectSQLite:
			err = rows.Scan(
				&view.ViewName,
				&view.SQL,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning View: %w", err)
			}
		case sq.DialectPostgres:
			err = rows.Scan(
				&view.ViewSchema,
				&view.ViewName,
				&view.IsMaterialized,
				&view.SQL,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning View: %w", err)
			}
		case sq.DialectMySQL:
			err = rows.Scan(
				&view.ViewSchema,
				&view.ViewName,
				&view.SQL,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning View: %w", err)
			}
		}
		views = append(views, view)
	}
	err = rows.Close()
	if err != nil {
		return nil, fmt.Errorf("rows.Close: %w", err)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return views, nil
}
