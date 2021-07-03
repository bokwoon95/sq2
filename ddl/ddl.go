package ddl

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/bokwoon95/sq"
)

type DB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type DDLer interface {
	sq.Table
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
	Dialect         string
	VersionString   string
	VersionNum      [2]int // MAJOR.MINOR (are PATCH versions ever significant in the case of databases?)
	GeneratedFromDB bool
	Schemas         []Schema
	schemasCache    map[string]int
}

// func (m *Metadata) LoadMetadata(db DB, opts ...LoatMetadataOptions)
// - IncludeSchemas, ExcludeSchemas, IncludeTables, ExcludeTables // tables are all schema qualified?
// - ExcludeConstraints, ExcludeIndices (ever relevant?)

func NewMetadata(dialect string) Metadata {
	return Metadata{Dialect: dialect}
}

func (m *Metadata) CachedSchemaIndex(schemaName string) (schemaIndex int) {
	if m == nil {
		return -1
	}
	schemaIndex, ok := m.schemasCache[schemaName]
	if !ok || schemaIndex < 0 || schemaIndex >= len(m.Schemas) {
		delete(m.schemasCache, schemaName)
		return -1
	}
	if m.Schemas[schemaIndex].SchemaName != schemaName {
		delete(m.schemasCache, schemaName)
		return -1
	}
	return schemaIndex
}

func (m *Metadata) AppendSchema(schema Schema) (schemaIndex int) {
	if m == nil {
		return -1
	}
	m.Schemas = append(m.Schemas, schema)
	if m.schemasCache == nil {
		m.schemasCache = make(map[string]int)
	}
	schemaIndex = len(m.Schemas) - 1
	m.schemasCache[schema.SchemaName] = schemaIndex
	return schemaIndex
}

func (m *Metadata) RefreshSchemaCache() {
	if m == nil {
		return
	}
	for i, schema := range m.Schemas {
		if m.schemasCache == nil {
			m.schemasCache = make(map[string]int)
		}
		m.schemasCache[schema.SchemaName] = i
	}
}

type Schema struct {
	SchemaName     string
	Tables         []Table
	Views          []Object
	Functions      []Object
	tablesCache    map[string]int
	viewsCache     map[string]int
	functionsCache map[string]int
}

func NewSchema(schemaName string) Schema {
	return Schema{SchemaName: schemaName}
}

func (s *Schema) CachedTableIndex(tableName string) (tableIndex int) {
	if s == nil {
		return -1
	}
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
	if s == nil {
		return -1
	}
	s.Tables = append(s.Tables, table)
	if s.tablesCache == nil {
		s.tablesCache = make(map[string]int)
	}
	tableIndex = len(s.Tables) - 1
	s.tablesCache[table.TableName] = tableIndex
	return tableIndex
}

func (s *Schema) RefreshTableCache() {
	if s == nil {
		return
	}
	for i, table := range s.Tables {
		if s.tablesCache == nil {
			s.tablesCache = make(map[string]int)
		}
		s.tablesCache[table.TableName] = i
	}
}

type Table struct {
	TableSchema      string
	TableName        string
	Columns          []Column
	Constraints      []Constraint
	Indices          []Index
	Triggers         []Object
	VirtualTable     string
	VirtualTableArgs []string
	columnsCache     map[string]int
	constraintsCache map[string]int
	indicesCache     map[string]int
	triggersCache    map[string]int
}

func NewTable(tableSchema, tableName string) Table {
	return Table{TableSchema: tableSchema, TableName: tableName}
}

func (tbl *Table) CachedColumnIndex(columnName string) (columnIndex int) {
	if tbl == nil {
		return -1
	}
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
	if tbl == nil {
		return -1
	}
	tbl.Columns = append(tbl.Columns, column)
	if tbl.columnsCache == nil {
		tbl.columnsCache = make(map[string]int)
	}
	columnIndex = len(tbl.Columns) - 1
	tbl.columnsCache[column.ColumnName] = columnIndex
	return columnIndex
}

func (tbl *Table) RefreshColumnCache() {
	if tbl == nil {
		return
	}
	for i, column := range tbl.Columns {
		if tbl.columnsCache == nil {
			tbl.columnsCache = make(map[string]int)
		}
		tbl.columnsCache[column.ColumnName] = i
	}
}

