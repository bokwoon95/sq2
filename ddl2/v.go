package ddl

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/bokwoon95/sq"
)

type DDLView interface {
	sq.SchemaTable
	DDL(dialect string, v *V)
}

type V struct {
	dialect     string
	view        *View
	wantColumns []string
}

func (v *V) IsMaterialized() {
	if v.dialect == sq.DialectPostgres {
		v.view.IsMaterialized = true
	}
}

func (v *V) AsQuery(query sq.Query) {
	if query == nil {
		panicErr(fmt.Errorf("query is nil"))
	}
	buf := bufpool.Get().(*bytes.Buffer)
	args := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		args = args[:0]
		bufpool.Put(buf)
		argspool.Put(args)
	}()
	err := query.AppendSQL(v.dialect, buf, &args, make(map[string][]int))
	if err != nil {
		panicErr(err)
	}
	v.view.SQL = buf.String()
	if len(args) > 0 {
		v.view.SQL, err = sq.Sprintf(v.dialect, v.view.SQL, args)
		if err != nil {
			panicErr(err)
		}
	}
	columnDiff := make(map[string]int)
	for _, column := range v.wantColumns {
		columnDiff[column]--
	}
	fields, err := query.GetFetchableFields()
	if err != nil {
		panicErr(err)
	}
	for i, field := range fields {
		column := field.GetAlias()
		if column == "" {
			column = field.GetName()
		}
		if column == "" {
			panicErr(fmt.Errorf("view %s query column #%d has no name and no alias", v.view.ViewName, i+1))
		}
		columnDiff[column]++
	}
	var extraColumns, missingColumns []string
	for column, n := range columnDiff {
		if n == 0 {
			continue
		}
		if n > 0 {
			extraColumns = append(extraColumns, column)
		} else {
			missingColumns = append(missingColumns, column)
		}
	}
	if len(missingColumns) > 0 || len(extraColumns) > 0 {
		errMsg := fmt.Sprintf("view %s query columns do not match struct fields:", v.view.ViewName)
		if len(missingColumns) > 0 {
			sort.Strings(missingColumns)
			errMsg += fmt.Sprintf(" (missingColumns=%s)", strings.Join(missingColumns, ", "))
		}
		if len(extraColumns) > 0 {
			sort.Strings(extraColumns)
			errMsg += fmt.Sprintf(" (extraColumns=%s)", strings.Join(extraColumns, ", "))
		}
		panicErr(fmt.Errorf(errMsg))
	}
}

func (v *V) Trigger(sql string) {
	if v.dialect != sq.DialectPostgres {
		return
	}
	trigger := Trigger{SQL: strings.TrimSpace(sql)}
	err := trigger.populateTriggerInfo(v.dialect)
	if err != nil {
		panicErr(fmt.Errorf("Trigger: %w", err))
	}
	if n := v.view.CachedTriggerPosition(trigger.TableSchema, trigger.TableName, trigger.TriggerName); n >= 0 {
		v.view.Triggers[n].SQL = trigger.SQL
	} else {
		v.view.AppendTrigger(trigger)
	}
}

func (v *V) Sprintf(format string, values ...interface{}) string {
	str, err := sprintf(v.dialect, format, values, nil)
	if err != nil {
		panicErr(fmt.Errorf("Sprintf: %w", err))
	}
	return str
}

type VIndex struct {
	dialect       string
	view          *View
	indexName     string
	indexPosition int
}

func (v *V) Index(fields ...sq.Field) *VIndex {
	if v.dialect != sq.DialectPostgres {
		return &VIndex{dialect: v.dialect}
	}
	if !v.view.IsMaterialized {
		panicErr(fmt.Errorf("Indexes can only be defined on Materialized Views"))
	}
	if v.view.IsMaterialized && v.dialect != sq.DialectPostgres {
		panicErr(fmt.Errorf("%s does not support Materialized Views", v.dialect))
	}
	columnNames, exprs, err := getColumnNamesAndExprs(v.dialect, v.view.ViewName, fields, false)
	if err != nil {
		panicErr(fmt.Errorf("Index: %w", err))
	}
	indexName := generateName(INDEX, v.view.ViewName, columnNames...)
	tIndex := &VIndex{
		dialect:   v.dialect,
		view:      v.view,
		indexName: indexName,
	}
	tIndex.indexPosition, err = v.view.createOrUpdateIndex(indexName, columnNames, exprs)
	if err != nil {
		panicErr(fmt.Errorf("Index: %w", err))
	}
	return tIndex
}

