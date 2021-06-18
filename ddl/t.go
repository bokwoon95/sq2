package ddl

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"

	"github.com/bokwoon95/sq"
)

func caller(skip int) (file string, line int) {
	var pc [1]uintptr
	// Skip two extra frames to account for this function
	// and runtime.Callers itself.
	n := runtime.Callers(skip+2, pc[:])
	if n == 0 {
		panic("ddl: zero callers found")
	}
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.File, frame.Line
}

func panicf(format string, a ...interface{}) {
	file, line := caller(2)
	panic(fmt.Errorf("%s:%d:%s", file, line, fmt.Sprintf(format, a...)))
}

type T struct {
	dialect string
	tbl     *Table
}

type TColumn struct {
	dialect     string
	tbl         *Table
	columnName  string
	columnIndex int
}

func (t *T) Column(field sq.Field) *TColumn {
	if field == nil {
		panicf("field is nil")
	}
	columnName := field.GetName()
	if columnName == "" {
		panicf("field has no name")
	}
	columnIndex := t.tbl.CachedColumnIndex(columnName)
	if columnIndex < 0 {
		panicf("table has no such column %s", columnName)
	}
	return &TColumn{
		dialect:     t.dialect,
		tbl:         t.tbl,
		columnName:  columnName,
		columnIndex: columnIndex,
	}
}

func (t *TColumn) Ignore() {
	t.tbl.Columns[t.columnIndex].Ignore = true
}

func (t *TColumn) Type(columnType string) *TColumn {
	t.tbl.Columns[t.columnIndex].ColumnType = columnType
	return t
}

func (t *TColumn) Config(config func(c *Column)) {
	col := t.tbl.Columns[t.columnIndex]
	config(&col)
	col.ColumnName = t.columnName
	t.tbl.Columns[t.columnIndex] = col
}

func sprintf(dialect, tableName string, format string, values []interface{}) (string, error) {
	if len(values) == 0 {
		return format, nil
	}
	str, err := appendSQLExclude(dialect, tableName, sq.Fieldf(format, values...))
	if err != nil {
		return "", err
	}
	return str, nil
}

func appendSQLExclude(dialect, tableName string, v sq.SQLExcludeAppender) (string, error) {
	buf := bufpool.Get().(*bytes.Buffer)
	args := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		args = args[:0]
		bufpool.Put(buf)
		argspool.Put(args)
	}()
	err := v.AppendSQLExclude(dialect, buf, &args, make(map[string][]int), []string{tableName})
	if err != nil {
		return "", err
	}
	if len(args) == 0 {
		return buf.String(), nil
	}
	str := sq.Sprintf(dialect, buf.String(), args)
	return str, nil
}

func (t *T) Sprintf(format string, values ...interface{}) string {
	expr, err := sprintf(t.dialect, t.tbl.TableName, format, values)
	if err != nil {
		panicf(err.Error())
	}
	return expr
}

func (t *TColumn) Generated(format string, values ...interface{}) *TColumn {
	expr, err := sprintf(t.dialect, t.tbl.TableName, format, values)
	if err != nil {
		panicf(err.Error())
	}
	t.tbl.Columns[t.columnIndex].GeneratedExpr = expr
	return t
}

func (t *TColumn) Stored() *TColumn {
	t.tbl.Columns[t.columnIndex].GeneratedExprStored = true
	return t
}

func (t *TColumn) Default(format string, values ...interface{}) *TColumn {
	expr, err := sprintf(t.dialect, t.tbl.TableName, format, values)
	if err != nil {
		panicf(err.Error())
	}
	t.tbl.Columns[t.columnIndex].ColumnDefault = expr
	return t
}

func (t *TColumn) Autoincrement() *TColumn {
	t.tbl.Columns[t.columnIndex].Autoincrement = true
	return t
}

func (t *TColumn) Identity() *TColumn {
	t.tbl.Columns[t.columnIndex].Identity = BY_DEFAULT_AS_IDENTITY
	return t
}

func (t *TColumn) AlwaysIdentity() *TColumn {
	t.tbl.Columns[t.columnIndex].Identity = ALWAYS_AS_IDENTITY
	return t
}

func (t *TColumn) OnUpdateCurrentTimestamp() *TColumn {
	t.tbl.Columns[t.columnIndex].OnUpdateCurrentTimestamp = true
	return t
}