func (tbl *Table) CachedConstraintIndex(constraintName string) (constraintIndex int) {
	if tbl == nil {
		return -1
	}
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
	if tbl == nil {
		return -1
	}
	tbl.Constraints = append(tbl.Constraints, constraint)
	if tbl.constraintsCache == nil {
		tbl.constraintsCache = make(map[string]int)
	}
	constraintIndex = len(tbl.Constraints) - 1
	tbl.constraintsCache[constraint.ConstraintName] = constraintIndex
	return constraintIndex
}

func (tbl *Table) RefreshConstraintCache() {
	if tbl == nil {
		return
	}
	for i, constraint := range tbl.Constraints {
		if tbl.constraintsCache == nil {
			tbl.constraintsCache = make(map[string]int)
		}
		tbl.constraintsCache[constraint.ConstraintName] = i
	}
}

func (tbl *Table) CachedIndexIndex(indexName string) (indexIndex int) {
	if tbl == nil {
		return -1
	}
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
	if tbl == nil {
		return -1
	}
	tbl.Indices = append(tbl.Indices, index)
	if tbl.indicesCache == nil {
		tbl.indicesCache = make(map[string]int)
	}
	indexIndex = len(tbl.Indices) - 1
	tbl.indicesCache[index.IndexName] = indexIndex
	return indexIndex
}

func (tbl *Table) RefreshIndexCache() {
	if tbl == nil {
		return
	}
	for i, index := range tbl.Indices {
		if tbl.indicesCache == nil {
			tbl.indicesCache = make(map[string]int)
		}
		tbl.indicesCache[index.IndexName] = i
	}
}

func (tbl *Table) CachedTriggerIndex(triggerName string) (triggerIndex int) {
	if tbl == nil {
		return -1
	}
	if triggerName == "" {
		delete(tbl.triggersCache, triggerName)
		return -1
	}
	triggerIndex, ok := tbl.triggersCache[triggerName]
	if !ok || triggerIndex < 0 || triggerIndex >= len(tbl.Triggers) || tbl.Triggers[triggerIndex].Name != triggerName {
		delete(tbl.triggersCache, triggerName)
		return -1
	}
	return triggerIndex
}

