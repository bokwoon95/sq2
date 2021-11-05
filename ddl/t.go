package ddl

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/bokwoon95/sq"
)

func panicErr(err error) {
	_, file, line, _ := runtime.Caller(2)
	panic(fmt.Errorf("%s:%d: %w", file, line, err))
}

type DDLTable interface {
	sq.SchemaTable
	DDL(dialect string, t *T)
}

type T struct {
	dialect string
	tbl     *Table
}

func (t *T) Ignore() {
	t.tbl.Ignore = true
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

func sprintf(dialect string, format string, values []interface{}, excludedTableQualifiers []string) (string, error) {
	buf := bufpool.Get().(*bytes.Buffer)
	args := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		args = args[:0]
		bufpool.Put(buf)
		argspool.Put(args)
	}()
	err := sq.BufferPrintf(dialect, buf, &args, make(map[string][]int), nil, excludedTableQualifiers, format, values)
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

func (t *TColumn) Generated(format string, values ...interface{}) *TColumn {
	expr, err := sprintf(t.dialect, format, values, []string{t.tbl.TableName})
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
	expr, err := sprintf(t.dialect, format, values, []string{t.tbl.TableName})
	if err != nil {
		panicErr(fmt.Errorf("Default: %w", err))
	}
	if t.dialect != sq.DialectPostgres {
		expr = "(" + expr + ")"
	}
	t.tbl.Columns[t.columnPosition].ColumnDefault = expr
	return t
}

func (t *TColumn) AutoIncrement() *TColumn {
	if t.dialect == sq.DialectMySQL {
		t.tbl.Columns[t.columnPosition].IsAutoincrement = true
	}
	return t
}

func (t *TColumn) Autoincrement() *TColumn {
	if t.dialect == sq.DialectSQLite {
		t.tbl.Columns[t.columnPosition].IsAutoincrement = true
	}
	return t
}

func (t *TColumn) Identity() *TColumn {
	if t.dialect == sq.DialectPostgres {
		t.tbl.Columns[t.columnPosition].Identity = BY_DEFAULT_AS_IDENTITY
	}
	return t
}

func (t *TColumn) AlwaysIdentity() *TColumn {
	if t.dialect == sq.DialectPostgres {
		t.tbl.Columns[t.columnPosition].Identity = ALWAYS_AS_IDENTITY
	}
	return t
}

func (t *TColumn) OnUpdateCurrentTimestamp() *TColumn {
	if t.dialect == sq.DialectMySQL {
		t.tbl.Columns[t.columnPosition].OnUpdateCurrentTimestamp = true
	}
	return t
}

func (t *TColumn) NotNull() *TColumn {
	t.tbl.Columns[t.columnPosition].IsNotNull = true
	return t
}

func (t *TColumn) PrimaryKey() *TColumn {
	constraintName := generateName(PRIMARY_KEY, t.tbl.TableName, t.columnName)
	_, err := createOrUpdateConstraint(t.dialect, t.tbl, PRIMARY_KEY, constraintName, []string{t.columnName}, "")
	if err != nil {
		panicErr(fmt.Errorf("PrimaryKey: %w", err))
	}
	return t
}

func (t *TColumn) Unique() *TColumn {
	constraintName := generateName(UNIQUE, t.tbl.TableName, t.columnName)
	_, err := createOrUpdateConstraint(t.dialect, t.tbl, UNIQUE, constraintName, []string{t.columnName}, "")
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

func createOrUpdateConstraint(dialect string, tbl *Table, constraintType, constraintName string, columns []string, checkExpr string) (constraintPosition int, err error) {
	if dialect != sq.DialectSQLite && constraintName == "" {
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

func createOrUpdateExclusionConstraint(dialect string, tbl *Table, constraintName, exclusionIndexType string, fields []sq.Field, operators []string) (constraintPosition int, err error) {
	if dialect != sq.DialectPostgres {
		return -1, fmt.Errorf("%s does not support exclusion constraints", dialect)
	}
	if constraintName == "" {
		return -1, fmt.Errorf("constraintName cannot be empty")
	}
	columnNames, exprs, err := getColumnNamesAndExprs(dialect, tbl.TableName, fields, true)
	if err != nil {
		return -1, err
	}
	if constraintPosition = tbl.CachedConstraintPosition(constraintName); constraintPosition >= 0 {
		constraint := tbl.Constraints[constraintPosition]
		constraint.TableSchema = tbl.TableSchema
		constraint.TableName = tbl.TableName
		constraint.ConstraintType = EXCLUDE
		constraint.Columns = columnNames
		constraint.Exprs = exprs
		constraint.ExclusionOperators = operators
		constraint.ExclusionIndexType = exclusionIndexType
		tbl.Constraints[constraintPosition] = constraint
	} else {
		constraintPosition = tbl.AppendConstraint(Constraint{
			TableSchema:        tbl.TableSchema,
			TableName:          tbl.TableName,
			ConstraintName:     constraintName,
			ConstraintType:     EXCLUDE,
			Columns:            columnNames,
			Exprs:              exprs,
			ExclusionOperators: operators,
			ExclusionIndexType: exclusionIndexType,
		})
	}
	return constraintPosition, nil
}

type Exclusions []struct {
	Field    sq.Field
	Operator string
}

func (t *T) Exclude(exclusionIndex string, exclusions Exclusions) *TConstraint {
	fields := make([]sq.Field, len(exclusions))
	operators := make([]string, len(exclusions))
	for i, exclusion := range exclusions {
		fields[i] = exclusion.Field
		operators[i] = exclusion.Operator
	}
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicErr(fmt.Errorf("Exclude: %w", err))
	}
	constraintName := generateName(EXCLUDE, t.tbl.TableName, columnNames...)
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintPosition, err = createOrUpdateExclusionConstraint(t.dialect, t.tbl, constraintName, exclusionIndex, fields, operators)
	if err != nil {
		panicErr(fmt.Errorf("Exclude: %w", err))
	}
	return tConstraint
}

func (t *T) Check(constraintName string, format string, values ...interface{}) *TConstraint {
	expr, err := sprintf(t.dialect, format, values, []string{t.tbl.TableName})
	if err != nil {
		panicErr(fmt.Errorf("Check: %w", err))
	}
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.dialect, t.tbl, CHECK, constraintName, nil, expr)
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
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.dialect, t.tbl, UNIQUE, constraintName, columnNames, "")
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
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.dialect, t.tbl, PRIMARY_KEY, constraintName, columnNames, "")
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
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.dialect, t.tbl, FOREIGN_KEY, constraintName, columnNames, "")
	if err != nil {
		panicErr(fmt.Errorf("ForeignKey: %w", err))
	}
	return tConstraint
}

