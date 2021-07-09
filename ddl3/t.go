package ddl3

import (
	"bytes"
	"fmt"
	"io/fs"
	"runtime"
	"strconv"
	"strings"

	"github.com/bokwoon95/sq"
)

func panicErr(err error) {
	_, file, line, _ := runtime.Caller(2)
	panic(fmt.Errorf("%s:%d:%w", file, line, err))
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
	dialect        string
	tbl            *Table
	columnName     string
	columnPosition int
}

func (tbl *Table) tcol(dialect, columnName string) *TColumn {
	return &TColumn{
		dialect:        dialect,
		tbl:            tbl,
		columnName:     columnName,
		columnPosition: tbl.CachedColumnPosition(columnName),
	}
}

func (t *T) Column(field sq.Field) *TColumn {
	if field == nil {
		panicErr(fmt.Errorf("Column: field is nil"))
	}
	columnName := field.GetName()
	if columnName == "" {
		panicErr(fmt.Errorf("Column: field has no name"))
	}
	columnPosition := t.tbl.CachedColumnPosition(columnName)
	if columnPosition < 0 {
		panicErr(fmt.Errorf("Column: table has no such column %s", columnName))
	}
	return &TColumn{
		dialect:        t.dialect,
		tbl:            t.tbl,
		columnName:     columnName,
		columnPosition: columnPosition,
	}
}

func (t *TColumn) Ignore() {
	t.tbl.Columns[t.columnPosition].Ignore = true
}

func (t *TColumn) Type(columnType string) *TColumn {
	t.tbl.Columns[t.columnPosition].ColumnType = columnType
	return t
}

func (t *TColumn) Config(config func(c *Column)) {
	column := t.tbl.Columns[t.columnPosition]
	config(&column)
	column.TableSchema = t.tbl.TableSchema
	column.TableName = t.tbl.TableName
	column.ColumnName = t.columnName
	t.tbl.Columns[t.columnPosition] = column
}

func sprintf(dialect, tableName string, format string, values []interface{}) (string, error) {
	if len(values) == 0 {
		return format, nil
	}
	// TODO: should I use sq.BufferPrintf here directly instead?
	str, err := appendSQLExclude(dialect, tableName, sq.Fieldf(format, values...))
	if err != nil {
		return "", err
	}
	return str, nil
}

func appendSQLExclude(dialect, tableName string, v sq.SQLExcludeAppender) (string, error) {
	query, args, _, err := sq.ToSQLExclude(dialect, v, []string{tableName})
	if err != nil {
		return "", err
	}
	if len(args) == 0 {
		return query, nil
	}
	query, err = sq.Sprintf(dialect, query, args)
	if err != nil {
		return "", err
	}
	return query, nil
}

func (t *T) Sprintf(format string, values ...interface{}) string {
	expr, err := sprintf(t.dialect, t.tbl.TableName, format, values)
	if err != nil {
		panicErr(fmt.Errorf("Sprintf: %w", err))
	}
	return expr
}

func (t *TColumn) Generated(format string, values ...interface{}) *TColumn {
	expr, err := sprintf(t.dialect, t.tbl.TableName, format, values)
	if err != nil {
		panicErr(fmt.Errorf("Generated: %w", err))
	}
	t.tbl.Columns[t.columnPosition].GeneratedExpr = expr
	return t
}

func (t *TColumn) Stored() *TColumn {
	t.tbl.Columns[t.columnPosition].GeneratedExprStored = true
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
		t.tbl.Columns[t.columnPosition].ColumnDefault = expr
		return t
	}
	expr, err := sprintf(t.dialect, t.tbl.TableName, format, values)
	if err != nil {
		panicErr(fmt.Errorf("Default: %w", err))
	}
	if t.dialect != sq.DialectPostgres {
		expr = "(" + expr + ")"
	}
	t.tbl.Columns[t.columnPosition].ColumnDefault = expr
	return t
}

func (t *TColumn) Autoincrement() *TColumn {
	t.tbl.Columns[t.columnPosition].Autoincrement = true
	return t
}

func (t *TColumn) Identity() *TColumn {
	t.tbl.Columns[t.columnPosition].Identity = BY_DEFAULT_AS_IDENTITY
	return t
}

