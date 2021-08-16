package sq

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"reflect"
	"runtime"
	"strconv"
	"strings"
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
	DialectOracle    = "oracle"
)

type (
	SQLiteQueryBuilder    struct{ env map[string]interface{} }
	PostgresQueryBuilder  struct{ env map[string]interface{} }
	MySQLQueryBuilder     struct{ env map[string]interface{} }
	SQLServerQueryBuilder struct{ env map[string]interface{} }
	OracleQueryBuilder    struct{ env map[string]interface{} }
)

var (
	SQLite    = SQLiteQueryBuilder{}
	Postgres  = PostgresQueryBuilder{}
	MySQL     = MySQLQueryBuilder{}
	SQLServer = SQLServerQueryBuilder{}
	Oracle    = OracleQueryBuilder{}
)

func SQLiteEnv(env map[string]interface{}) SQLiteQueryBuilder {
	return SQLiteQueryBuilder{env: env}
}

func PostgresEnv(env map[string]interface{}) PostgresQueryBuilder {
	return PostgresQueryBuilder{env: env}
}

func MySQLEnv(env map[string]interface{}) MySQLQueryBuilder {
	return MySQLQueryBuilder{env: env}
}

func SQLServerEnv(env map[string]interface{}) SQLServerQueryBuilder {
	return SQLServerQueryBuilder{env: env}
}

func OracleEnv(env map[string]interface{}) OracleQueryBuilder {
	return OracleQueryBuilder{env: env}
}

type SQLAppender interface {
	AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error
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
	//
	// excludedTableQualifiers must be sorted in ascending order, as the fields
	// will rely on binary search to find a match.
	AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error
}

type Table interface {
	SQLAppender
	GetAlias() string
	GetName() string // Table name must exclude the schema (if any)
}

type PredicateHook interface {
	InjectPredicate(env map[string]interface{}) (Predicate, error)
}

type SchemaTable interface {
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
	GetDialect() string
}

func ToSQL(dialect string, q SQLAppender) (query string, args []interface{}, params map[string][]int, err error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	params = make(map[string][]int)
	if dialect == "" {
		if q, ok := q.(Query); ok {
			dialect = q.GetDialect()
		}
	}
	err = q.AppendSQL(dialect, buf, &args, params, nil)
	return buf.String(), args, params, err
}

func ToSQLExclude(dialect string, f SQLExcludeAppender, excludedTableQualifiers []string) (query string, args []interface{}, params map[string][]int, err error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	params = make(map[string][]int)
	err = f.AppendSQLExclude(dialect, buf, &args, params, nil, excludedTableQualifiers)
	return buf.String(), args, params, err
}

type Predicate interface {
	Field
	Not() Predicate
}

type DB interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
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
	var err error
	for i := 0; i < slice.Len(); i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		v := slice.Index(i).Interface()
		if v, ok := v.(SQLExcludeAppender); ok && v != nil {
			err = v.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
			if err != nil {
				return err
			}
			continue
		}
		if v, ok := v.(SQLAppender); ok && v != nil {
			err = v.AppendSQL(dialect, buf, args, params, nil)
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

func EscapeQuote(str string, quote byte) string {
	i := strings.IndexByte(str, quote)
	if i < 0 {
		return str
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	buf.Grow(len(str))
	escapedQuote := string([]byte{quote, quote})
	for i >= 0 {
		buf.WriteString(str[:i] + escapedQuote)
		if len(str[i:]) > 2 && str[i:i+2] == escapedQuote {
			str = str[i+2:]
		} else {
			str = str[i+1:]
		}
		i = strings.IndexByte(str, quote)
	}
	buf.WriteString(str)
	return buf.String()
}

func QuoteIdentifier(dialect string, identifier string) string {
	needsQuoting := identifier == ""
	for i, char := range identifier {
		if i == 0 && (char >= '0' && char <= '9') {
			// first character cannot be a number
			needsQuoting = true
			break
		}
		if char == '_' || (char >= '0' && char <= '9') || (char >= 'a' && char <= 'z') {
			continue
		}
		// If there are capital letters, the identifier is quoted to preserve
		// capitalization information (because databases treat capital letters
		// differently based on their dialect or configuration).
		// If the character is anything else, we also quote. In general there
		// may be some special characters that are allowed in unquoted
		// identifiers (e.g. '$'), but different databases allow different
		// things. We only recognize _a-z0-9 as the true standard.
		needsQuoting = true
		break
	}
	if !needsQuoting {
		return identifier
	}
	switch dialect {
	case DialectMySQL:
		return "`" + EscapeQuote(identifier, '`') + "`"
	case DialectSQLServer:
		return "[" + EscapeQuote(identifier, ']') + "]"
	default:
		return `"` + EscapeQuote(identifier, '"') + `"`
	}
}

func caller(skip int) (file string, line int, function string) {
	/* https://talks.godoc.org/github.com/davecheney/go-1.9-release-party/presentation.slide#20
	 * "Code that queries a single caller at a specific depth should use Caller
	 * rather than passing a slice of length 1 to Callers."
	 */
	// Skip two extra frames to account for this function and runtime.Caller
	// itself.
	pc, file, line, _ := runtime.Caller(skip + 2)
	fn := runtime.FuncForPC(pc)
	function = fn.Name()
	return file, line, function
}