func (t *TColumn) NotNull() *TColumn {
	t.tbl.Columns[t.columnIndex].IsNotNull = true
	return t
}

func (t *TColumn) PrimaryKey() *TColumn {
	constraintName := pgName(PRIMARY_KEY, t.tbl.TableName, t.columnName)
	if i := t.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		t.tbl.Constraints[i].ConstraintType = PRIMARY_KEY
		t.tbl.Constraints[i].TableName = t.tbl.TableName
		t.tbl.Constraints[i].Columns = []string{t.columnName}
	} else {
		t.tbl.AppendConstraint(Constraint{
			ConstraintSchema: t.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   PRIMARY_KEY,
			TableSchema:      t.tbl.TableSchema,
			TableName:        t.tbl.TableName,
			Columns:          []string{t.columnName},
		})
	}
	return t
}

func (t *TColumn) Unique() *TColumn {
	constraintName := pgName(UNIQUE, t.tbl.TableName, t.columnName)
	if i := t.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		t.tbl.Constraints[i].ConstraintType = UNIQUE
		t.tbl.Constraints[i].TableName = t.tbl.TableName
		t.tbl.Constraints[i].Columns = []string{t.columnName}
	} else {
		t.tbl.AppendConstraint(Constraint{
			ConstraintSchema: t.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   UNIQUE,
			TableSchema:      t.tbl.TableSchema,
			TableName:        t.tbl.TableName,
			Columns:          []string{t.columnName},
		})
	}
	return t
}

func (t *TColumn) Collate(collation string) *TColumn {
	t.tbl.Columns[t.columnIndex].CollationName = collation
	return t
}

type TConstraint struct {
	dialect         string
	tbl             *Table
	constraintName  string
	constraintIndex int
}

func (t *TConstraint) Config(config func(constraint *Constraint)) {
	constraint := t.tbl.Constraints[t.constraintIndex]
	config(&constraint)
	constraint.ConstraintName = t.constraintName
	t.tbl.Constraints[t.constraintIndex] = constraint
}

func (t *T) Check(constraintName string, format string, values ...interface{}) *TConstraint {
	expr, err := sprintf(t.dialect, t.tbl.TableName, format, values)
	if err != nil {
		panicf(err.Error())
	}
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	if i := t.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		t.tbl.Constraints[i].ConstraintType = CHECK
		t.tbl.Constraints[i].TableName = t.tbl.TableName
		t.tbl.Constraints[i].CheckExpr = expr
		tConstraint.constraintIndex = i
	} else {
		t.tbl.AppendConstraint(Constraint{
			ConstraintSchema: t.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   CHECK,
			TableSchema:      t.tbl.TableSchema,
			TableName:        t.tbl.TableName,
			CheckExpr:        expr,
		})
		tConstraint.constraintIndex = t.tbl.CachedConstraintIndex(constraintName)
	}
	if tConstraint.constraintIndex < 0 {
		panicf("could not create or update constraint '%s'", constraintName)
	}
	return tConstraint
}

func (t *T) Unique(fields ...sq.Field) *TConstraint {
	var columnNames []string
	for i, field := range fields {
		if field == nil {
			panicf("field at index %d is nil", i)
		}
		columnName := field.GetName()
		if columnName == "" {
			panicf("field at index %d has no name", i)
		}
		columnNames = append(columnNames, columnName)
	}
	constraintName := pgName(UNIQUE, t.tbl.TableName, columnNames...)
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	if i := t.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		t.tbl.Constraints[i].ConstraintType = UNIQUE
		t.tbl.Constraints[i].TableSchema = t.tbl.TableSchema
		t.tbl.Constraints[i].TableName = t.tbl.TableName
		t.tbl.Constraints[i].Columns = columnNames
		tConstraint.constraintIndex = i
	} else {
		t.tbl.AppendConstraint(Constraint{
			ConstraintSchema: t.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   UNIQUE,
			TableSchema:      t.tbl.TableSchema,
			TableName:        t.tbl.TableName,
			Columns:          columnNames,
		})
		tConstraint.constraintIndex = t.tbl.CachedConstraintIndex(constraintName)
	}
	if tConstraint.constraintIndex < 0 {
		panicf("could not create or update constraint '%s'", constraintName)
	}
	return tConstraint
}

