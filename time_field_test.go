package sq

import (
	"bytes"
	"testing"
	"time"
)

func Test_TimeField(t *testing.T) {
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

	timeval := time.Unix(0, 0)

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
			if diff := testdiff(gotParams, tt.wantParams); diff != "" {
				t.Error(testcallers(), diff)
			}
		}
	}

	t.Run("TimeField", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"})
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("TimeField with alias", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"}).As("f")
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("TimeField ASC NULLS FIRST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"}).Asc().NullsFirst()
		tt.wantQuery = "tbl.field ASC NULLS FIRST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("TimeField DESC NULLS LAST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"}).Desc().NullsLast()
		tt.wantQuery = "tbl.field DESC NULLS LAST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("TimeField IS NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"}).IsNull()
		tt.wantQuery = "tbl.field IS NULL"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("TimeField IS NOT NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"}).IsNotNull()
		tt.wantQuery = "tbl.field IS NOT NULL"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("TimeFieldf in (slice)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = TimeFieldf(
			"date_trunc('day', {timestamp})",
			Param("timestamp", timeval),
		).In([]time.Time{timeval, timeval, timeval})
		tt.wantQuery = "date_trunc('day', ?) IN (?, ?, ?)"
		tt.wantArgs = []interface{}{timeval, timeval, timeval, timeval}
		tt.wantParams = map[string][]int{"timestamp": {0}}
		assert(t, tt)
	})

	t.Run("TimeFieldf in (rowvalue)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = TimeFieldf(
			"date_trunc('day', {timestamp})",
			Param("timestamp", timeval),
		).In(RowValue{USERS.NAME, USERS.EMAIL})
		tt.wantQuery = "date_trunc('day', ?) IN (name, email)"
		tt.wantArgs = []interface{}{timeval}
		tt.wantParams = map[string][]int{"timestamp": {0}}
		assert(t, tt)
	})

	t.Run("TimeField Eq", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewTimeField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Eq(field)
		tt.wantQuery = "tbl.field = tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("TimeField Ne", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewTimeField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Ne(field)
		tt.wantQuery = "tbl.field <> tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("TimeField Gt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewTimeField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Gt(field)
		tt.wantQuery = "tbl.field > tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("TimeField Ge", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewTimeField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Ge(field)
		tt.wantQuery = "tbl.field >= tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("TimeField Lt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewTimeField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Lt(field)
		tt.wantQuery = "tbl.field < tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("TimeField Le", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewTimeField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Le(field)
		tt.wantQuery = "tbl.field <= tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("TimeField EqTime", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"}).EqTime(timeval)
		tt.wantQuery = "tbl.field = ?"
		tt.wantArgs = []interface{}{timeval}
		assert(t, tt)
	})

	t.Run("TimeField NeTime", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"}).NeTime(timeval)
		tt.wantQuery = "tbl.field <> ?"
		tt.wantArgs = []interface{}{timeval}
		assert(t, tt)
	})

	t.Run("TimeField GtTime", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"}).GtTime(timeval)
		tt.wantQuery = "tbl.field > ?"
		tt.wantArgs = []interface{}{timeval}
		assert(t, tt)
	})

	t.Run("TimeField GeTime", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"}).GeTime(timeval)
		tt.wantQuery = "tbl.field >= ?"
		tt.wantArgs = []interface{}{timeval}
		assert(t, tt)
	})

	t.Run("TimeField LtTime", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"}).LtTime(timeval)
		tt.wantQuery = "tbl.field < ?"
		tt.wantArgs = []interface{}{timeval}
		assert(t, tt)
	})

	t.Run("TimeField LeTime", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"}).LeTime(timeval)
		tt.wantQuery = "tbl.field <= ?"
		tt.wantArgs = []interface{}{timeval}
		assert(t, tt)
	})

	t.Run("TimeField SetTime", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewTimeField("field", TableInfo{TableName: "tbl"}).SetTime(timeval)
		tt.wantQuery = "tbl.field = ?"
		tt.wantArgs = []interface{}{timeval}
		assert(t, tt)
	})
}