func (v *V) NameIndex(indexName string, fields ...sq.Field) *VIndex {
	if v.dialect != sq.DialectPostgres {
		return &VIndex{dialect: v.dialect}
	}
	if !v.view.IsMaterialized {
		panicErr(fmt.Errorf("NameIndex: Indexes can only be defined on Materialized Views"))
	}
	if v.view.IsMaterialized && v.dialect != sq.DialectPostgres {
		panicErr(fmt.Errorf("NameIndex: %s does not support Materialized Views (which are needed for indexes)", v.dialect))
	}
	columnNames, exprs, err := getColumnNamesAndExprs(v.dialect, v.view.ViewName, fields, false)
	if err != nil {
		panicErr(fmt.Errorf("NameIndex: %w", err))
	}
	tIndex := &VIndex{
		dialect:   v.dialect,
		view:      v.view,
		indexName: indexName,
	}
	tIndex.indexPosition, err = v.view.createOrUpdateIndex(indexName, columnNames, exprs)
	if err != nil {
		panicErr(fmt.Errorf("NameIndex: %w", err))
	}
	return tIndex
}

func (v *VIndex) Unique() *VIndex {
	if v.dialect != sq.DialectPostgres {
		return v
	}
	v.view.Indexes[v.indexPosition].IsUnique = true
	return v
}

func (v *VIndex) Using(indexType string) *VIndex {
	if v.dialect != sq.DialectPostgres {
		return v
	}
	v.view.Indexes[v.indexPosition].IndexType = strings.ToUpper(indexType)
	return v
}

func (v *VIndex) Where(format string, values ...interface{}) *VIndex {
	if v.dialect != sq.DialectPostgres {
		return v
	}
	expr, err := sprintf(v.dialect, format, values, []string{v.view.ViewName})
	if err != nil {
		panicErr(fmt.Errorf("Where: %w", err))
	}
	v.view.Indexes[v.indexPosition].Predicate = expr
	return v
}

func (v *VIndex) Include(fields ...sq.Field) *VIndex {
	if v.dialect != sq.DialectPostgres {
		return v
	}
	columnNames, err := getColumnNames(fields)
	if err != nil {
		panicErr(fmt.Errorf("Include: %w", err))
	}
	v.view.Indexes[v.indexPosition].IncludeColumns = columnNames
	return v
}

func (v *VIndex) Config(config func(index *Index)) {
	if v.dialect != sq.DialectPostgres {
		return
	}
	index := v.view.Indexes[v.indexPosition]
	config(&index)
	index.TableSchema = v.view.ViewSchema
	index.TableName = v.view.ViewName
	index.IndexName = v.indexName
	v.view.Indexes[v.indexPosition] = index
}

func (view *View) LoadDDLView(dialect string, ddlView DDLView) (err error) {
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
	if ddlView == nil {
		return fmt.Errorf("view is nil")
	}
	ddlViewValue := reflect.ValueOf(ddlView)
	ddlViewType := ddlViewValue.Type()
	if ddlViewType.Kind() != reflect.Struct {
		return fmt.Errorf("view is not a struct")
	}
	view.ViewSchema, view.ViewName = ddlView.GetSchema(), ddlView.GetName()
	if view.ViewName == "" {
		return fmt.Errorf("view name is empty")
	}
	v := &V{
		dialect: dialect,
		view:    view,
	}
	for i := 0; i < ddlViewValue.NumField(); i++ {
		field, ok := ddlViewValue.Field(i).Interface().(sq.Field)
		if !ok {
			continue
		}
		fieldName := field.GetName()
		if fieldName == "" {
			return fmt.Errorf("view struct %s field #%d has no name set for it", ddlViewType.Name(), i)
		}
		v.wantColumns = append(v.wantColumns, fieldName)
	}
	ddlView.DDL(dialect, v)
	return nil
}
