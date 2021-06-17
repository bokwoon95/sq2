package ddl

import (
	"bytes"
	"database/sql"
	"fmt"
	"runtime"

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

func (f *TColumn) Ignore() {
	f.tbl.Columns[f.columnIndex].Ignore = true
}

func (f *TColumn) Type(columnType string) *TColumn {
	f.tbl.Columns[f.columnIndex].ColumnType = columnType
	return f
}

func (f *TColumn) Config(config func(col *Column)) {
	col := f.tbl.Columns[f.columnIndex]
	config(&col)
	col.ColumnName = f.columnName
	f.tbl.Columns[f.columnIndex] = col
}

func (f *TColumn) Generated(expr sq.Field) *TColumn {
	buf := bufpool.Get().(*bytes.Buffer)
	args := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		args = args[:0]
		bufpool.Put(buf)
		argspool.Put(args)
	}()
	err := expr.AppendSQLExclude(f.dialect, buf, &args, make(map[string][]int), []string{f.tbl.TableName})
	if err != nil {
		panicf(err.Error())
	}
	f.tbl.Columns[f.columnIndex].GeneratedExpr.Valid = true
	f.tbl.Columns[f.columnIndex].GeneratedExpr.String = sq.Sprintf(f.dialect, buf.String(), args)
	return f
}

func (f *TColumn) Stored() *TColumn {
	f.tbl.Columns[f.columnIndex].GeneratedExprStored.Valid = true
	f.tbl.Columns[f.columnIndex].GeneratedExprStored.Bool = true
	return f
}

func (f *TColumn) Default(expr sq.Field) *TColumn {
	buf := bufpool.Get().(*bytes.Buffer)
	args := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		args = args[:0]
		bufpool.Put(buf)
		argspool.Put(args)
	}()
	err := expr.AppendSQLExclude(f.dialect, buf, &args, make(map[string][]int), []string{f.tbl.TableName})
	if err != nil {
		panicf(err.Error())
	}
	f.tbl.Columns[f.columnIndex].ColumnDefault.Valid = true
	f.tbl.Columns[f.columnIndex].ColumnDefault.String = sq.Sprintf(f.dialect, buf.String(), args)
	return f
}

func (f *TColumn) Autoincrement() *TColumn {
	f.tbl.Columns[f.columnIndex].Autoincrement = true
	return f
}

func (f *TColumn) Identity() *TColumn {
	f.tbl.Columns[f.columnIndex].Identity = IDENTITY_DEFAULT
	return f
}

func (f *TColumn) AlwaysIdentity() *TColumn {
	f.tbl.Columns[f.columnIndex].Identity = IDENTITY_ALWAYS
	return f
}

func (f *TColumn) OnUpdateCurrentTimestamp() *TColumn {
	f.tbl.Columns[f.columnIndex].OnUpdateCurrentTimestamp = true
	return f
}

func (f *TColumn) NotNull() *TColumn {
	f.tbl.Columns[f.columnIndex].IsNotNull = true
	return f
}

func (f *TColumn) PrimaryKey() *TColumn {
	constraintName := pgName(PRIMARY_KEY, f.tbl.TableName, f.columnName)
	if i := f.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		f.tbl.Constraints[i].ConstraintType = PRIMARY_KEY
		f.tbl.Constraints[i].TableName = f.tbl.TableName
		f.tbl.Constraints[i].Columns = []string{f.columnName}
	} else {
		f.tbl.AppendConstraint(Constraint{
			ConstraintSchema: f.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   PRIMARY_KEY,
			TableSchema:      f.tbl.TableSchema,
			TableName:        f.tbl.TableName,
			Columns:          []string{f.columnName},
		})
	}
	return f
}

func (f *TColumn) Unique() *TColumn {
	constraintName := pgName(UNIQUE, f.tbl.TableName, f.columnName)
	if i := f.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		f.tbl.Constraints[i].ConstraintType = UNIQUE
		f.tbl.Constraints[i].TableName = f.tbl.TableName
		f.tbl.Constraints[i].Columns = []string{f.columnName}
	} else {
		f.tbl.AppendConstraint(Constraint{
			ConstraintSchema: f.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   UNIQUE,
			TableSchema:      f.tbl.TableSchema,
			TableName:        f.tbl.TableName,
			Columns:          []string{f.columnName},
		})
	}
	return f
}

func (f *TColumn) Collate(collation string) *TColumn {
	f.tbl.Columns[f.columnIndex].CollationName.Valid = true
	f.tbl.Columns[f.columnIndex].CollationName.String = collation
	return f
}

type TConstraint struct {
	dialect         string
	tbl             *Table
	constraintName  string
	constraintIndex int
}

func (tc *TConstraint) Config(config func(constraint *Constraint)) {
	constraint := tc.tbl.Constraints[tc.constraintIndex]
	config(&constraint)
	constraint.ConstraintName = tc.constraintName
	tc.tbl.Constraints[tc.constraintIndex] = constraint
}

