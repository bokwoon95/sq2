package ddl

import (
	"bytes"
	"fmt"
	"io/fs"
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
	err := sq.BufferPrintf(dialect, buf, &args, make(map[string][]int), excludedTableQualifiers, format, values)
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
	expr, err := sprintf(t.dialect, format, values, nil)
	if err != nil {
		panicErr(fmt.Errorf("Sprintf: %w", err))
	}
	return expr
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
	expr, err := sprintf(t.dialect, format, values, []string{t.tbl.TableName})
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
			err = field.AppendSQLExclude(dialect, buf, &args, make(map[string][]int), []string{tableName})
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
			expr = "(" + expr + ")"
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
	tIndex.indexPosition, err = t.tbl.createOrUpdateIndex(indexName, columnNames, exprs)
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

func (t *T) Trigger(sql string) {
	trigger := Trigger{SQL: sql}
	err := trigger.populateTriggerInfo(t.dialect)
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

func (t *T) TriggerFile(fsys fs.FS, name string) {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		panicErr(fmt.Errorf("TriggerFile: %w", err))
	}
	trigger := Trigger{SQL: string(b)}
	err = trigger.populateTriggerInfo(t.dialect)
	if err != nil {
		panicErr(fmt.Errorf("TriggerFile: %w", err))
	}
	if trigger.TableSchema != "" && trigger.TableSchema != t.tbl.TableSchema {
		panicErr(fmt.Errorf("TriggerFile: table schema does not match (got=%s, want=%s)", trigger.TableSchema, t.tbl.TableSchema))
	}
	if trigger.TableName != t.tbl.TableName {
		panicErr(fmt.Errorf("TriggerFile: table name does not match (got=%s, want=%s)", trigger.TableName, t.tbl.TableName))
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
	tableModifiers := tableType.Field(0).Tag.Get("ddl")
	modifiers, _, err := tokenizeModifiers(tableModifiers)
	if err != nil {
		return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
	}
	for _, modifier := range modifiers {
		switch modifier[0] {
		case "virtual":
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
		case "primarykey":
			err = tbl.LoadConstraintConfig(PRIMARY_KEY, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "references":
			err = tbl.LoadConstraintConfig(FOREIGN_KEY, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "unique":
			err = tbl.LoadConstraintConfig(UNIQUE, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "check":
			err = tbl.LoadConstraintConfig(CHECK, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "index":
			err = tbl.LoadIndexConfig(tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		default:
			return fmt.Errorf("%s: unknown modifier '%s'", qualifiedTable, modifier[0])
		}
	}
	for i := 0; i < tableValue.NumField(); i++ {
		field, ok := tableValue.Field(i).Interface().(sq.Field)
		if !ok {
			continue
		}
		columnName := field.GetName()
		if columnName == "" {
			return fmt.Errorf("table %s field #%d has no name", tbl.TableName, i)
		}
		columnType := defaultColumnType(dialect, field)
		config := tableType.Field(i).Tag.Get("ddl")
		err = tbl.LoadColumnConfig(dialect, columnName, columnType, config)
		if err != nil {
			return err
		}
	}
	defer func() {
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
