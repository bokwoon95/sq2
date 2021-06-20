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

type DDLer interface {
	DDL(dialect string, t *T)
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

	BY_DEFAULT_AS_IDENTITY = "BY DEFAULT AS IDENTITY"
	ALWAYS_AS_IDENTITY     = "ALWAYS AS IDENTITY"

	RESTRICT    = "RESTRICT"
	CASCADE     = "CASCADE"
	NO_ACTION   = "NO ACTION"
	SET_NULL    = "SET NULL"
	SET_DEFAULT = "SET DEFAULT"
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

func (m *Metadata) CachedSchemaIndex(schemaName string) (schemaIndex int) {
	schemaIndex, ok := m.schemasCache[schemaName]
	if !ok || schemaIndex < 0 || schemaIndex >= len(m.Schemas) {
		delete(m.schemasCache, schemaName)
		return -1
	}
	return schemaIndex
}

func (m *Metadata) AppendSchema(schema Schema) (schemaIndex int) {
	m.Schemas = append(m.Schemas, schema)
	if m.schemasCache == nil {
		m.schemasCache = make(map[string]int)
	}
	schemaIndex = len(m.Schemas) - 1
	m.schemasCache[schema.SchemaName] = schemaIndex
	return schemaIndex
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

func (s *Schema) CachedTableIndex(tableName string) (tableIndex int) {
	if tableName == "" {
		delete(s.tablesCache, tableName)
		return -1
	}
	tableIndex, ok := s.tablesCache[tableName]
	if !ok || tableIndex < 0 || tableIndex >= len(s.Tables) {
		delete(s.tablesCache, tableName)
		return -1
	}
	return tableIndex
}

func (s *Schema) AppendTable(table Table) (tableIndex int) {
	s.Tables = append(s.Tables, table)
	if s.tablesCache == nil {
		s.tablesCache = make(map[string]int)
	}
	tableIndex = len(s.Tables) - 1
	s.tablesCache[table.TableName] = tableIndex
	return tableIndex
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

func (tbl *Table) CachedColumnIndex(columnName string) (columnIndex int) {
	if columnName == "" {
		delete(tbl.columnsCache, columnName)
		return -1
	}
	columnIndex, ok := tbl.columnsCache[columnName]
	if !ok || columnIndex < 0 && columnIndex >= len(tbl.Columns) || tbl.Columns[columnIndex].ColumnName != columnName {
		delete(tbl.columnsCache, columnName)
		return -1
	}
	return columnIndex
}

func (tbl *Table) AppendColumn(column Column) (columnIndex int) {
	tbl.Columns = append(tbl.Columns, column)
	if tbl.columnsCache == nil {
		tbl.columnsCache = make(map[string]int)
	}
	columnIndex = len(tbl.Columns) - 1
	tbl.columnsCache[column.ColumnName] = columnIndex
	return columnIndex
}

func (tbl *Table) RefreshColumnCache() {
	if tbl.columnsCache == nil {
		tbl.columnsCache = make(map[string]int)
	}
	for i, column := range tbl.Columns {
		tbl.columnsCache[column.ColumnName] = i
	}
}

func (tbl *Table) CachedConstraintIndex(constraintName string) (constraintIndex int) {
	if constraintName == "" {
		delete(tbl.constraintsCache, constraintName)
		return -1
	}
	constraintIndex, ok := tbl.constraintsCache[constraintName]
	if !ok || constraintIndex < 0 || constraintIndex >= len(tbl.Constraints) || tbl.Constraints[constraintIndex].ConstraintName != constraintName {
		delete(tbl.constraintsCache, constraintName)
		return -1
	}
	return constraintIndex
}

func (tbl *Table) AppendConstraint(constraint Constraint) (constraintIndex int) {
	tbl.Constraints = append(tbl.Constraints, constraint)
	if tbl.constraintsCache == nil {
		tbl.constraintsCache = make(map[string]int)
	}
	constraintIndex = len(tbl.Constraints) - 1
	tbl.constraintsCache[constraint.ConstraintName] = constraintIndex
	return constraintIndex
}

func (tbl *Table) RefreshConstraintCache() {
	if tbl.constraintsCache == nil {
		tbl.constraintsCache = make(map[string]int)
	}
	for i, constraint := range tbl.Constraints {
		tbl.constraintsCache[constraint.ConstraintName] = i
	}
}

func (tbl *Table) CachedIndexIndex(indexName string) (indexIndex int) {
	if indexName == "" {
		delete(tbl.indicesCache, indexName)
		return -1
	}
	indexIndex, ok := tbl.indicesCache[indexName]
	if !ok || indexIndex < 0 || indexIndex >= len(tbl.Indices) || tbl.Indices[indexIndex].IndexName != indexName {
		delete(tbl.indicesCache, indexName)
		return -1
	}
	return indexIndex
}

func (tbl *Table) AppendIndex(index Index) (indexIndex int) {
	tbl.Indices = append(tbl.Indices, index)
	if tbl.indicesCache == nil {
		tbl.indicesCache = make(map[string]int)
	}
	indexIndex = len(tbl.Indices) - 1
	tbl.indicesCache[index.IndexName] = indexIndex
	return indexIndex
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
	IsUnique                 bool
	IsPrimaryKey             bool
	OnUpdateCurrentTimestamp bool
	GeneratedExpr            string
	GeneratedExprStored      bool
	CollationName            string
	ColumnDefault            string
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
	ReferencesSchema    string
	ReferencesTable     string
	ReferencesColumns   []string
	OnUpdate            string
	OnDelete            string
	MatchOption         string
	CheckExpr           string
	IsDeferrable        bool
	IsInitiallyDeferred bool
	Comment             string
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
	Predicate   string
	Comment     string
}

type View struct {
}