func (t *T) NameExclude(constraintName, exclusionIndex string, exclusions Exclusions) *TConstraint {
	var err error
	tConstraint := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	fields := make([]sq.Field, len(exclusions))
	operators := make([]string, len(exclusions))
	for i, exclusion := range exclusions {
		fields[i] = exclusion.Field
		operators[i] = exclusion.Operator
	}
	tConstraint.constraintPosition, err = createOrUpdateExclusionConstraint(t.dialect, t.tbl, constraintName, exclusionIndex, fields, operators)
	if err != nil {
		panicErr(fmt.Errorf("Exclude: %w", err))
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
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.dialect, t.tbl, UNIQUE, constraintName, columnNames, "")
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
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.dialect, t.tbl, PRIMARY_KEY, constraintName, columnNames, "")
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
	tConstraint.constraintPosition, err = createOrUpdateConstraint(t.dialect, t.tbl, FOREIGN_KEY, constraintName, columnNames, "")
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
		panicErr(fmt.Errorf("References: table is nil"))
	}
	referencesTable := table.GetName()
	if referencesTable == "" {
		panicErr(fmt.Errorf("References: table has no name"))
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
	constraint.UpdateRule = action
	t.tbl.Constraints[t.constraintPosition] = constraint
	return t
}

func (t *TConstraint) OnDelete(action string) *TConstraint {
	constraint := t.tbl.Constraints[t.constraintPosition]
	constraint.DeleteRule = action
	t.tbl.Constraints[t.constraintPosition] = constraint
	return t
}

func (t *TConstraint) Deferrable() *TConstraint {
	if t.dialect == sq.DialectMySQL {
		return t
	}
	constraint := t.tbl.Constraints[t.constraintPosition]
	constraint.IsDeferrable = true
	t.tbl.Constraints[t.constraintPosition] = constraint
	return t
}

func (t *TConstraint) InitiallyDeferred() *TConstraint {
	if t.dialect == sq.DialectMySQL {
		return t
	}
	constraint := t.tbl.Constraints[t.constraintPosition]
	constraint.IsDeferrable = true
	constraint.IsInitiallyDeferred = true
	t.tbl.Constraints[t.constraintPosition] = constraint
	return t
}

func (t *TConstraint) Where(format string, values ...interface{}) *TConstraint {
	expr, err := sprintf(t.dialect, format, values, []string{t.tbl.TableName})
	if err != nil {
		panicErr(fmt.Errorf("Where: %w", err))
	}
	t.tbl.Constraints[t.constraintPosition].Predicate = expr
	return t
}

type TIndex struct {
	dialect       string
	tbl           *Table
	indexName     string
	indexPosition int
}

