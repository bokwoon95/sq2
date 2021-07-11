package ddl3

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

type Column struct {
	TableSchema              string
	TableName                string
	TableAlias               string
	ColumnName               string
	ColumnAlias              string
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
	TableSchema         string
	TableName           string
	ConstraintName      string
	ConstraintType      string
	Columns             []string
	Exprs               []string
	ReferencesSchema    string
	ReferencesTable     string
	ReferencesColumns   []string
	OnUpdate            string
	OnDelete            string
	MatchOption         string
	CheckExpr           string
	Operators           []string
	IndexType           string
	Where               string
	IsDeferrable        bool
	IsInitiallyDeferred bool
}

type Exclusions []struct {
	Field    sq.Field
	Operator string
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
	return "TEXT"
}

type CommandType int

const (
	// TODO: the top should include classes of operations that are usually used together
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
