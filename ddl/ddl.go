package ddl

import (
	"bytes"
	"context"
	"database/sql"
	"strings"

	"github.com/bokwoon95/sq"
)

type DB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// I -don't- have to demand that the first field is some anonymous table
// bullshit. Just skip struct fields that aren't UserDefinedColumns, but offer
// to parse their struct tags as long as they declare a ddl:"" inside. Just
// that user can only define the constraint-related ddl inside the
// non-UserDefinedColumn tags.

func pgName(typ string, tableName string, columnNames ...string) string {
	buf := &bytes.Buffer{}
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
	return Metadata{
		Dialect:      dialect,
		schemasCache: make(map[string]int),
	}
}

func (m *Metadata) CachedSchemaIndex(schemaName string) int {
	i, ok := m.schemasCache[schemaName]
	if ok && i >= 0 && i < len(m.Schemas) {
		if schemaName != "" && m.Schemas[i].SchemaName == schemaName {
			return i
		}
	}
	delete(m.schemasCache, schemaName)
	return -1
}

func (m *Metadata) AppendSchema(schema Schema) {
	m.Schemas = append(m.Schemas, schema)
	m.schemasCache[schema.SchemaName] = len(m.Schemas) - 1
}

func (m *Metadata) RefreshSchemaCache() {
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
	return Schema{
		SchemaName:  schemaName,
		tablesCache: make(map[string]int),
		viewsCache:  make(map[string]int),
	}
}

func (s *Schema) CachedTableIndex(tableName string) int {
	i, ok := s.tablesCache[tableName]
	if ok && i >= 0 && i < len(s.Tables) {
		if tableName != "" && s.Tables[i].TableName == tableName {
			return i
		}
	}
	delete(s.tablesCache, tableName)
	return -1
}

func (s *Schema) AppendTable(table Table) {
	s.Tables = append(s.Tables, table)
	s.tablesCache[table.TableName] = len(s.Tables) - 1
}

func (s *Schema) RefreshTableCache() {
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
	columnsCache     map[string]int
	constraintsCache map[string]int
	indicesCache     map[string]int
}

func NewTable(tableSchema, tableName string) Table {
	return Table{
		TableSchema:      tableSchema,
		TableName:        tableName,
		columnsCache:     make(map[string]int),
		constraintsCache: make(map[string]int),
		indicesCache:     make(map[string]int),
	}
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
	tbl.columnsCache[column.ColumnName] = len(tbl.Columns) - 1
}

func (tbl *Table) RefreshColumnCache() {
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
	tbl.constraintsCache[constraint.ConstraintName] = len(tbl.Constraints) - 1
}

func (tbl *Table) RefreshConstraintCache() {
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
	tbl.indicesCache[index.IndexName] = len(tbl.Indices) - 1
}

func (tbl *Table) RefreshIndexCache() {
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
	Identity                 string // NONE | BY DEFAULT AS IDENTITY | ALWAYS AS IDENTITY
	Autoincrement            bool
	IsNotNull                bool
	OnUpdateCurrentTimestamp bool
	GeneratedExpr            sql.NullString
	GeneratedExprStored      sql.NullBool
	CollationName            sql.NullString
	ColumnDefault            sql.NullString
	Modifiers                string
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
