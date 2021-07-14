package ddl2

import (
	"bytes"
	"sort"
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

type CommandType int

const (
	CREATE_SCHEMA CommandType = iota
	RENAME_SCHEMA
	DROP_SCHEMA
	CREATE_TABLE
	RENAME_TABLE
	DROP_TABLE
	ADD_COLUMN
	ALTER_COLUMN
	RENAME_COLUMN
	DROP_COLUMN
	ADD_CONSTRAINT
	RENAME_CONSTRAINT
	DROP_CONSTRAINT
	CREATE_INDEX
	RENAME_INDEX
	DROP_INDEX
	CREATE_FUNCTION
	RENAME_FUNCTION
	DROP_FUNCTION
	CREATE_VIEW
	RENAME_VIEW
	DROP_VIEW
	CREATE_TRIGGER
	RENAME_TRIGGER
	DROP_TRIGGER
	TABLE_DML
)

type Command interface {
	sq.SQLAppender
}

type Column struct {
	TableSchema              string `json:",omitempty"`
	TableName                string `json:",omitempty"`
	TableAlias               string `json:",omitempty"`
	ColumnName               string `json:",omitempty"`
	ColumnAlias              string `json:",omitempty"`
	ColumnType               string `json:",omitempty"`
	Precision                int    `json:",omitempty"`
	Scale                    int    `json:",omitempty"`
	Identity                 string `json:",omitempty"`
	Autoincrement            bool   `json:",omitempty"`
	IsNotNull                bool   `json:",omitempty"`
	IsUnique                 bool   `json:",omitempty"`
	IsPrimaryKey             bool   `json:",omitempty"`
	OnUpdateCurrentTimestamp bool   `json:",omitempty"`
	GeneratedExpr            string `json:",omitempty"`
	GeneratedExprStored      bool   `json:",omitempty"`
	CollationName            string `json:",omitempty"`
	ColumnDefault            string `json:",omitempty"`
	Ignore                   bool   `json:",omitempty"`
}

var _ sq.Field = Column{}

func (c Column) GetName() string { return c.ColumnName }

func (c Column) GetAlias() string { return c.ColumnAlias }

func (c Column) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	tableQualifier := c.TableAlias
	if tableQualifier == "" {
		tableQualifier = c.TableName
	}
	if tableQualifier != "" {
		i := sort.SearchStrings(excludedTableQualifiers, tableQualifier)
		if i < len(excludedTableQualifiers) && excludedTableQualifiers[i] == tableQualifier {
			tableQualifier = ""
		}
	}
	if tableQualifier != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, tableQualifier) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, c.ColumnName))
	return nil
}

type Constraint struct {
	TableSchema         string   `json:",omitempty"`
	TableName           string   `json:",omitempty"`
	ConstraintName      string   `json:",omitempty"`
	ConstraintType      string   `json:",omitempty"`
	Columns             []string `json:",omitempty"`
	Exprs               []string `json:",omitempty"`
	ReferencesSchema    string   `json:",omitempty"`
	ReferencesTable     string   `json:",omitempty"`
	ReferencesColumns   []string `json:",omitempty"`
	OnUpdate            string   `json:",omitempty"`
	OnDelete            string   `json:",omitempty"`
	MatchOption         string   `json:",omitempty"`
	CheckExpr           string   `json:",omitempty"`
	Operators           []string `json:",omitempty"`
	IndexType           string   `json:",omitempty"`
	Where               string   `json:",omitempty"`
	IsDeferrable        bool     `json:",omitempty"`
	IsInitiallyDeferred bool     `json:",omitempty"`
}

type Exclusions []struct {
	Field    sq.Field
	Operator string
}

type Index struct {
	TableSchema string   `json:",omitempty"`
	TableName   string   `json:",omitempty"`
	IndexName   string   `json:",omitempty"`
	IndexType   string   `json:",omitempty"`
	IsUnique    bool     `json:",omitempty"`
	Columns     []string `json:",omitempty"`
	Exprs       []string `json:",omitempty"`
	Include     []string `json:",omitempty"`
	Where       string   `json:",omitempty"`
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