func getColumnNamesAndExprs(dialect, tableName string, fields []sq.Field, dontWrapExpr bool) (columnNames, exprs []string, err error) {
	// TODO: cleanup the dirty use of an ad-hoc boolean flag "dontWrapExpr". Think of a better way to write it.
	buf := bufpool.Get().(*bytes.Buffer)
	args := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		args = args[:0]
		bufpool.Put(buf)
		argspool.Put(args)
	}()
	for i, field := range fields {
		if field == nil {
			return nil, nil, fmt.Errorf("field #%d is nil", i+1)
		}
		var expr string
		columnName := field.GetName()
		if _, ok := field.(sq.FieldLiteral); ok {
			expr, columnName = columnName, ""
		} else if columnName == "" {
			buf.Reset()
			args = args[:0]
			err = field.AppendSQLExclude(dialect, buf, &args, make(map[string][]int), nil, []string{tableName})
			if err != nil {
				return nil, nil, fmt.Errorf("field #%d, :%w", i+1, err)
			}
			expr = buf.String()
			if len(args) > 0 {
				expr, err = sq.Sprintf(dialect, expr, args)
				if err != nil {
					return nil, nil, fmt.Errorf("field #%d, :%w", i+1, err)
				}
			}
			if !dontWrapExpr {
				expr = "(" + expr + ")"
			}
		}
		columnNames = append(columnNames, columnName)
		exprs = append(exprs, expr)
	}
	return columnNames, exprs, nil
}

func (tbl *Table) createOrUpdateIndex(indexName string, columns []string, exprs []string) (indexPosition int, err error) {
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
			Columns:     columns,
			Exprs:       exprs,
		})
	}
	return indexPosition, nil
}

func (t *T) Index(fields ...sq.Field) *TIndex {
	columnNames, exprs, err := getColumnNamesAndExprs(t.dialect, t.tbl.TableName, fields, false)
	if err != nil {
		panicErr(fmt.Errorf("Index: %w", err))
	}
	indexName := generateName(INDEX, t.tbl.TableName, columnNames...)
	tIndex := &TIndex{
		dialect:   t.dialect,
		tbl:       t.tbl,
		indexName: indexName,
	}
	tIndex.indexPosition, err = t.tbl.createOrUpdateIndex(indexName, columnNames, exprs)
	if err != nil {
		panicErr(fmt.Errorf("Index: %w", err))
	}
	return tIndex
}

func (t *T) NameIndex(indexName string, fields ...sq.Field) *TIndex {
	columnNames, exprs, err := getColumnNamesAndExprs(t.dialect, t.tbl.TableName, fields, false)
	if err != nil {
		panicErr(fmt.Errorf("NameIndex: %w", err))
	}
	tIndex := &TIndex{
		dialect:   t.dialect,
		tbl:       t.tbl,
		indexName: indexName,
	}
	tIndex.indexPosition, err = t.tbl.createOrUpdateIndex(indexName, columnNames, exprs)
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
	expr, err := sprintf(t.dialect, format, values, []string{t.tbl.TableName})
	if err != nil {
		panicErr(fmt.Errorf("Where: %w", err))
	}
	t.tbl.Indexes[t.indexPosition].Predicate = expr
	return t
}

func (t *TIndex) Include(fields ...sq.Field) *TIndex {
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicErr(fmt.Errorf("Include: %w", err))
	}
	t.tbl.Indexes[t.indexPosition].IncludeColumns = columnNames
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

func (t *T) Trigger(format string, values ...interface{}) {
	sql, err := sprintf(t.dialect, format, values, nil)
	if err != nil {
		panicErr(fmt.Errorf("Trigger: %w", err))
	}
	trigger := Trigger{SQL: strings.TrimSpace(sql)}
	err = trigger.populateTriggerInfo(t.dialect)
	if err != nil {
		panicErr(fmt.Errorf("Trigger: %w", err))
	}
	if trigger.TableSchema != "" && trigger.TableSchema != t.tbl.TableSchema {
		panicErr(fmt.Errorf("Trigger: table schema does not match (got=%s, want=%s)", trigger.TableSchema, t.tbl.TableSchema))
	}
	if trigger.TableName != t.tbl.TableName {
		panicErr(fmt.Errorf("Trigger: table name does not match (got=%s, want=%s)", trigger.TableName, t.tbl.TableName))
	}
	if n := t.tbl.CachedTriggerPosition(trigger.TriggerName); n >= 0 {
		t.tbl.Triggers[n].SQL = trigger.SQL
	} else {
		t.tbl.AppendTrigger(trigger)
	}
}

