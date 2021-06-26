package sq

import (
	"bytes"
	"testing"
)

func Test_NumberField(t *testing.T) {
	USERS := struct {
		USER_ID CustomField
		NAME    CustomField
		EMAIL   CustomField
		AGE     CustomField
	}{
		USER_ID: NewCustomField("user_id", TableInfo{}),
		NAME:    NewCustomField("name", TableInfo{}),
		EMAIL:   NewCustomField("email", TableInfo{}),
		AGE:     NewCustomField("age", TableInfo{}),
	}

	type TT struct {
		dialect                 string
		item                    SQLExcludeAppender
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
		wantParams              map[string][]int
	}

	assert := func(t *testing.T, tt TT) {
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, gotParams := []interface{}{}, map[string][]int{}
		err := tt.item.AppendSQLExclude(tt.dialect, buf, &gotArgs, gotParams, tt.excludedTableQualifiers)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(buf.String(), tt.wantQuery); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testcallers(), diff)
		}
		if tt.wantParams != nil {
			if diff := testdiff(tt.wantParams, gotParams); diff != "" {
				t.Error(testcallers(), diff)
			}
		}
	}

	t.Run("NumberField", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"})
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("NumberField with alias", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).As("f")
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("NumberField ASC NULLS LAST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).Asc().NullsLast()
		tt.wantQuery = "tbl.field ASC NULLS LAST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("NumberField DESC NULLS FIRST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).Desc().NullsFirst()
		tt.wantQuery = "tbl.field DESC NULLS FIRST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("NumberField IS NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).IsNull()
		tt.wantQuery = "tbl.field IS NULL"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("NumberField IS NOT NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).IsNotNull()
		tt.wantQuery = "tbl.field IS NOT NULL"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("NumberFieldf in (slice)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NumberFieldf(
			"(MAX(AVG({avg1}), AVG({avg2}), SUM({sum})) + {incr})",
			Param("avg1", USERS.USER_ID),
			Param("avg2", USERS.AGE),
			Param("incr", 1),
			Param("sum", USERS.AGE),
		).In([]int{1, 2, 3, 4})
		tt.wantQuery = "(MAX(AVG(user_id), AVG(age), SUM(age)) + ?) IN (?, ?, ?, ?)"
		tt.wantArgs = []interface{}{1, 1, 2, 3, 4}
		tt.wantParams = map[string][]int{"incr": {0}}
		assert(t, tt)
	})

	t.Run("NumberFieldf in (rowvalue)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NumberFieldf(
			"(MAX(AVG({avg1}), AVG({avg2}), SUM({sum})) + {incr})",
			Param("avg1", USERS.USER_ID),
			Param("avg2", USERS.AGE),
			Param("incr", 1),
			Param("sum", USERS.AGE),
		).In(RowValue{USERS.AGE, USERS.USER_ID})
		tt.wantQuery = "(MAX(AVG(user_id), AVG(age), SUM(age)) + ?) IN (age, user_id)"
		tt.wantArgs = []interface{}{1}
		tt.wantParams = map[string][]int{"incr": {0}}
		assert(t, tt)
	})

	t.Run("NumberField Eq", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewNumberField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Eq(field)
		tt.wantQuery = "tbl.field = tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("NumberField Ne", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewNumberField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Ne(field)
		tt.wantQuery = "tbl.field <> tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("NumberField Gt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewNumberField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Gt(field)
		tt.wantQuery = "tbl.field > tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("NumberField Ge", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewNumberField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Ge(field)
		tt.wantQuery = "tbl.field >= tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("NumberField Lt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewNumberField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Lt(field)
		tt.wantQuery = "tbl.field < tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("NumberField Le", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewNumberField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Le(field)
		tt.wantQuery = "tbl.field <= tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("NumberField EqInt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).EqInt(22)
		tt.wantQuery = "tbl.field = ?"
		tt.wantArgs = []interface{}{22}
		assert(t, tt)
	})

	t.Run("NumberField NeInt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).NeInt(22)
		tt.wantQuery = "tbl.field <> ?"
		tt.wantArgs = []interface{}{22}
		assert(t, tt)
	})

	t.Run("NumberField GtInt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).GtInt(22)
		tt.wantQuery = "tbl.field > ?"
		tt.wantArgs = []interface{}{22}
		assert(t, tt)
	})

	t.Run("NumberField GeInt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).GeInt(22)
		tt.wantQuery = "tbl.field >= ?"
		tt.wantArgs = []interface{}{22}
		assert(t, tt)
	})

	t.Run("NumberField LtInt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).LtInt(22)
		tt.wantQuery = "tbl.field < ?"
		tt.wantArgs = []interface{}{22}
		assert(t, tt)
	})

	t.Run("NumberField LeInt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).LeInt(22)
		tt.wantQuery = "tbl.field <= ?"
		tt.wantArgs = []interface{}{22}
		assert(t, tt)
	})

	t.Run("NumberField EqInt64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).EqInt64(22)
		tt.wantQuery = "tbl.field = ?"
		tt.wantArgs = []interface{}{int64(22)}
		assert(t, tt)
	})

	t.Run("NumberField NeInt64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).NeInt64(22)
		tt.wantQuery = "tbl.field <> ?"
		tt.wantArgs = []interface{}{int64(22)}
		assert(t, tt)
	})

	t.Run("NumberField GtInt64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).GtInt64(22)
		tt.wantQuery = "tbl.field > ?"
		tt.wantArgs = []interface{}{int64(22)}
		assert(t, tt)
	})

	t.Run("NumberField GeInt64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).GeInt64(22)
		tt.wantQuery = "tbl.field >= ?"
		tt.wantArgs = []interface{}{int64(22)}
		assert(t, tt)
	})

	t.Run("NumberField LtInt64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).LtInt64(22)
		tt.wantQuery = "tbl.field < ?"
		tt.wantArgs = []interface{}{int64(22)}
		assert(t, tt)
	})

	t.Run("NumberField LeInt64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).LeInt64(22)
		tt.wantQuery = "tbl.field <= ?"
		tt.wantArgs = []interface{}{int64(22)}
		assert(t, tt)
	})

	t.Run("NumberField EqFloat64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).EqFloat64(3.14)
		tt.wantQuery = "tbl.field = ?"
		tt.wantArgs = []interface{}{float64(3.14)}
		assert(t, tt)
	})

	t.Run("NumberField NeFloat64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).NeFloat64(3.14)
		tt.wantQuery = "tbl.field <> ?"
		tt.wantArgs = []interface{}{float64(3.14)}
		assert(t, tt)
	})

	t.Run("NumberField GtFloat64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).GtFloat64(3.14)
		tt.wantQuery = "tbl.field > ?"
		tt.wantArgs = []interface{}{float64(3.14)}
		assert(t, tt)
	})

	t.Run("NumberField GeFloat64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).GeFloat64(3.14)
		tt.wantQuery = "tbl.field >= ?"
		tt.wantArgs = []interface{}{float64(3.14)}
		assert(t, tt)
	})

	t.Run("NumberField LtFloat64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).LtFloat64(3.14)
		tt.wantQuery = "tbl.field < ?"
		tt.wantArgs = []interface{}{float64(3.14)}
		assert(t, tt)
	})

	t.Run("NumberField LeFloat64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).LeFloat64(3.14)
		tt.wantQuery = "tbl.field <= ?"
		tt.wantArgs = []interface{}{float64(3.14)}
		assert(t, tt)
	})

	t.Run("NumberField SetInt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).SetInt(22)
		tt.wantQuery = "tbl.field = ?"
		tt.wantArgs = []interface{}{22}
		assert(t, tt)
	})

	t.Run("NumberField SetInt64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).SetInt64(22)
		tt.wantQuery = "tbl.field = ?"
		tt.wantArgs = []interface{}{int64(22)}
		assert(t, tt)
	})

	t.Run("NumberField SetFloat64", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewNumberField("field", TableInfo{TableName: "tbl"}).SetFloat64(3.14)
		tt.wantQuery = "tbl.field = ?"
		tt.wantArgs = []interface{}{float64(3.14)}
		assert(t, tt)
	})
}