func (t *TColumn) AlwaysIdentity() *TColumn {
	t.tbl.Columns[t.columnPosition].Identity = ALWAYS_AS_IDENTITY
	return t
}

func (t *TColumn) OnUpdateCurrentTimestamp() *TColumn {
	t.tbl.Columns[t.columnPosition].OnUpdateCurrentTimestamp = true
	return t
}

func (t *TColumn) NotNull() *TColumn {
	t.tbl.Columns[t.columnPosition].IsNotNull = true
	return t
}

func (t *TColumn) PrimaryKey() *TColumn {
	constraintName := generateName(PRIMARY_KEY, t.tbl.TableName, t.columnName)
	_, err := createOrUpdateConstraint(t.tbl, PRIMARY_KEY, constraintName, []string{t.columnName}, "")
	if err != nil {
		panicErr(fmt.Errorf("PrimaryKey: %w", err))
	}
	return t
}

func (t *TColumn) Unique() *TColumn {
	constraintName := generateName(UNIQUE, t.tbl.TableName, t.columnName)
	_, err := createOrUpdateConstraint(t.tbl, UNIQUE, constraintName, []string{t.columnName}, "")
	if err != nil {
		panicErr(fmt.Errorf("Unique: %w", err))
	}
	return t
}

func (t *TColumn) Collate(collation string) *TColumn {
	t.tbl.Columns[t.columnPosition].CollationName = collation
	return t
}

type TConstraint struct {
	dialect            string
	tbl                *Table
	constraintName     string
	constraintPosition int
}

func getColumnNames(fields []sq.Field) ([]string, error) {
	var columnNames []string
	for i, field := range fields {
		if field == nil {
			return nil, fmt.Errorf("field #%d is nil", i+1)
		}
		columnName := field.GetName()
		if columnName == "" {
			return nil, fmt.Errorf("field #%d has no name", i+1)
		}
		columnNames = append(columnNames, columnName)
	}
	return columnNames, nil
}

func createOrUpdateConstraint(tbl *Table, constraintType, constraintName string, columns []string, checkExpr string) (constraintPosition int, err error) {
	if constraintName == "" {
		return -1, fmt.Errorf("constraintName cannot be empty")
	}
	if constraintPosition = tbl.CachedConstraintPosition(constraintName); constraintPosition >= 0 {
		constraint := tbl.Constraints[constraintPosition]
		constraint.TableSchema = tbl.TableSchema
		constraint.TableName = tbl.TableName
		constraint.ConstraintType = constraintType
		constraint.Columns = columns
		constraint.CheckExpr = checkExpr
		tbl.Constraints[constraintPosition] = constraint
	} else {
		constraintPosition = tbl.AppendConstraint(Constraint{
			TableSchema:    tbl.TableSchema,
			TableName:      tbl.TableName,
			ConstraintName: constraintName,
			ConstraintType: constraintType,
			Columns:        columns,
			CheckExpr:      checkExpr,
		})
	}
	return constraintPosition, nil
}

func (t *T) Check(constraintName string, format string, values ...interface{}) *TConstraint {
	expr, err := sprintf(t.dialect, t.tbl.TableName, format, values)
	if err != nil {
		panicErr(fmt.Errorf("Check: %w", err))
	}
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.tbl, CHECK, constraintName, nil, expr)
	if err != nil {
		panicErr(fmt.Errorf("Check: %w", err))
	}
	return tConstraint
}

func (t *T) Unique(fields ...sq.Field) *TConstraint {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicErr(fmt.Errorf("Unique: %w", err))
	}
	constraintName := generateName(UNIQUE, t.tbl.TableName, columnNames...)
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.tbl, UNIQUE, constraintName, columnNames, "")
	if err != nil {
		panicErr(fmt.Errorf("Unique: %w", err))
	}
	return tConstraint
}

func (t *T) PrimaryKey(fields ...sq.Field) *TConstraint {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicErr(fmt.Errorf("PrimaryKey: %w", err))
	}
	constraintName := generateName(PRIMARY_KEY, t.tbl.TableName, columnNames...)
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.tbl, PRIMARY_KEY, constraintName, columnNames, "")
	if err != nil {
		panicErr(fmt.Errorf("PrimaryKey: %w", err))
	}
	return tConstraint
}

