package ddl

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
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

func (t *T) VirtualTable(moduleName string, moduleArgs ...string) {
	t.tbl.VirtualTable = moduleName
	t.tbl.VirtualTableArgs = moduleArgs
}

type TColumn struct {
	dialect     string
	tbl         *Table
	columnName  string
	columnIndex int
}

func (tbl *Table) tcol(dialect, columnName string) *TColumn {
	return &TColumn{
		dialect:     dialect,
		tbl:         tbl,
		columnName:  columnName,
		columnIndex: tbl.CachedColumnIndex(columnName),
	}
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
	column := t.tbl.Columns[t.columnIndex]
	config(&column)
	column.TableSchema = t.tbl.TableSchema
	column.TableName = t.tbl.TableName
	column.ColumnName = t.columnName
	t.tbl.Columns[t.columnIndex] = column
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
	str, err := sq.Sprintf(dialect, buf.String(), args)
	if err != nil {
		return "", err
	}
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
	var expr string
	if len(values) == 0 {
		if len(format) >= 2 && format[0] == '\'' && format[len(format)-1] == '\'' {
			expr = format
		} else if strings.EqualFold(format, "TRUE") ||
			strings.EqualFold(format, "FALSE") ||
			strings.EqualFold(format, "CURRENT_DATE") ||
			strings.EqualFold(format, "CURRENT_TIME") ||
			strings.EqualFold(format, "CURRENT_TIMESTAMP") {
			expr = format
		} else if _, err := strconv.ParseInt(format, 10, 64); err == nil {
			expr = format
		} else if _, err := strconv.ParseFloat(format, 64); err == nil {
			expr = format
		} else if t.dialect == sq.DialectPostgres {
			expr = format
		} else {
			expr = "(" + format + ")"
		}
		t.tbl.Columns[t.columnIndex].ColumnDefault = expr
		return t
	}
	expr, err := sprintf(t.dialect, t.tbl.TableName, format, values)
	if err != nil {
		panicf(err.Error())
	}
	if t.dialect != sq.DialectPostgres {
		expr = "(" + expr + ")"
	}
	t.tbl.Columns[t.columnIndex].ColumnDefault = expr
	return t
}

func (t *TColumn) DefaultLiteral(value interface{}) *TColumn {
	literal, err := sq.Sprint(value)
	if err != nil {
		panicf(err.Error())
	}
	t.tbl.Columns[t.columnIndex].ColumnDefault = literal
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
	constraintName := generateName(PRIMARY_KEY, t.tbl.TableName, t.columnName)
	_, err := createOrUpdateConstraint(t.tbl, PRIMARY_KEY, constraintName, []string{t.columnName}, "")
	if err != nil {
		panicf(err.Error())
	}
	t.tbl.Columns[t.columnIndex].IsPrimaryKey = true
	return t
}

func (t *TColumn) Unique() *TColumn {
	constraintName := generateName(UNIQUE, t.tbl.TableName, t.columnName)
	_, err := createOrUpdateConstraint(t.tbl, UNIQUE, constraintName, []string{t.columnName}, "")
	if err != nil {
		panicf(err.Error())
	}
	t.tbl.Columns[t.columnIndex].IsUnique = true
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

func getColumnNames(fields []sq.Field) ([]string, error) {
	var columnNames []string
	for i, field := range fields {
		if field == nil {
			return nil, fmt.Errorf("ddl: field #%d is nil", i+1)
		}
		columnName := field.GetName()
		if columnName == "" {
			return nil, fmt.Errorf("ddl: field #%d has no name", i+1)
		}
		columnNames = append(columnNames, columnName)
	}
	return columnNames, nil
}

func createOrUpdateConstraint(tbl *Table, constraintType, constraintName string, columns []string, checkExpr string) (constraintIndex int, err error) {
	if constraintName == "" {
		return -1, fmt.Errorf("constraintName cannot be empty")
	}
	if constraintIndex = tbl.CachedConstraintIndex(constraintName); constraintIndex >= 0 {
		constraint := tbl.Constraints[constraintIndex]
		constraint.ConstraintType = constraintType
		constraint.TableName = tbl.TableName
		constraint.Columns = columns
		constraint.CheckExpr = checkExpr
		tbl.Constraints[constraintIndex] = constraint
	} else {
		constraintIndex = tbl.AppendConstraint(Constraint{
			ConstraintSchema: tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   constraintType,
			TableSchema:      tbl.TableSchema,
			TableName:        tbl.TableName,
			Columns:          columns,
			CheckExpr:        checkExpr,
		})
	}
	return constraintIndex, nil
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
	tConstraint.constraintIndex, err = createOrUpdateConstraint(t.tbl, CHECK, constraintName, nil, expr)
	if err != nil {
		panicf(err.Error())
	}
	return tConstraint
}

func (t *T) Unique(fields ...sq.Field) *TConstraint {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicf(err.Error())
	}
	constraintName := generateName(UNIQUE, t.tbl.TableName, columnNames...)
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintIndex, err = createOrUpdateConstraint(t.tbl, UNIQUE, constraintName, columnNames, "")
	if err != nil {
		panicf(err.Error())
	}
	return tConstraint
}

