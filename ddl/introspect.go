package ddl

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

type Filter struct {
	IncludeSystemCatalogs bool
	IncludeComments       bool
	SortOutput            bool
	WithSchemas           []string
	WithoutSchemas        []string
	WithTables            []string
	WithoutTables         []string
	WithFunctions         []string
	WithoutFunctions      []string
}

type Introspector interface {
	GetVersionNums(context.Context) (versionNums []int, err error)
	GetDatabaseName(context.Context) (databaseName string, err error)
	GetCurrentSchema(context.Context) (currentSchema string, err error)
	GetExtensions(context.Context, *Filter) (extensions []string, err error)
	GetTables(context.Context, *Filter) ([]Table, error)
	GetColumns(context.Context, *Filter) ([]Column, error)
	GetConstraints(context.Context, *Filter) ([]Constraint, error)
	GetIndexes(context.Context, *Filter) ([]Index, error)
	GetTriggers(context.Context, *Filter) ([]Trigger, error)
	GetViews(context.Context, *Filter) ([]View, error)
	GetFunctions(context.Context, *Filter) ([]Function, error)
}

type DatabaseIntrospector struct {
	dialect       string
	db            sq.DB
	defaultFilter *Filter
	mu            *sync.RWMutex
	templates     map[string]*template.Template
}

func NewDatabaseIntrospector(dialect string, db sq.DB, defaultFilter *Filter) (*DatabaseIntrospector, error) {
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
		dialect:       dialect,
		db:            db,
		mu:            &sync.RWMutex{},
		defaultFilter: defaultFilter,
		templates:     make(map[string]*template.Template),
	}
	return dbi, nil
}

func printList(strs []string) string {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	for i, str := range strs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString("'" + sq.EscapeQuote(str, '\'') + "'")
	}
	return buf.String()
}

