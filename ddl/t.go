package ddl

import (
	"bytes"
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
	// if i := f.tbl.CachedColumnIndex(f.columnName); i >= 0 {
	// 	f.tbl.Columns[i].Autoincrement = true
	// }
	return f
}

func (f *TColumn) Identity() *TColumn {
	// if i := f.tbl.CachedColumnIndex(f.columnName); i >= 0 {
	// 	f.tbl.Columns[i].Identity = "BY DEFAULT AS IDENTITY"
	// }
	return f
}

func (f *TColumn) AlwaysIdentity() *TColumn {
	// if i := f.tbl.CachedColumnIndex(f.columnName); i >= 0 {
	// 	f.tbl.Columns[i].Identity = "ALWAYS AS IDENTITY"
	// }
	return f
}

func (f *TColumn) OnUpdateCurrentTimestamp() *TColumn {
	// if i := f.tbl.CachedColumnIndex(f.columnName); i >= 0 {
	// 	f.tbl.Columns[i].OnUpdateCurrentTimestamp = true
	// }
	return f
}

func (f *TColumn) NotNull() *TColumn {
	// if i := f.tbl.CachedColumnIndex(f.columnName); i >= 0 {
	// 	f.tbl.Columns[i].IsNotNull = true
	// }
	return f
}

func (f *TColumn) PrimaryKey() *TColumn {
	// constraintName := pgName("PRIMARY KEY", f.tbl.TableName, f.columnName)
	// if i := f.tbl.CachedConstraintIndex(constraintName); i >= 0 {
	// 	f.tbl.Constraints[i].ConstraintType = "PRIMARY KEY"
	// 	f.tbl.Constraints[i].TableName = f.tbl.TableName
	// 	f.tbl.Constraints[i].Columns = []string{f.columnName}
	// } else {
	// 	f.tbl.AppendConstraint(Constraint{
	// 		ConstraintSchema: f.tbl.TableSchema,
	// 		ConstraintName:   constraintName,
	// 		ConstraintType:   "PRIMARY KEY",
	// 		TableSchema:      f.tbl.TableSchema,
	// 		TableName:        f.tbl.TableName,
	// 		Columns:          []string{f.columnName},
	// 	})
	// }
	return f
}

func (f *TColumn) Unique() *TColumn {
	// constraintName := pgName("UNIQUE", f.tbl.TableName, f.columnName)
	// if i := f.tbl.CachedConstraintIndex(constraintName); i >= 0 {
	// 	f.tbl.Constraints[i].ConstraintType = "UNIQUE"
	// 	f.tbl.Constraints[i].TableName = f.tbl.TableName
	// 	f.tbl.Constraints[i].Columns = []string{f.columnName}
	// } else {
	// 	f.tbl.AppendConstraint(Constraint{
	// 		ConstraintSchema: f.tbl.TableSchema,
	// 		ConstraintName:   constraintName,
	// 		ConstraintType:   "UNIQUE",
	// 		TableSchema:      f.tbl.TableSchema,
	// 		TableName:        f.tbl.TableName,
	// 		Columns:          []string{f.columnName},
	// 	})
	// }
	return f
}

func (f *TColumn) Collate(collation string) *TColumn {
	// if i := f.tbl.CachedColumnIndex(f.columnName); i >= 0 {
	// 	f.tbl.Columns[i].CollationName = sql.NullString{Valid: true, String: collation}
	// }
	return f
}

type TConstraint struct {
	dialect        string
	tbl            *Table
	constraintName string
}

func (t *T) Check(name string, expr string, fields ...sq.Field) *TConstraint {
	tc := &TConstraint{
		dialect:        t.dialect,
		tbl:            t.tbl,
		constraintName: name,
	}
	// checkExpr := sql.NullString{Valid: true, String: Sprintf(expr, fields...)}
	// if i := t.tbl.CachedConstraintIndex(name); i >= 0 {
	// 	t.tbl.Constraints[i].ConstraintType = "CHECK"
	// 	t.tbl.Constraints[i].TableName = t.tbl.TableName
	// 	t.tbl.Constraints[i].CheckExpr = checkExpr
	// } else {
	// 	t.tbl.AppendConstraint(Constraint{
	// 		ConstraintSchema: t.tbl.TableSchema,
	// 		ConstraintName:   name,
	// 		ConstraintType:   "UNIQUE",
	// 		TableSchema:      t.tbl.TableSchema,
	// 		TableName:        t.tbl.TableName,
	// 		CheckExpr:        checkExpr,
	// 	})
	// }
	return tc
}

func (t *T) Unique(fields ...sq.Field) {
}

func (t *T) PrimaryKey(fields ...sq.Field) {
}

func (t *T) NameUnique(name string, fields ...sq.Field) {
}

func (t *T) NamePrimaryKey(name string, fields ...sq.Field) {
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