func (t *T) Check(constraintName string, predicate sq.Predicate) *TConstraint {
	buf := bufpool.Get().(*bytes.Buffer)
	args := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		args = args[:0]
		bufpool.Put(buf)
		argspool.Put(args)
	}()
	err := predicate.AppendSQLExclude(t.dialect, buf, &args, make(map[string][]int), []string{t.tbl.TableName})
	if err != nil {
		panicf(err.Error())
	}
	exprString := sq.Sprintf(t.dialect, buf.String(), args)
	tc := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	if i := t.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		t.tbl.Constraints[i].ConstraintType = CHECK
		t.tbl.Constraints[i].TableName = t.tbl.TableName
		t.tbl.Constraints[i].CheckExpr.Valid = true
		t.tbl.Constraints[i].CheckExpr.String = exprString
		tc.constraintIndex = i
	} else {
		t.tbl.AppendConstraint(Constraint{
			ConstraintSchema: t.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   CHECK,
			TableSchema:      t.tbl.TableSchema,
			TableName:        t.tbl.TableName,
			CheckExpr: sql.NullString{
				Valid:  true,
				String: exprString,
			},
		})
		tc.constraintIndex = t.tbl.CachedConstraintIndex(constraintName)
	}
	if tc.constraintIndex < 0 {
		panicf("could not create or update constraint '%s'", constraintName)
	}
	return tc
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
	tc := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	if i := t.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		t.tbl.Constraints[i].ConstraintType = UNIQUE
		t.tbl.Constraints[i].TableSchema = t.tbl.TableSchema
		t.tbl.Constraints[i].TableName = t.tbl.TableName
		t.tbl.Constraints[i].Columns = columnNames
		tc.constraintIndex = i
	} else {
		t.tbl.AppendConstraint(Constraint{
			ConstraintSchema: t.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   UNIQUE,
			TableSchema:      t.tbl.TableSchema,
			TableName:        t.tbl.TableName,
			Columns:          columnNames,
		})
		tc.constraintIndex = t.tbl.CachedConstraintIndex(constraintName)
	}
	if tc.constraintIndex < 0 {
		panicf("could not create or update constraint '%s'", constraintName)
	}
	return tc
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
	tc := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	if i := t.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		t.tbl.Constraints[i].ConstraintType = PRIMARY_KEY
		t.tbl.Constraints[i].TableSchema = t.tbl.TableSchema
		t.tbl.Constraints[i].TableName = t.tbl.TableName
		t.tbl.Constraints[i].Columns = columnNames
		tc.constraintIndex = i
	} else {
		t.tbl.AppendConstraint(Constraint{
			ConstraintSchema: t.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   PRIMARY_KEY,
			TableSchema:      t.tbl.TableSchema,
			TableName:        t.tbl.TableName,
			Columns:          columnNames,
		})
		tc.constraintIndex = t.tbl.CachedConstraintIndex(constraintName)
	}
	if tc.constraintIndex < 0 {
		panicf("could not create or update constraint '%s'", constraintName)
	}
	return tc
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
	tc := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	if i := t.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		t.tbl.Constraints[i].ConstraintType = UNIQUE
		t.tbl.Constraints[i].TableSchema = t.tbl.TableSchema
		t.tbl.Constraints[i].TableName = t.tbl.TableName
		t.tbl.Constraints[i].Columns = columnNames
		tc.constraintIndex = i
	} else {
		t.tbl.AppendConstraint(Constraint{
			ConstraintSchema: t.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   UNIQUE,
			TableSchema:      t.tbl.TableSchema,
			TableName:        t.tbl.TableName,
			Columns:          columnNames,
		})
		tc.constraintIndex = t.tbl.CachedConstraintIndex(constraintName)
	}
	if tc.constraintIndex < 0 {
		panicf("could not create or update constraint '%s'", constraintName)
	}
	return tc
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
	tc := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: constraintName,
	}
	if i := t.tbl.CachedConstraintIndex(constraintName); i >= 0 {
		t.tbl.Constraints[i].ConstraintType = PRIMARY_KEY
		t.tbl.Constraints[i].TableSchema = t.tbl.TableSchema
		t.tbl.Constraints[i].TableName = t.tbl.TableName
		t.tbl.Constraints[i].Columns = columnNames
		tc.constraintIndex = i
	} else {
		t.tbl.AppendConstraint(Constraint{
			ConstraintSchema: t.tbl.TableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   PRIMARY_KEY,
			TableSchema:      t.tbl.TableSchema,
			TableName:        t.tbl.TableName,
			Columns:          columnNames,
		})
		tc.constraintIndex = t.tbl.CachedConstraintIndex(constraintName)
	}
	if tc.constraintIndex < 0 {
		panicf("could not create or update constraint '%s'", constraintName)
	}
	return tc
}

type TIndex struct {
}

func (t *T) Index(fields ...sq.Field) *TIndex {
	return &TIndex{}
}

func (t *T) NameIndex(name string) *TIndex {
	return &TIndex{}
}

func (i *TIndex) Field(fields ...sq.Field) *TIndex {
	return i
}

func (i *TIndex) Expr(expr string, fields ...sq.Field) *TIndex {
	return i
}

func (i *TIndex) Unique() *TIndex {
	return i
}

func (i *TIndex) Schema(schema string) *TIndex {
	return i
}

func (i *TIndex) Using(method string) *TIndex {
	return i
}

func (i *TIndex) Where(expr string, fields ...sq.Field) *TIndex {
	return i
}

func (i *TIndex) Include(fields ...sq.Field) *TIndex {
	return i
}