func (dbi *DatabaseIntrospector) queryContext(ctx context.Context, fsys fs.FS, name string, filter *Filter) (*sql.Rows, error) {
	var err error
	dbi.mu.RLock()
	tmpl := dbi.templates[name]
	dbi.mu.RUnlock()
	if tmpl == nil {
		b, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", name, err)
		}
		tmpl, err = template.
			New(name).
			Funcs(template.FuncMap{"printList": printList}).
			Parse(string(b))
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
	if filter == nil {
		if dbi.defaultFilter == nil {
			filter = &Filter{}
		} else {
			filter = dbi.defaultFilter
		}
	}
	err = tmpl.Execute(buf, *filter)
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
		// rows, err = dbi.db.QueryContext(ctx, "SELECT current_settings('server_version')")
		rows, err = dbi.db.QueryContext(ctx, "SHOW server_version")
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

func (dbi *DatabaseIntrospector) GetDatabaseName(ctx context.Context) (databaseName string, err error) {
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.db.QueryContext(ctx, "SELECT file FROM pragma_database_list WHERE file = 'main'")
		if err != nil {
			return "", err
		}
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
		err = rows.Scan(&databaseName)
		if err != nil {
			return "", fmt.Errorf("scanning databaseName: %w", err)
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
	return databaseName, nil
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

func (dbi *DatabaseIntrospector) GetExtensions(ctx context.Context, filter *Filter) (extensions []string, err error) {
	if dbi.dialect != sq.DialectPostgres {
		return nil, fmt.Errorf("%w dialect=%s, feature=extensions", ErrUnsupportedFeature, dbi.dialect)
	}
	rows, err := dbi.queryContext(ctx, embeddedFiles, "sql/postgres_extensions.sql", filter)
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

func (dbi *DatabaseIntrospector) GetTables(ctx context.Context, filter *Filter) ([]Table, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/sqlite_tables.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/postgres_tables.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/mysql_tables.sql", filter)
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

func (dbi *DatabaseIntrospector) GetColumns(ctx context.Context, filter *Filter) ([]Column, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/sqlite_columns.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/postgres_columns.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/mysql_columns.sql", filter)
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
			column.ColumnType = strings.TrimSuffix(column.ColumnType, " GENERATED ALWAYS")
			if column.ColumnDefault != "" {
				column.ColumnDefault = toExpr(dbi.dialect, column.ColumnDefault)
			}
		case sq.DialectPostgres:
			err = rows.Scan(
				&column.TableSchema,
				&column.TableName,
				&column.ColumnName,
				&column.ColumnType,
				&column.NumericPrecision,
				&column.NumericScale,
				&column.Identity,
				&column.IsNotNull,
				&column.GeneratedExpr,
				&column.GeneratedExprStored,
				&column.CollationName,
				&column.ColumnDefault,
				&column.ColumnComment,
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
				&column.NumericPrecision,
				&column.NumericScale,
				&column.IsAutoincrement,
				&column.IsNotNull,
				&column.OnUpdateCurrentTimestamp,
				&column.GeneratedExpr,
				&column.GeneratedExprStored,
				&column.CollationName,
				&column.ColumnDefault,
				&column.ColumnComment,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Column: %w", err)
			}
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

func (dbi *DatabaseIntrospector) GetConstraints(ctx context.Context, filter *Filter) ([]Constraint, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/sqlite_constraints.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/postgres_constraints.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/mysql_constraints.sql", filter)
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
				&constraint.ExclusionIndex,
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
		constraint.Exprs = make([]string, len(constraint.Columns))
		if dbi.dialect == sq.DialectPostgres {
			exprs := splitArgs(rawExprs)
			for i, column := range constraint.Columns {
				if column == "" && len(exprs) > 0 {
					constraint.Exprs[i] = strings.TrimSpace(exprs[0])
					exprs = exprs[1:]
				}
			}
			if rawOperators != "" {
				constraint.ExclusionOperators = strings.Split(rawOperators, ",")
			}
		}
		if rawReferencesColumns != "" {
			constraint.ReferencesColumns = strings.Split(rawReferencesColumns, ",")
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

func (dbi *DatabaseIntrospector) GetIndexes(ctx context.Context, filter *Filter) ([]Index, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/sqlite_indexes.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/postgres_indexes.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/mysql_indexes.sql", filter)
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

func (dbi *DatabaseIntrospector) GetTriggers(ctx context.Context, filter *Filter) ([]Trigger, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/sqlite_triggers.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/postgres_triggers.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/mysql_triggers.sql", filter)
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

func (dbi *DatabaseIntrospector) GetFunctions(ctx context.Context, filter *Filter) ([]Function, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		return nil, fmt.Errorf("{%w} dialect=sqlite feature=functions", ErrUnsupportedFeature)
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/postgres_functions.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		return nil, fmt.Errorf("(*DatabaseIntrospector).GetFunctions() has not yet been implemented for MySQL (TODO)")
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dbi.dialect)
	}
	defer rows.Close()
	var functions []Function
	for rows.Next() {
		var function Function
		switch dbi.dialect {
		case sq.DialectPostgres:
			err = rows.Scan(
				&function.FunctionSchema,
				&function.FunctionName,
				&function.SQL,
				&function.RawArgs,
				&function.ReturnType,
			)
			if err != nil {
				return nil, fmt.Errorf("scanning Function: %w", err)
			}
		}
		functions = append(functions, function)
	}
	err = rows.Close()
	if err != nil {
		return nil, fmt.Errorf("rows.Close: %w", err)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return functions, nil
}

func (dbi *DatabaseIntrospector) GetViews(ctx context.Context, filter *Filter) ([]View, error) {
	var err error
	var rows *sql.Rows
	switch dbi.dialect {
	case sq.DialectSQLite:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/sqlite_views.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectPostgres:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/postgres_views.sql", filter)
		if err != nil {
			return nil, err
		}
	case sq.DialectMySQL:
		rows, err = dbi.queryContext(ctx, embeddedFiles, "sql/mysql_views.sql", filter)
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
			if i := strings.Index(view.SQL, " AS "); i >= 0 {
				view.SQL = view.SQL[i+4:]
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