func (t *T) PrimaryKey(fields ...sq.Field) *TConstraint {
	var columnNames []string
	for i, field := range fields {
		if field == nil {
			panicf("field at index %d is nil", i)
		}
		columnName := field.GetName()
		if columnName == "" {
			panicf("field at index %d has no name", i)
		}
		columnNames = append(columnNames, columnName)
	}
	constraintName := pgName(PRIMARY_KEY, t.tbl.TableName, columnNames...)
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	if i := t.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		t.tbl.Constraints[i].ConstraintType = PRIMARY_KEY
		t.tbl.Constraints[i].TableSchema = t.tbl.TableSchema
		t.tbl.Constraints[i].TableName = t.tbl.TableName
		t.tbl.Constraints[i].Columns = columnNames
		tConstraint.constraintIndex = i
	} else {
		t.tbl.AppendConstraint(Constraint{
			ConstraintSchema: t.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   PRIMARY_KEY,
			TableSchema:      t.tbl.TableSchema,
			TableName:        t.tbl.TableName,
			Columns:          columnNames,
		})
		tConstraint.constraintIndex = t.tbl.CachedConstraintIndex(constraintName)
	}
	if tConstraint.constraintIndex < 0 {
		panicf("could not create or update constraint '%s'", constraintName)
	}
	return tConstraint
}

func (t *T) NameUnique(constraintName string, fields ...sq.Field) *TConstraint {
	var columnNames []string
	for i, field := range fields {
		if field == nil {
			panicf("field at index %d is nil", i)
		}
		columnName := field.GetName()
		if columnName == "" {
			panicf("field at index %d has no name", i)
		}
		columnNames = append(columnNames, columnName)
	}
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	if i := t.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		t.tbl.Constraints[i].ConstraintType = UNIQUE
		t.tbl.Constraints[i].TableSchema = t.tbl.TableSchema
		t.tbl.Constraints[i].TableName = t.tbl.TableName
		t.tbl.Constraints[i].Columns = columnNames
		tConstraint.constraintIndex = i
	} else {
		t.tbl.AppendConstraint(Constraint{
			ConstraintSchema: t.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   UNIQUE,
			TableSchema:      t.tbl.TableSchema,
			TableName:        t.tbl.TableName,
			Columns:          columnNames,
		})
		tConstraint.constraintIndex = t.tbl.CachedConstraintIndex(constraintName)
	}
	if tConstraint.constraintIndex < 0 {
		panicf("could not create or update constraint '%s'", constraintName)
	}
	return tConstraint
}

func (t *T) NamePrimaryKey(constraintName string, fields ...sq.Field) *TConstraint {
	var columnNames []string
	for i, field := range fields {
		if field == nil {
			panicf("field at index %d is nil", i)
		}
		columnName := field.GetName()
		if columnName == "" {
			panicf("field at index %d has no name", i)
		}
		columnNames = append(columnNames, columnName)
	}
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	if i := t.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		t.tbl.Constraints[i].ConstraintType = PRIMARY_KEY
		t.tbl.Constraints[i].TableSchema = t.tbl.TableSchema
		t.tbl.Constraints[i].TableName = t.tbl.TableName
		t.tbl.Constraints[i].Columns = columnNames
		tConstraint.constraintIndex = i
	} else {
		t.tbl.AppendConstraint(Constraint{
			ConstraintSchema: t.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   PRIMARY_KEY,
			TableSchema:      t.tbl.TableSchema,
			TableName:        t.tbl.TableName,
			Columns:          columnNames,
		})
		tConstraint.constraintIndex = t.tbl.CachedConstraintIndex(constraintName)
	}
	if tConstraint.constraintIndex < 0 {
		panicf("could not create or update constraint '%s'", constraintName)
	}
	return tConstraint
}

type TIndex struct {
	dialect    string
	tbl        *Table
	indexName  string
	indexIndex int
}

