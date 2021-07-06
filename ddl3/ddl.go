package ddl3

import (
	"bytes"
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
	// EXCLUDE = "EXCLUDE" // TODO: support postgres exclusion constraints

	BY_DEFAULT_AS_IDENTITY = "BY DEFAULT AS IDENTITY"
	ALWAYS_AS_IDENTITY     = "ALWAYS AS IDENTITY"

	RESTRICT    = "RESTRICT"
	CASCADE     = "CASCADE"
	NO_ACTION   = "NO ACTION"
	SET_NULL    = "SET NULL"
	SET_DEFAULT = "SET DEFAULT"
)

type Column struct {
	TableSchema              string
	TableName                string
	ColumnName               string
	ColumnType               string
	NormalizedColumnType     string // Not needed if we have a IsEquivalentType(dialect, typeA, typeB string) bool
	Precision                int
	Scale                    int
	Identity                 string
	Autoincrement            bool
	IsNotNull                bool
	IsUnique                 bool
	IsPrimaryKey             bool
	OnUpdateCurrentTimestamp bool
	GeneratedExpr            string
	GeneratedExprStored      bool
	CollationName            string
	ColumnDefault            string
	Ignore                   bool
}

type Constraint struct {
	TableSchema         string
	TableName           string
	ConstraintName      string
	ConstraintType      string
	Columns             []string
	ReferencesSchema    string
	ReferencesTable     string
	ReferencesColumns   []string
	OnUpdate            string
	OnDelete            string
	MatchOption         string
	CheckExpr           string
	IsDeferrable        bool
	IsInitiallyDeferred bool
}

type Index struct {
	TableSchema string
	TableName   string
	IndexName   string
	IndexType   string
	IsUnique    bool
	Columns     []string
	Exprs       []string
	Include     []string
	Where       string
}

const (
	DROP_SCHEMA     = "DROP SCHEMA"
	DROP_TABLE      = "DROP TABLE"
	DROP_COLUMN     = "ALTER TABLE DROP COLUMN"
	DROP_CONSTRAINT = "ALTER TABLE DROP CONSTRAINT"
	DROP_INDEX      = "DROP INDEX"
	DROP_VIEW       = "DROP VIEW"
	DROP_FUNCTION   = "DROP FUNCTION"
	DROP_TRIGGER    = "DROP TRIGGER"

	CREATE_SCHEMA   = "CREATE SCHEMA"
	CREATE_TABLE    = "CREATE TABLE"
	ADD_COLUMN      = "ALTER TABLE ADD COLUMN"
	ADD_CONSTRAINT  = "ALTER TABLE ADD CONSTRAINT"
	CREATE_INDEX    = "CREATE INDEX"
	CREATE_VIEW     = "CREATE VIEW"     // use CREATE OR REPLACE where possible
	CREATE_FUNCTION = "CREATE FUNCTION" // use CREATE OR REPLACE where possible
	CREATE_TRIGGER  = "CREATE TRIGGER"

	ALTER_COLUMN = "ALTER TABLE ALTER COLUMN"
)

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