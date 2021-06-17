package ddl

import (
	"bytes"
	"context"
	"database/sql"
	"strings"
	"sync"

	"github.com/bokwoon95/sq"
)

type DB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

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

	IDENTITY_DEFAULT = "BY DEFAULT AS IDENTITY"
	IDENTITY_ALWAYS  = "ALWAYS AS IDENTITY"
)

// I -don't- have to demand that the first field is some anonymous table
// bullshit. Just skip struct fields that aren't UserDefinedColumns, but offer
// to parse their struct tags as long as they declare a ddl:"" inside. Just
// that user can only define the constraint-related ddl inside the
// non-UserDefinedColumn tags.

func pgName(typ string, tableName string, columnNames ...string) string {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	buf.WriteString(strings.ReplaceAll(tableName, " ", "_"))
	for _, columnName := range columnNames {
		buf.WriteString("_" + strings.ReplaceAll(columnName, " ", "_"))
	}
	switch typ {
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

type Metadata struct {
	Dialect       string
	VersionString string
	VersionNum    [2]int // MAJOR.MINOR (are PATCH versions ever significant in the case of databases?)
	Schemas       []Schema
	schemasCache  map[string]int
}

// func (m *Metadata) LoadMetadata(db DB, opts ...LoatMetadataOptions)
// - IncludeSchemas, ExcludeSchemas, IncludeTables, ExcludeTables // tables are all schema qualified?
// - ExcludeConstraints, ExcludeIndices (ever relevant?)

func NewMetadata(dialect string) Metadata {
	return Metadata{Dialect: dialect}
}

func (m *Metadata) CachedSchemaIndex(schemaName string) int {
	if schemaName == "" {
		delete(m.schemasCache, schemaName)
		return -1
	}
	i, ok := m.schemasCache[schemaName]
	if !ok || i < 0 || i >= len(m.Schemas) {
		delete(m.schemasCache, schemaName)
		return -1
	}
	return i
}

func (m *Metadata) AppendSchema(schema Schema) {
	m.Schemas = append(m.Schemas, schema)
	if m.schemasCache == nil {
		m.schemasCache = make(map[string]int)
	}
	m.schemasCache[schema.SchemaName] = len(m.Schemas) - 1
}

func (m *Metadata) RefreshSchemaCache() {
	if m.schemasCache == nil {
		m.schemasCache = make(map[string]int)
	}
	for i, schema := range m.Schemas {
		m.schemasCache[schema.SchemaName] = i
	}
}

type Schema struct {
	SchemaName  string
	Tables      []Table
	Views       []View
	tablesCache map[string]int
	viewsCache  map[string]int
}

func NewSchema(schemaName string) Schema {
	return Schema{SchemaName: schemaName}
}

func (s *Schema) CachedTableIndex(tableName string) int {
	if tableName == "" {
		delete(s.tablesCache, tableName)
		return -1
	}
	i, ok := s.tablesCache[tableName]
	if !ok || i < 0 || i >= len(s.Tables) {
		delete(s.tablesCache, tableName)
		return -1
	}
	return i
}

func (s *Schema) AppendTable(table Table) {
	s.Tables = append(s.Tables, table)
	if s.tablesCache == nil {
		s.tablesCache = make(map[string]int)
	}
	s.tablesCache[table.TableName] = len(s.Tables) - 1
}

func (s *Schema) RefreshTableCache() {
	if s.tablesCache == nil {
		s.tablesCache = make(map[string]int)
	}
	for i, table := range s.Tables {
		s.tablesCache[table.TableName] = i
	}
}

type Table struct {
	TableSchema      string
	TableName        string
	Columns          []Column
	Constraints      []Constraint
	Indices          []Index
	Modifiers        string
	Comment          string
	columnsCache     map[string]int
	constraintsCache map[string]int
	indicesCache     map[string]int
}

func NewTable(tableSchema, tableName string) Table {
	return Table{TableSchema: tableSchema, TableName: tableName}
}

func (tbl *Table) CachedColumnIndex(columnName string) int {
	if columnName == "" {
		delete(tbl.columnsCache, columnName)
		return -1
	}
	i, ok := tbl.columnsCache[columnName]
	if !ok || i < 0 && i >= len(tbl.Columns) || tbl.Columns[i].ColumnName != columnName {
		delete(tbl.columnsCache, columnName)
		return -1
	}
	return i
}

func (tbl *Table) AppendColumn(column Column) {
	tbl.Columns = append(tbl.Columns, column)
	if tbl.columnsCache == nil {
		tbl.columnsCache = make(map[string]int)
	}
	tbl.columnsCache[column.ColumnName] = len(tbl.Columns) - 1
}

func (tbl *Table) RefreshColumnCache() {
	if tbl.columnsCache == nil {
		tbl.columnsCache = make(map[string]int)
	}
	for i, column := range tbl.Columns {
		tbl.columnsCache[column.ColumnName] = i
	}
}

func (tbl *Table) CachedConstraintIndex(constraintName string) int {
	if constraintName == "" {
		delete(tbl.constraintsCache, constraintName)
		return -1
	}
	i, ok := tbl.constraintsCache[constraintName]
	if !ok || i < 0 || i >= len(tbl.Constraints) || tbl.Constraints[i].ConstraintName != constraintName {
		delete(tbl.constraintsCache, constraintName)
		return -1
	}
	return i
}

func (tbl *Table) AppendConstraint(constraint Constraint) {
	tbl.Constraints = append(tbl.Constraints, constraint)
	if tbl.constraintsCache == nil {
		tbl.constraintsCache = make(map[string]int)
	}
	tbl.constraintsCache[constraint.ConstraintName] = len(tbl.Constraints) - 1
}

func (tbl *Table) RefreshConstraintCache() {
	if tbl.constraintsCache == nil {
		tbl.constraintsCache = make(map[string]int)
	}
	for i, constraint := range tbl.Constraints {
		tbl.constraintsCache[constraint.ConstraintName] = i
	}
}

func (tbl *Table) CachedIndexIndex(indexName string) int {
	if indexName == "" {
		delete(tbl.indicesCache, indexName)
		return -1
	}
	i, ok := tbl.indicesCache[indexName]
	if ok && i >= 0 && i < len(tbl.Indices) {
		if tbl.Indices[i].IndexName == indexName {
			return i
		}
	}
	delete(tbl.indicesCache, indexName)
	return -1
}

func (tbl *Table) AppendIndex(index Index) {
	tbl.Indices = append(tbl.Indices, index)
	if tbl.indicesCache == nil {
		tbl.indicesCache = make(map[string]int)
	}
	tbl.indicesCache[index.IndexName] = len(tbl.Indices) - 1
}

func (tbl *Table) RefreshIndexCache() {
	if tbl.indicesCache == nil {
		tbl.indicesCache = make(map[string]int)
	}
	for i, index := range tbl.Indices {
		tbl.indicesCache[index.IndexName] = i
	}
}

type Column struct {
	TableSchema              string
	TableName                string
	ColumnName               string
	ColumnType               string
	NormalizedColumnType     string
	Precision                int
	Scale                    int
	Identity                 string
	Autoincrement            bool
	IsNotNull                bool
	OnUpdateCurrentTimestamp bool
	GeneratedExpr            sql.NullString
	GeneratedExprStored      sql.NullBool
	CollationName            sql.NullString
	ColumnDefault            sql.NullString
	Modifiers                string
	Comment                  string
	Ignore                   bool
}

type Constraint struct {
	ConstraintSchema    string
	ConstraintName      string
	ConstraintType      string
	TableSchema         string
	TableName           string
	Columns             []string
	ReferencesSchema    sql.NullString
	ReferencesTable     sql.NullString
	ReferencesColumns   []string
	OnUpdate            sql.NullString
	OnDelete            sql.NullString
	MatchOption         sql.NullString
	CheckExpr           sql.NullString
	IsDeferrable        bool
	IsInitiallyDeferred bool
}

type Index struct {
	IndexSchema string
	IndexName   string
	IndexType   string
	IsUnique    bool
	TableSchema string
	TableName   string
	Columns     []string
	Exprs       []string
	Include     []string
	Predicate   sql.NullString
}

type View struct {
}