func (tbl *Table) RefreshTriggerCache() {
	if tbl == nil {
		return
	}
	for i, trigger := range tbl.Triggers {
		if tbl.triggersCache == nil {
			tbl.triggersCache = make(map[string]int)
		}
		tbl.triggersCache[trigger.Name] = i
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
	Where       string
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

type Config struct {
	DiffColumn     func(dialect string, gotColumn, wantColumn Column) ([]string, error)
	DiffConstraint func(dialect string, gotConstraint, wantConstraint Constraint) ([]string, error)
	DiffIndex      func(dialect string, gotIndex, wantIndex Index) ([]string, error)
	WantFunctions  []Function
	WantViews      []View
}

func NewMetadataFromDB(dialect string, db sq.Queryer) (Metadata, error) {
	m := NewMetadata(dialect)
	return m, nil
}

func NewMetadataFromTables(dialect string, tables []sq.SchemaTable) (Metadata, error) {
	m := NewMetadata(dialect)
	var err error
	for _, table := range tables {
		err = m.LoadTable(table)
		if err != nil {
			qualifiedTableName := table.GetSchema() + "." + table.GetName()
			if qualifiedTableName[0] == '.' {
				qualifiedTableName = qualifiedTableName[1:]
			}
			return m, fmt.Errorf("table %s: %w", qualifiedTableName, err)
		}
	}
	return m, nil
}

func Diff(dialect string, gotMetadata, wantMetadata Metadata, config Config) ([]string, error) {
	gotMetadata.RefreshSchemaCache()
	diffColumn := config.DiffColumn
	if diffColumn == nil {
		diffColumn = DiffColumn
	}
	diffConstraint := config.DiffConstraint
	if diffConstraint == nil {
		diffConstraint = DiffConstraint
	}
	diffIndex := config.DiffIndex
	if diffIndex == nil {
		diffIndex = DiffIndex
	}
	var stmts []string
	schemaViews := make(map[string][]View)
	for i, view := range config.WantViews {
		if view == nil {
			return nil, fmt.Errorf("config: view #%d is nil", i+1)
		}
		viewSchema := view.GetSchema()
		schemaViews[viewSchema] = append(schemaViews[viewSchema], view)
	}
	schemaFunctions := make(map[string][]Function)
	for _, function := range config.WantFunctions {
		schemaFunctions[function.FunctionSchema] = append(schemaFunctions[function.FunctionSchema], function)
	}
	var fkeyStmts []string
	var functionStmts []string
	var viewStmts []string
	var triggerStmts []string
	for _, wantSchema := range wantMetadata.Schemas {
		var gotSchema Schema
		gotSchemaIndex := gotMetadata.CachedSchemaIndex(wantSchema.SchemaName)
		if gotSchemaIndex < 0 {
			if dialect == sq.DialectSQLite {
				return nil, fmt.Errorf("cannot create missing schema '%s' for database because sqlite does not support CREATE SCHEMA", wantSchema.SchemaName)
			}
			stmts = append(stmts, "CREATE SCHEMA "+wantSchema.SchemaName+";")
			gotSchema.SchemaName = wantSchema.SchemaName
		} else if wantSchema.SchemaName != "" {
			gotSchema = gotMetadata.Schemas[gotSchemaIndex]
			gotSchema.RefreshTableCache()
		}
		schemaConfig := config
		schemaConfig.WantViews = schemaViews[wantSchema.SchemaName]
		schemaConfig.WantFunctions = schemaFunctions[wantSchema.SchemaName]
		for _, wantTable := range wantSchema.Tables {
			var gotTable Table
			gotTableIndex := gotSchema.CachedTableIndex(wantTable.TableName)
			if gotTableIndex < 0 {
				s, err := CreateTable(dialect, wantTable)
				if err != nil {
					return nil, fmt.Errorf("table %s: %w", wantTable.TableName, err)
				}
				stmts = append(stmts, s)
				gotTable.TableName = wantTable.TableName
			} else {
				gotTable = gotSchema.Tables[gotTableIndex]
				gotTable.RefreshColumnCache()
				for _, wantColumn := range wantTable.Columns {
					gotColumnIndex := gotTable.CachedColumnIndex(wantColumn.ColumnName)
					if gotColumnIndex < 0 {
						s, err := CreateColumn(dialect, wantColumn)
						if err != nil {
							return nil, fmt.Errorf("table %s column %s: %w", wantTable.TableName, wantColumn.ColumnName, err)
						}
						stmts = append(stmts, s)
					} else {
						gotColumn := gotTable.Columns[gotColumnIndex]
						ss, err := diffColumn(dialect, gotColumn, wantColumn)
						if err != nil {
							return nil, fmt.Errorf("table %s column %s: %w", wantTable.TableName, wantColumn.ColumnName, err)
						}
						stmts = append(stmts, ss...)
					}
				}
				for _, wantConstraint := range wantTable.Constraints {
					if wantConstraint.ConstraintType == FOREIGN_KEY {
						continue
					}
					gotConstraintIndex := gotTable.CachedConstraintIndex(wantConstraint.ConstraintName)
					if gotConstraintIndex < 0 {
						s, err := CreateConstraint(dialect, wantConstraint)
						if err != nil {
							return nil, fmt.Errorf("table %s constraint %s: %w", wantTable.TableName, wantConstraint.ConstraintName, err)
						}
						stmts = append(stmts, s)
					} else {
						gotConstraint := gotTable.Constraints[gotConstraintIndex]
						ss, err := diffConstraint(dialect, gotConstraint, wantConstraint)
						if err != nil {
							return nil, fmt.Errorf("table %s constraint %s: %w", wantTable.TableName, wantConstraint.ConstraintName, err)
						}
						stmts = append(stmts, ss...)
					}
				}
			}
			for _, wantConstraint := range wantTable.Constraints {
				if wantConstraint.ConstraintType != FOREIGN_KEY {
					continue
				}
				gotConstraintIndex := gotTable.CachedConstraintIndex(wantConstraint.ConstraintName)
				if gotConstraintIndex < 0 {
					s, err := CreateConstraint(dialect, wantConstraint)
					if err != nil {
						return nil, fmt.Errorf("table %s constraint %s: %w", wantTable.TableName, wantConstraint.ConstraintName, err)
					}
					fkeyStmts = append(fkeyStmts, s)
				} else {
					gotConstraint := gotTable.Constraints[gotConstraintIndex]
					ss, err := diffConstraint(dialect, gotConstraint, wantConstraint)
					if err != nil {
						return nil, fmt.Errorf("table %s constraint %s: %w", wantTable.TableName, wantConstraint.ConstraintName, err)
					}
					fkeyStmts = append(fkeyStmts, ss...)
				}
			}
			for _, wantIndex := range wantTable.Indices {
				gotIndexIndex := gotTable.CachedIndexIndex(wantIndex.IndexName)
				if gotIndexIndex < 0 {
					s, err := CreateIndex(dialect, wantIndex)
					if err != nil {
						return nil, fmt.Errorf("table %s index %s: %w", wantTable.TableName, wantIndex.IndexName, err)
					}
					stmts = append(stmts, s)
				} else {
					gotIndex := gotTable.Indices[gotIndexIndex]
					ss, err := diffIndex(dialect, gotIndex, wantIndex)
					if err != nil {
						return nil, fmt.Errorf("table %s index %s: %w", wantTable.TableName, wantIndex.IndexName, err)
					}
					stmts = append(stmts, ss...)
				}
			}
			for _, wantTrigger := range wantTable.Triggers {
				gotTriggerIndex := gotTable.CachedTriggerIndex(wantTrigger.Name)
				if gotTriggerIndex < 0 {
					if len(wantTrigger.SQL) == 0 {
						return nil, fmt.Errorf("table %s trigger %s has no SQL", wantTable.TableName, wantTrigger.Name)
					}
					triggerStmts = append(triggerStmts, wantTrigger.SQL[0])
				}
			}
		}
	}
	// TODO: the depedency between functions and views can be circular. The only way to be sure is to pass these things to the user to handle themselves.
	// There are 5 classes of stmts []string
	// Class 1)
	// CREATE TABLE
	// -or-
	// ALTER TABLE ADD COLUMN | ALTER TABLE ALTER COLUMN | ALTER TABLE ADD CONSTRAINT | ALTER TABLE ALTER CONSTRAINT | CREATE INDEX | ALTER INDEX
	// Class 2)
	// ALTER TABLE ADD CONSTRAINT (fkeys)
	// Class 3)
	// CREATE VIEW
	// Class 4)
	// CREATE FUNCTION
	// Class 5)
	// CREATE TRIGGER
	// by default all Class 1s are executed first, followed by Class 2s, Class 3s, Class 4s and then Class 5s.
	// but the important thing is that Diff will return all these Classes as
	// distinct items so that the user can reorder them as he wishes. For
	// example he may know that all functions can be created after views,
	// except for one function which a view requires and so he can move that
	// CREATE FUNCTION statement up the hierarchy.
	// type Stmts { []TableStmts; []ForeignKeyStmts; []ViewStmts; []FunctionStmts; []TriggerStmts }
	// type TableStmts { TableSchema, TableName string; Stmts []string }
	// type ForeignKeyStmts { TableSchema, ConstraintName string; Stmts []string }
	// type ViewStmts { ViewSchema, ViewName string; Stmts []string }
	// type FunctionStmts { FunctionSchema, FunctionName string, Stmts []string }
	// type TriggerStmts { TableSchema, TableName, TriggerName string; Stmts []string }
	stmts = append(stmts, fkeyStmts...)
	stmts = append(stmts, functionStmts...)
	stmts = append(stmts, viewStmts...)
	stmts = append(stmts, triggerStmts...)
	return stmts, nil
}

func MigrateTables(dialect string, gotSchema, wantSchema Schema, config Config) ([]string, error) {
	var stmts []string
	return stmts, nil
}

func DiffColumn(dialect string, gotColumn, wantColumn Column) ([]string, error) {
	return nil, nil
}

func DiffConstraint(dialect string, gotConstraint, wantConstraint Constraint) ([]string, error) {
	return nil, nil
}

func DiffIndex(dialect string, gotIndex, wantIndex Index) ([]string, error) {
	return nil, nil
}

func AutoMigrateContext(ctx context.Context, dialect string, db sq.Queryer, tables []sq.SchemaTable, config Config) error {
	gotMetadata, err := NewMetadataFromDB(dialect, db)
	if err != nil {
		return fmt.Errorf("error obtaining metadata from DB: %w", err)
	}
	wantMetadata, err := NewMetadataFromTables(dialect, tables)
	if err != nil {
		return fmt.Errorf("error obtaining metadata from tables: %w", err)
	}
	stmts, err := Diff(dialect, gotMetadata, wantMetadata, config)
	if err != nil {
		return fmt.Errorf("error when diffing the metadata: %w", err)
	}
	for _, stmt := range stmts {
		_, err = db.ExecContext(ctx, stmt)
		if err != nil {
			return fmt.Errorf("%s: %w", stmt, err)
		}
	}
	return nil
}