func (tbl *Table) LoadTable(dialect string, table sq.SchemaTable) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case error:
				err = r
			default:
				err = fmt.Errorf("panic: " + fmt.Sprint(r))
			}
		}
	}()
	if table == nil {
		return fmt.Errorf("table is nil")
	}
	tableValue := reflect.ValueOf(table)
	tableType := tableValue.Type()
	if tableType.Kind() != reflect.Struct {
		return fmt.Errorf("table is not a struct")
	}
	tbl.TableSchema, tbl.TableName = table.GetSchema(), table.GetName()
	if tbl.TableName == "" {
		return fmt.Errorf("table name is empty")
	}
	qualifiedTable := tbl.TableName
	if tbl.TableSchema != "" {
		qualifiedTable = tbl.TableSchema + "." + tbl.TableName
	}
	for i := 0; i < tableValue.NumField(); i++ {
		fieldValue := tableValue.Field(i)
		if !fieldValue.CanInterface() {
			err = tbl.loadTableConfig(dialect, qualifiedTable, tableType.Field(i).Tag.Get("ddl"))
			if err != nil {
				return err
			}
			if tbl.Ignore {
				return nil
			}
			continue
		}
		field, ok := tableValue.Field(i).Interface().(sq.Field)
		if !ok {
			err = tbl.loadTableConfig(dialect, qualifiedTable, tableType.Field(i).Tag.Get("ddl"))
			if err != nil {
				return err
			}
			if tbl.Ignore {
				return nil
			}
			continue
		}
		columnName := field.GetName()
		if columnName == "" {
			return fmt.Errorf("table %s field #%d has no name", tbl.TableName, i)
		}
		columnType := defaultColumnType(dialect, field)
		config := tableType.Field(i).Tag.Get("ddl")
		err = tbl.loadColumnConfig(dialect, columnName, columnType, config)
		if err != nil {
			return err
		}
	}
	defer func() {
		if strings.EqualFold(tbl.VirtualTable, "FTS5") {
			var columnNames []string
			for _, column := range tbl.Columns {
				if column.Ignore {
					continue
				}
				if strings.EqualFold(column.ColumnType, "TEXT") {
					columnNames = append(columnNames, column.ColumnName)
				}
			}
			tbl.VirtualTableArgs = append(columnNames, tbl.VirtualTableArgs...)
		}
		for _, constraint := range tbl.Constraints {
			if len(constraint.Columns) != 1 {
				continue
			}
			n := tbl.CachedColumnPosition(constraint.Columns[0])
			if n < 0 {
				continue
			}
			switch constraint.ConstraintType {
			case PRIMARY_KEY:
				tbl.Columns[n].IsPrimaryKey = true
			case UNIQUE:
				tbl.Columns[n].IsUnique = true
			}
		}
	}()
	if ddlTable, ok := table.(DDLTable); ok {
		ddlTable.DDL(dialect, &T{dialect: dialect, tbl: tbl})
	}
	return nil
}

func (tbl *Table) loadTableConfig(dialect, qualifiedTable, tableModifiers string) error {
	modifiers, _, err := tokenizeModifiers(tableModifiers)
	if err != nil {
		return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
	}
	for _, modifier := range modifiers {
		modifierName := modifier[0]
		if i := strings.IndexByte(modifierName, ':'); i >= 0 {
			modifierDialect := modifierName[:i]
			modifierName = modifierName[i+1:]
			if modifierDialect != dialect {
				continue
			}
		}
		switch modifierName {
		case "virtual":
			if dialect == sq.DialectSQLite {
				virtualTable, submodifiers, _, err := tokenizeValue(modifier[1])
				if err != nil {
					return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
				}
				tbl.VirtualTable = virtualTable
				for _, submodifier := range submodifiers {
					virtualTableArg := submodifier[0]
					if submodifier[1] != "" {
						virtualTableArg += "=" + submodifier[1]
					}
					tbl.VirtualTableArgs = append(tbl.VirtualTableArgs, virtualTableArg)
				}
			}
		case "primarykey":
			err = tbl.loadConstraintConfig(dialect, PRIMARY_KEY, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "references":
			err = tbl.loadConstraintConfig(dialect, FOREIGN_KEY, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "unique":
			err = tbl.loadConstraintConfig(dialect, UNIQUE, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "index":
			err = tbl.loadIndexConfig(dialect, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "ignore":
			if modifier[1] == "" {
				tbl.Ignore = true
			} else {
				ignoredDialects := strings.Split(modifier[1], ",")
				for _, ignoredDialect := range ignoredDialects {
					if dialect == ignoredDialect {
						tbl.Ignore = true
						break
					}
				}
			}
		default:
			return fmt.Errorf("%s: unknown modifier '%s'", qualifiedTable, modifier[0])
		}
	}
	return nil
}
