package ddl2

import (
	"bytes"
	"strconv"
	"strings"
	"sync"

	"github.com/bokwoon95/sq"
)

var (
	bufpool  = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
	argspool = sync.Pool{New: func() interface{} { return make([]interface{}, 0) }}
)

const (
	PRIMARY_KEY = "PRIMARY KEY"
	FOREIGN_KEY = "FOREIGN KEY"
	UNIQUE      = "UNIQUE"
	CHECK       = "CHECK"
	INDEX       = "INDEX"
	EXCLUDE     = "EXCLUDE"

	BY_DEFAULT_AS_IDENTITY = "BY DEFAULT AS IDENTITY"
	ALWAYS_AS_IDENTITY     = "ALWAYS AS IDENTITY"

	RESTRICT    = "RESTRICT"
	CASCADE     = "CASCADE"
	NO_ACTION   = "NO ACTION"
	SET_NULL    = "SET NULL"
	SET_DEFAULT = "SET DEFAULT"
)

type Command interface {
	sq.SQLAppender
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
	case PRIMARY_KEY:
		buf.WriteString("_pkey")
	case FOREIGN_KEY:
		buf.WriteString("_fkey")
	case UNIQUE:
		buf.WriteString("_key")
	case INDEX:
		buf.WriteString("_idx")
	case CHECK:
		buf.WriteString("_check")
	case EXCLUDE:
		buf.WriteString("_excl")
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
	return "VARCHAR(255)"
}

// for outgoing values to the db
func needsExpressionBrackets(s string) bool {
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return false
	} else if strings.EqualFold(s, "TRUE") ||
		strings.EqualFold(s, "FALSE") ||
		strings.EqualFold(s, "CURRENT_DATE") ||
		strings.EqualFold(s, "CURRENT_TIME") ||
		strings.EqualFold(s, "CURRENT_TIMESTAMP") {
		return false
	} else if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return false
	} else if _, err := strconv.ParseFloat(s, 64); err == nil {
		return false
	} else if len(s) >= 2 && s[0] == '(' && s[len(s)-1] == ')' {
		return false
	}
	return true
}

// for incoming values from the db
func toExpr(dialect, s string) string {
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return s
	} else if strings.EqualFold(s, "TRUE") ||
		strings.EqualFold(s, "FALSE") ||
		strings.EqualFold(s, "CURRENT_DATE") ||
		strings.EqualFold(s, "CURRENT_TIME") ||
		strings.EqualFold(s, "CURRENT_TIMESTAMP") {
		return s
	} else if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return s
	} else if _, err := strconv.ParseFloat(s, 64); err == nil {
		return s
	} else if len(s) >= 2 && s[0] == '(' && s[len(s)-1] == ')' {
		return s
	}
	switch dialect {
	case sq.DialectMySQL:
		return `'` + sq.EscapeQuote(s, '\'') + `'`
	case sq.DialectPostgres:
		return s
	default:
		return "(" + s + ")"
	}
}