func (t *T) PrimaryKey(fields ...sq.Field) *TConstraint {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicf(err.Error())
	}
	constraintName := generateName(PRIMARY_KEY, t.tbl.TableName, columnNames...)
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintIndex, err = createOrUpdateConstraint(t.tbl, PRIMARY_KEY, constraintName, columnNames, "")
	if err != nil {
		panicf(err.Error())
	}
	return tConstraint
}

func (t *T) ForeignKey(fields ...sq.Field) *TConstraint {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicf(err.Error())
	}
	constraintName := generateName(FOREIGN_KEY, t.tbl.TableName, columnNames...)
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintIndex, err = createOrUpdateConstraint(t.tbl, FOREIGN_KEY, constraintName, columnNames, "")
	if err != nil {
		panicf(err.Error())
	}
	return tConstraint
}

func (t *T) NameUnique(constraintName string, fields ...sq.Field) *TConstraint {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicf(err.Error())
	}
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintIndex, err = createOrUpdateConstraint(t.tbl, UNIQUE, constraintName, columnNames, "")
	if err != nil {
		panicf(err.Error())
	}
	return tConstraint
}

func (t *T) NamePrimaryKey(constraintName string, fields ...sq.Field) *TConstraint {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicf(err.Error())
	}
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintIndex, err = createOrUpdateConstraint(t.tbl, PRIMARY_KEY, constraintName, columnNames, "")
	if err != nil {
		panicf(err.Error())
	}
	return tConstraint
}

func (t *T) NameForeignKey(constraintName string, fields ...sq.Field) *TConstraint {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicf(err.Error())
	}
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintIndex, err = createOrUpdateConstraint(t.tbl, FOREIGN_KEY, constraintName, columnNames, "")
	if err != nil {
		panicf(err.Error())
	}
	return tConstraint
}

func (t *TConstraint) Config(config func(constraint *Constraint)) {
	constraint := t.tbl.Constraints[t.constraintIndex]
	config(&constraint)
	constraint.TableSchema = t.tbl.TableSchema
	constraint.TableName = t.tbl.TableName
	constraint.ConstraintName = t.constraintName
	t.tbl.Constraints[t.constraintIndex] = constraint
}

func (t *TConstraint) References(table sq.Table, fields ...sq.Field) *TConstraint {
	if table == nil {
		panicf("referenced table is nil")
	}
	referencesTable := table.GetName()
	if referencesTable == "" {
		panicf("referenced table has no name")
	}
	referencesColumns, err := getColumnNames(fields)
	if err != nil {
		panicf("referenced " + err.Error())
	}
	constraint := t.tbl.Constraints[t.constraintIndex]
	constraint.ReferencesTable = referencesTable
	constraint.ReferencesColumns = referencesColumns
	t.tbl.Constraints[t.constraintIndex] = constraint
	return t
}

func (t *TConstraint) OnUpdate(action string) *TConstraint {
	constraint := t.tbl.Constraints[t.constraintIndex]
	constraint.OnUpdate = action
	t.tbl.Constraints[t.constraintIndex] = constraint
	return t
}