func (t *T) ForeignKey(fields ...sq.Field) *TConstraint {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicErr(fmt.Errorf("ForeignKey: %w", err))
	}
	constraintName := generateName(FOREIGN_KEY, t.tbl.TableName, columnNames...)
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.tbl, FOREIGN_KEY, constraintName, columnNames, "")
	if err != nil {
		panicErr(fmt.Errorf("ForeignKey: %w", err))
	}
	return tConstraint
}

func (t *T) NameUnique(constraintName string, fields ...sq.Field) *TConstraint {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicErr(fmt.Errorf("NameUnique: %w", err))
	}
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.tbl, UNIQUE, constraintName, columnNames, "")
	if err != nil {
		panicErr(fmt.Errorf("NameUnique: %w", err))
	}
	return tConstraint
}

func (t *T) NamePrimaryKey(constraintName string, fields ...sq.Field) *TConstraint {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicErr(fmt.Errorf("NamePrimaryKey: %w", err))
	}
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.tbl, PRIMARY_KEY, constraintName, columnNames, "")
	if err != nil {
		panicErr(fmt.Errorf("NamePrimaryKey: %w", err))
	}
	return tConstraint
}

func (t *T) NameForeignKey(constraintName string, fields ...sq.Field) *TConstraint {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicErr(fmt.Errorf("NameForeignKey: %w", err))
	}
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.tbl, FOREIGN_KEY, constraintName, columnNames, "")
	if err != nil {
		panicErr(fmt.Errorf("NameForeignKey: %w", err))
	}
	return tConstraint
}

func (t *TConstraint) Config(config func(constraint *Constraint)) {
	constraint := t.tbl.Constraints[t.constraintPosition]
	config(&constraint)
	constraint.TableSchema = t.tbl.TableSchema
	constraint.TableName = t.tbl.TableName
	constraint.ConstraintName = t.constraintName
	t.tbl.Constraints[t.constraintPosition] = constraint
}

func (t *TConstraint) References(table sq.Table, fields ...sq.Field) *TConstraint {
	if table == nil {
		panicErr(fmt.Errorf("References: referenced table is nil"))
	}
	referencesTable := table.GetName()
	if referencesTable == "" {
		panicErr(fmt.Errorf("References: referenced table has no name"))
	}
	referencesColumns, err := getColumnNames(fields)
	if err != nil {
		panicErr(fmt.Errorf("References: %w", err))
	}
	constraint := t.tbl.Constraints[t.constraintPosition]
	constraint.ReferencesTable = referencesTable
	constraint.ReferencesColumns = referencesColumns
	t.tbl.Constraints[t.constraintPosition] = constraint
	return t
}

func (t *TConstraint) OnUpdate(action string) *TConstraint {
	constraint := t.tbl.Constraints[t.constraintPosition]
	constraint.OnUpdate = action
	t.tbl.Constraints[t.constraintPosition] = constraint
	return t
}

func (t *TConstraint) OnDelete(action string) *TConstraint {
	constraint := t.tbl.Constraints[t.constraintPosition]
	constraint.OnDelete = action
	t.tbl.Constraints[t.constraintPosition] = constraint
	return t
}

func (t *TConstraint) Deferrable() *TConstraint {
	constraint := t.tbl.Constraints[t.constraintPosition]
	constraint.IsDeferrable = true
	t.tbl.Constraints[t.constraintPosition] = constraint
	return t
}

func (t *TConstraint) InitiallyDeferred() *TConstraint {
	constraint := t.tbl.Constraints[t.constraintPosition]
	constraint.IsInitiallyDeferred = true
	t.tbl.Constraints[t.constraintPosition] = constraint
	return t
}

type TIndex struct {
	dialect       string
	tbl           *Table
	indexName     string
	indexPosition int
}

