package sq

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strconv"
	"sync"
)

var ErrNonFetchableQuery = errors.New("query does not support fetchable fields")

var (
	bufpool  = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
	argspool = sync.Pool{New: func() interface{} { return make([]interface{}, 0) }}
)

const (
	DialectSQLite    = "sqlite"
	DialectPostgres  = "postgres"
	DialectMySQL     = "mysql"
	DialectSQLServer = "sqlserver"
	// DialectOracle   = "Oracle"
)

type (
	SQLiteDialect   struct{}
	PostgresDialect struct{}
	MySQLDialect    struct{}
	// MSSQLDialect    struct{}
	// OracleDialect   struct{}
)

var (
	SQLite   = SQLiteDialect{}
	Postgres = PostgresDialect{}
	MySQL    = MySQLDialect{}
	// MSSQL    = MSSQLDialect{}
	// Oracle   = OracleDialect{}
)

type SelectType string

const (
	SelectTypeDefault    SelectType = "SELECT"
	SelectTypeDistinct   SelectType = "SELECT DISTINCT"
	SelectTypeDistinctOn SelectType = "SELECT DISTINCT ON"
)

type SQLAppender interface {
	AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error
}

type SQLExcludeAppender interface {
	// Fields should respect the excludedTableQualifiers argument in ToSQL().
	// E.g. if the field 'name' belongs to a table called 'users' and the
	// excludedTableQualifiers contains 'users', the field should present itself
	// as 'name' and not 'users.name'. i.e. any table qualifiers in the list
	// must be excluded.
	//
	// This is to play nice with certain clauses in the INSERT and UPDATE
	// queries that expressly forbid table qualified columns.
	AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error
}

type Table interface {
	SQLAppender
	GetAlias() string
	GetName() string // Table name must exclude the schema (if any)
}

type BaseTable interface {
	Table
	GetSchema() string
}

type Field interface {
	SQLExcludeAppender
	GetAlias() string
	GetName() string // Field name must exclude the table name
}

type Query interface {
	SQLAppender
	SetFetchableFields([]Field) (Query, error)
	GetFetchableFields() ([]Field, error)
	Dialect() string
}

func ToSQL(dialect string, q Query) (query string, args []interface{}, params map[string][]int, err error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	params = make(map[string][]int)
	if dialect == "" {
		dialect = q.Dialect()
	}
	err = q.AppendSQL(dialect, buf, &args, params)
	return buf.String(), args, params, err
}

type Predicate interface {
	Field
	Not() Predicate
}

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type Transactor interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

func isExplodableSlice(value interface{}) bool {
	valueType := reflect.TypeOf(value)
	if valueType == nil {
		return false
	}
	return valueType.Kind() == reflect.Slice && valueType.Elem().Kind() != reflect.Uint8
}

func explodeSlice(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string, value interface{}) error {
	slice := reflect.ValueOf(value)
	length := slice.Len()
	var err error
	for i := 0; i < length; i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		v := slice.Index(i).Interface()
		if v, ok := v.(SQLExcludeAppender); ok && v != nil {
			err = v.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
			if err != nil {
				return err
			}
			continue
		}
		if v, ok := v.(SQLAppender); ok && v != nil {
			err = v.AppendSQL(dialect, buf, args, params)
			if err != nil {
				return err
			}
			continue
		}
		switch dialect {
		case DialectPostgres, DialectSQLite:
			buf.WriteString("$" + strconv.Itoa(len(*args)+1))
		case DialectSQLServer:
			buf.WriteString("@p" + strconv.Itoa(len(*args)+1))
		default:
			buf.WriteString("?")
		}
		*args = append(*args, v)
	}
	return nil
}

func QuoteIdentifier(dialect string, identifier string) string {
	// TODO: think about doing away this function entirely. People shouldn't be
	// using non-standaed identifier names at all. All it's doing is slowing
	// down the happy path, which is that identifier names aren't quoted.  The
	// other alternative is to always quote all identifiers.
	var needsQuoting bool
	// TODO: Run each loop iteration in parallel. Wait for the first "success"
	// (finding a character that warrants quoting), then terminate the rest of
	// the goroutines. Else if all goroutines exit without any reporting, then
	// the identifier doesn't have to be quoted.
	for i, char := range identifier {
		if i == 0 && (char >= '0' && char <= '9') {
			// first character cannot be a number
			needsQuoting = true
			break
		}
		switch {
		case char == '_',
			char >= '0' && char <= '9',
			char >= 'a' && char <= 'z':
			continue
		case char >= 'A' && char <= 'Z':
			// If there are capital letters, the identifier is quoted to
			// preserve capitalization information (because databases treat
			// capital letters differently based on their dialect or
			// configuration)
			fallthrough
		default:
			// In general there may be some other characters that are allowed
			// in unquoted identifiers (e.g. '$'), but different databases
			// allow different things. We only recognize a-z0-9 as the true
			// standard.
			needsQuoting = true
			break
		}
	}
	if !needsQuoting {
		return identifier
	}
	switch dialect {
	case DialectMySQL:
		return "`" + EscapeQuote(identifier, '`') + "`"
	case DialectSQLServer:
		return "[" + EscapeQuote(identifier, '[') + "]"
	default:
		return `"` + EscapeQuote(identifier, '"') + `"`
	}
}
