package ddl2

import (
	"bytes"
	"io"
	"strings"
	"sync"

	"github.com/bokwoon95/sq"
)

var (
	bufpool = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
	// argspool = sync.Pool{New: func() interface{} { return make([]interface{}, 0) }}
)

const (
	PRIMARY_KEY = "PRIMARY KEY"
	FOREIGN_KEY = "FOREIGN KEY"
	UNIQUE      = "UNIQUE"
	CHECK       = "CHECK"
	INDEX       = "INDEX"

	BY_DEFAULT_AS_IDENTITY = "BY DEFAULT AS IDENTITY"
	ALWAYS_AS_IDENTITY     = "ALWAYS AS IDENTITY"

	RESTRICT    = "RESTRICT"
	CASCADE     = "CASCADE"
	NO_ACTION   = "NO ACTION"
	SET_NULL    = "SET NULL"
	SET_DEFAULT = "SET DEFAULT"
)

type Config struct {
	DiffColumn     func(dialect string, gotColumn, wantColumn Column) ([]string, error)
	DiffConstraint func(dialect string, gotConstraint, wantConstraint Constraint) ([]string, error)
	DiffIndex      func(dialect string, gotIndex, wantIndex Index) ([]string, error)
	WantFunctions  []Function
	WantViews      []View
}

type Object struct {
	Type   string // VIEW | FUNCTION | TRIGGER
	Schema string
	Name   string
	SQL    []string
}

type Function struct {
	FunctionSchema string
	FunctionName   string
	SQL            []io.Reader
}

type View interface {
	sq.SchemaTable
	// TODO: extra argument that can be used to register certain view
	// porperties like MATERIALIZED or RECURSIVE.
	View(dialect string) sq.Query
}

func generateName(nameType string, tableName string, columnNames ...string) string {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	buf.WriteString(strings.ReplaceAll(tableName, " ", "_"))
	for _, columnName := range columnNames {
		buf.WriteString("_" + strings.ReplaceAll(columnName, " ", "_"))
	}
	switch nameType {
	case "PRIMARY KEY":
		buf.WriteString("_pkey")
	case "FOREIGN KEY":
		buf.WriteString("_fkey")
	case "UNIQUE":
		buf.WriteString("_key")
	case "INDEX":
		buf.WriteString("_idx")
	case "CHECK":
		buf.WriteString("_check")
	}
	return buf.String()
}

func defaultColumnType(dialect string, field sq.Field) (columnType string) {
	switch field.(type) {
	case sq.BlobField:
		switch dialect {
		case sq.DialectPostgres:
			return "BYTEA"
		default:
			return "BLOB"
		}
	case sq.BooleanField:
		return "BOOLEAN"
	case sq.JSONField:
		switch dialect {
		case sq.DialectPostgres:
			return "JSONB"
		default:
			return "JSON"
		}
	case sq.NumberField:
		return "INT"
	case sq.StringField:
		switch dialect {
		case sq.DialectPostgres, sq.DialectSQLite:
			return "TEXT"
		default:
			return "VARCHAR(255)"
		}
	case sq.TimeField:
		switch dialect {
		case sq.DialectPostgres:
			return "TIMESTAMPTZ"
		default:
			return "DATETIME"
		}
	}
	return "TEXT"
}