func (t *TConstraint) OnDelete(action string) *TConstraint {
	constraint := t.tbl.Constraints[t.constraintIndex]
	constraint.OnDelete = action
	t.tbl.Constraints[t.constraintIndex] = constraint
	return t
}

func (t *TConstraint) Deferrable() *TConstraint {
	constraint := t.tbl.Constraints[t.constraintIndex]
	constraint.IsDeferrable = true
	t.tbl.Constraints[t.constraintIndex] = constraint
	return t
}

func (t *TConstraint) IsInitiallyDeferred() *TConstraint {
	constraint := t.tbl.Constraints[t.constraintIndex]
	constraint.IsInitiallyDeferred = true
	t.tbl.Constraints[t.constraintIndex] = constraint
	return t
}

type TIndex struct {
	dialect    string
	tbl        *Table
	indexName  string
	indexIndex int
}

func getColumnNamesAndExprs(dialect, tableName string, fields []sq.Field) (columnNames, exprs []string, err error) {
	for i, field := range fields {
		if field == nil {
			return nil, nil, fmt.Errorf("field #%d is nil", i+1)
		}
		var columnName, expr string
		columnName = field.GetName()
		if columnName == "" {
			if fieldliteral, ok := field.(sq.FieldLiteral); ok {
				expr = string(fieldliteral)
			} else {
				var err error
				expr, err = appendSQLExclude(dialect, tableName, field)
				if err != nil {
					return nil, nil, err
				}
				expr = "(" + expr + ")"
			}
		}
		columnNames = append(columnNames, columnName)
		exprs = append(exprs, expr)
	}
	return columnNames, exprs, nil
}

func createOrUpdateIndex(tbl *Table, indexName string, columns []string, exprs []string) (indexIndex int, err error) {
	if indexName == "" {
		return -1, fmt.Errorf("indexName cannot be empty")
	}
	if indexIndex = tbl.CachedIndexIndex(indexName); indexIndex >= 0 {
		index := tbl.Indices[indexIndex]
		index.TableSchema = tbl.TableSchema
		index.TableName = tbl.TableName
		index.Columns = columns
		index.Exprs = exprs
		tbl.Indices[indexIndex] = index
	} else {
		indexIndex = tbl.AppendIndex(Index{
			IndexSchema: tbl.TableSchema,
			IndexName:   indexName,
			IndexType:   "BTREE",
			TableSchema: tbl.TableSchema,
			TableName:   tbl.TableName,
			Columns:     columns,
			Exprs:       exprs,
		})
	}
	return indexIndex, nil
}

func (t *T) Index(fields ...sq.Field) *TIndex {
	columnNames, exprs, err := getColumnNamesAndExprs(t.dialect, t.tbl.TableName, fields)
	if err != nil {
		panicf(err.Error())
	}
	indexName := generateName(INDEX, t.tbl.TableName, columnNames...)
	tIndex := &TIndex{
		dialect:   t.dialect,
		tbl:       t.tbl,
		indexName: indexName,
	}
	tIndex.indexIndex, err = createOrUpdateIndex(t.tbl, indexName, columnNames, exprs)
	if err != nil {
		panicf(err.Error())
	}
	return tIndex
}

func (t *T) NameIndex(indexName string, fields ...sq.Field) *TIndex {
	columnNames, exprs, err := getColumnNamesAndExprs(t.dialect, t.tbl.TableName, fields)
	if err != nil {
		panicf(err.Error())
	}
	tIndex := &TIndex{
		dialect:   t.dialect,
		tbl:       t.tbl,
		indexName: indexName,
	}
	tIndex.indexIndex, err = createOrUpdateIndex(t.tbl, indexName, columnNames, exprs)
	if err != nil {
		panicf(err.Error())
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
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicf(err.Error())
	}
	t.tbl.Indices[t.indexIndex].Include = columnNames
	return t
}

func (t *TIndex) Config(config func(index *Index)) {
	index := t.tbl.Indices[t.indexIndex]
	config(&index)
	index.TableSchema = t.tbl.TableSchema
	index.TableName = t.tbl.TableName
	index.IndexName = t.indexName
	t.tbl.Indices[t.indexIndex] = index
}