func getColumnNamesAndExprs(dialect, tableName string, fields []sq.Field) (columnNames, exprs []string, err error) {
	for i, field := range fields {
		if field == nil {
			return nil, nil, fmt.Errorf("field #%d is nil", i+1)
		}
		var columnName, expr string
		columnName = field.GetName()
		if columnName == "" {
			if f, ok := field.(sq.FieldLiteral); ok {
				expr = f.GetName()
			} else {
				var err error
				expr, err = appendSQLExclude(dialect, tableName, field)
				if err != nil {
					return nil, nil, fmt.Errorf("field #%d, :%w", i+1, err)
				}
				expr = "(" + expr + ")"
			}
		}
		columnNames = append(columnNames, columnName)
		exprs = append(exprs, expr)
	}
	return columnNames, exprs, nil
}

func createOrUpdateIndex(tbl *Table, indexName string, columns []string, exprs []string) (indexPosition int, err error) {
	if indexName == "" {
		return -1, fmt.Errorf("indexName cannot be empty")
	}
	if indexPosition = tbl.CachedIndexPosition(indexName); indexPosition >= 0 {
		index := tbl.Indexes[indexPosition]
		index.TableSchema = tbl.TableSchema
		index.TableName = tbl.TableName
		index.Columns = columns
		index.Exprs = exprs
		tbl.Indexes[indexPosition] = index
	} else {
		indexPosition = tbl.AppendIndex(Index{
			TableSchema: tbl.TableSchema,
			TableName:   tbl.TableName,
			IndexName:   indexName,
			IndexType:   "BTREE",
			Columns:     columns,
			Exprs:       exprs,
		})
	}
	return indexPosition, nil
}

func (t *T) Index(fields ...sq.Field) *TIndex {
	columnNames, exprs, err := getColumnNamesAndExprs(t.dialect, t.tbl.TableName, fields)
	if err != nil {
		panicErr(fmt.Errorf("Index: %w", err))
	}
	indexName := generateName(INDEX, t.tbl.TableName, columnNames...)
	tIndex := &TIndex{
		dialect:   t.dialect,
		tbl:       t.tbl,
		indexName: indexName,
	}
	tIndex.indexPosition, err = createOrUpdateIndex(t.tbl, indexName, columnNames, exprs)
	if err != nil {
		panicErr(fmt.Errorf("Index: %w", err))
	}
	return tIndex
}

func (t *T) NameIndex(indexName string, fields ...sq.Field) *TIndex {
	columnNames, exprs, err := getColumnNamesAndExprs(t.dialect, t.tbl.TableName, fields)
	if err != nil {
		panicErr(fmt.Errorf("NameIndex: %w", err))
	}
	tIndex := &TIndex{
		dialect:   t.dialect,
		tbl:       t.tbl,
		indexName: indexName,
	}
	tIndex.indexPosition, err = createOrUpdateIndex(t.tbl, indexName, columnNames, exprs)
	if err != nil {
		panicErr(fmt.Errorf("NameIndex: %w", err))
	}
	return tIndex
}

func (t *TIndex) Unique() *TIndex {
	t.tbl.Indexes[t.indexPosition].IsUnique = true
	return t
}

func (t *TIndex) Using(indexType string) *TIndex {
	t.tbl.Indexes[t.indexPosition].IndexType = strings.ToUpper(indexType)
	return t
}

func (t *TIndex) Where(format string, values ...interface{}) *TIndex {
	expr, err := sprintf(t.dialect, t.tbl.TableName, format, values)
	if err != nil {
		panicErr(fmt.Errorf("Where: %w", err))
	}
	t.tbl.Indexes[t.indexPosition].Where = expr
	return t
}

func (t *TIndex) Include(fields ...sq.Field) *TIndex {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicErr(fmt.Errorf("Include: %w", err))
	}
	t.tbl.Indexes[t.indexPosition].Include = columnNames
	return t
}

func (t *TIndex) Config(config func(index *Index)) {
	index := t.tbl.Indexes[t.indexPosition]
	config(&index)
	index.TableSchema = t.tbl.TableSchema
	index.TableName = t.tbl.TableName
	index.IndexName = t.indexName
	t.tbl.Indexes[t.indexPosition] = index
}

type TTrigger struct {
	dialect         string
	tbl             *Table
	triggerName     string
	triggerPosition int
}