func (t *T) Index(fields ...sq.Field) *TIndex {
	var columnNames []string
	var exprs []string
	for i, field := range fields {
		if field == nil {
			panicf("field at index %d is nil", i)
		}
		var columnName, expr string
		columnName = field.GetName()
		if columnName == "" {
			var err error
			expr, err = appendSQLExclude(t.dialect, t.tbl.TableName, field)
			if err != nil {
				panicf(err.Error())
			}
		}
		columnNames = append(columnNames, columnName)
		exprs = append(exprs, expr)
	}
	indexName := pgName(INDEX, t.tbl.TableName, columnNames...)
	tindex := &TIndex{
		dialect:   t.dialect,
		tbl:       t.tbl,
		indexName: indexName,
	}
	if i := t.tbl.CachedIndexIndex(indexName); i >= 0 {
		t.tbl.Indices[i].TableSchema = t.tbl.TableSchema
		t.tbl.Indices[i].TableName = t.tbl.TableName
		t.tbl.Indices[i].Columns = columnNames
		t.tbl.Indices[i].Exprs = exprs
		tindex.indexIndex = i
	} else {
		t.tbl.AppendIndex(Index{
			IndexSchema: t.tbl.TableSchema,
			IndexName:   indexName,
			IndexType:   "BTREE",
			IsUnique:    false,
			TableSchema: t.tbl.TableSchema,
			TableName:   t.tbl.TableName,
			Columns:     columnNames,
			Exprs:       exprs,
		})
		tindex.indexIndex = t.tbl.CachedIndexIndex(indexName)
	}
	if tindex.indexIndex < 0 {
		panicf("could not create or update index '%s'", indexName)
	}
	return tindex
}

func (t *T) NameIndex(indexName string, fields ...sq.Field) *TIndex {
	var columnNames []string
	var exprs []string
	for i, field := range fields {
		if field == nil {
			panicf("field at index %d is nil", i)
		}
		var columnName, expr string
		columnName = field.GetName()
		if columnName == "" {
			var err error
			expr, err = appendSQLExclude(t.dialect, t.tbl.TableName, field)
			if err != nil {
				panicf(err.Error())
			}
		}
		columnNames = append(columnNames, columnName)
		exprs = append(exprs, expr)
	}
	tIndex := &TIndex{
		dialect:   t.dialect,
		tbl:       t.tbl,
		indexName: indexName,
	}
	if i := t.tbl.CachedIndexIndex(indexName); i >= 0 {
		t.tbl.Indices[i].TableSchema = t.tbl.TableSchema
		t.tbl.Indices[i].TableName = t.tbl.TableName
		t.tbl.Indices[i].Columns = columnNames
		t.tbl.Indices[i].Exprs = exprs
		tIndex.indexIndex = i
	} else {
		t.tbl.AppendIndex(Index{
			IndexSchema: t.tbl.TableSchema,
			IndexName:   indexName,
			IndexType:   "BTREE",
			IsUnique:    false,
			TableSchema: t.tbl.TableSchema,
			TableName:   t.tbl.TableName,
			Columns:     columnNames,
			Exprs:       exprs,
		})
		tIndex.indexIndex = t.tbl.CachedIndexIndex(indexName)
	}
	if tIndex.indexIndex < 0 {
		panicf("could not create or update index '%s'", indexName)
	}
	return tIndex
}

func (t *TIndex) Unique() *TIndex {
	t.tbl.Indices[t.indexIndex].IsUnique = true
	return t
}

func (t *TIndex) Using(indexType string) *TIndex {
	t.tbl.Indices[t.indexIndex].IndexType = strings.ToUpper(indexType)
	return t
}

func (t *TIndex) Where(format string, values ...interface{}) *TIndex {
	expr, err := sprintf(t.dialect, t.tbl.TableName, format, values)
	if err != nil {
		panicf(err.Error())
	}
	t.tbl.Indices[t.indexIndex].Predicate = expr
	return t
}

func (t *TIndex) Include(fields ...sq.Field) *TIndex {
	var columnNames []string
	for i, field := range fields {
		if field == nil {
			panicf("field at index %d is nil", i)
		}
		columnName := field.GetName()
		if columnName == "" {
			panicf("field at index %d has no name", i)
		}
		columnNames = append(columnNames, columnName)
	}
	t.tbl.Indices[t.indexIndex].Include = columnNames
	return t
}

func (t *TIndex) Config(config func(index *Index)) {
	index := t.tbl.Indices[t.indexIndex]
	config(&index)
	index.IndexName = t.indexName
	t.tbl.Indices[t.indexIndex] = index
}