func (t *T) Trigger(sql string) {
	tableSchema, tableName, triggerName, err := getTriggerInfo(sql)
	if err != nil {
		panicErr(fmt.Errorf("Trigger: %w", err))
	}
	if tableSchema == "" {
		tableSchema = t.tbl.TableSchema
	}
	if tableSchema != t.tbl.TableSchema {
		panicErr(fmt.Errorf("Trigger: table schema does not match (got=%s, want=%s)", t.tbl.TableSchema, tableSchema))
	}
	if tableName != t.tbl.TableName {
		panicErr(fmt.Errorf("Trigger: table name does not match (got=%s, want=%s)", t.tbl.TableName, tableName))
	}
	triggerPosition := t.tbl.CachedTriggerPosition(triggerName)
	if triggerPosition < 0 {
		t.tbl.AppendTrigger(Trigger{
			TableSchema: t.tbl.TableSchema,
			TableName:   t.tbl.TableName,
			TriggerName: triggerName,
			SQL:         sql,
		})
	} else {
		t.tbl.Triggers[triggerPosition].SQL = sql
	}
}

func (t *T) TriggerFile(fsys fs.FS, name string) {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		panicErr(fmt.Errorf("TriggerFile: %w", err))
	}
	sql := string(b)
	tableSchema, tableName, triggerName, err := getTriggerInfo(sql)
	if err != nil {
		panicErr(fmt.Errorf("TriggerFile: %w", err))
	}
	if tableSchema == "" {
		tableSchema = t.tbl.TableSchema
	}
	if tableSchema != t.tbl.TableSchema {
		panicErr(fmt.Errorf("TriggerFile: table schema does not match (got=%s, want=%s)", t.tbl.TableSchema, tableSchema))
	}
	if tableName != t.tbl.TableName {
		panicErr(fmt.Errorf("TriggerFile: table name does not match (got=%s, want=%s)", t.tbl.TableName, tableName))
	}
	triggerPosition := t.tbl.CachedTriggerPosition(triggerName)
	if triggerPosition < 0 {
		t.tbl.AppendTrigger(Trigger{
			TableSchema: t.tbl.TableSchema,
			TableName:   t.tbl.TableName,
			TriggerName: triggerName,
			SQL:         sql,
		})
	} else {
		t.tbl.Triggers[triggerPosition].SQL = sql
	}
}

func (t *TTrigger) Sprintf(format string, values ...interface{}) *TTrigger {
	if len(values) == 0 {
		t.tbl.Triggers[t.triggerPosition].SQL = format
		return t
	}
	buf := bufpool.Get().(*bytes.Buffer)
	args := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		args = args[:0]
		bufpool.Put(buf)
		argspool.Put(args)
	}()
	err := sq.BufferPrintf(t.dialect, buf, &args, make(map[string][]int), nil, format, values)
	if err != nil {
		panicErr(fmt.Errorf("Sprintf: %w", err))
	}
	if len(args) == 0 {
		t.tbl.Triggers[t.triggerPosition].SQL = buf.String()
		return t
	}
	sql, err := sq.Sprintf(t.dialect, buf.String(), args)
	if err != nil {
		panicErr(fmt.Errorf("Sprintf: %w", err))
	}
	t.tbl.Triggers[t.triggerPosition].SQL = sql
	return t
}

func (t *TTrigger) Filef(fsys fs.FS, fileName string, values ...interface{}) *TTrigger {
	b, err := fs.ReadFile(fsys, fileName)
	if err != nil {
		panicErr(fmt.Errorf("Filef: %w", err))
	}
	sql := string(b)
	if len(values) == 0 {
		t.tbl.Triggers[t.triggerPosition].SQL = sql
		return t
	}
	buf := bufpool.Get().(*bytes.Buffer)
	args := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		args = args[:0]
		bufpool.Put(buf)
		argspool.Put(args)
	}()
	err = sq.BufferPrintf(t.dialect, buf, &args, make(map[string][]int), nil, sql, values)
	if err != nil {
		panicErr(fmt.Errorf("Filef: %w", err))
	}
	if len(args) == 0 {
		t.tbl.Triggers[t.triggerPosition].SQL = buf.String()
		return t
	}
	sql, err = sq.Sprintf(t.dialect, sql, args)
	if err != nil {
		panicErr(fmt.Errorf("Filef: %w", err))
	}
	t.tbl.Triggers[t.triggerPosition].SQL = sql
	return t
}
