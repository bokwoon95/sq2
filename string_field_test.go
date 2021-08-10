package sq

import (
	"bytes"
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
)

func Test_StringField(t *testing.T) {
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
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(buf.String(), tt.wantQuery); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
		if diff := testutil.Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
		if tt.wantParams != nil {
			if diff := testutil.Diff(gotParams, tt.wantParams); diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		}
	}

	t.Run("StringField", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewStringField("field", TableInfo{TableName: "tbl"})
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("StringField with alias", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewStringField("field", TableInfo{TableName: "tbl"}).As("f")
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("StringField ASC NULLS FIRST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewStringField("field", TableInfo{TableName: "tbl"}).Asc().NullsFirst()
		tt.wantQuery = "tbl.field ASC NULLS FIRST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("StringField DESC NULLS LAST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewStringField("field", TableInfo{TableName: "tbl"}).Desc().NullsLast()
		tt.wantQuery = "tbl.field DESC NULLS LAST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("StringField IS NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewStringField("field", TableInfo{TableName: "tbl"}).IsNull()
		tt.wantQuery = "tbl.field IS NULL"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("StringField IS NOT NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewStringField("field", TableInfo{TableName: "tbl"}).IsNotNull()
		tt.wantQuery = "tbl.field IS NOT NULL"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("StringFieldf in (slice)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = StringFieldf(
			"{name} || {delim} || {email}",
			Param("name", USERS.NAME),
			Param("delim", ": "),
			Param("email", USERS.EMAIL),
		).In([]string{"a", "b", "c"})
		tt.wantQuery = "name || ? || email IN (?, ?, ?)"
		tt.wantArgs = []interface{}{": ", "a", "b", "c"}
		tt.wantParams = map[string][]int{"delim": {0}}
		assert(t, tt)
	})

	t.Run("StringFieldf in (rowvalue)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = StringFieldf(
			"{name} || {delim} || {email}",
			Param("name", USERS.NAME),
			Param("delim", ": "),
			Param("email", USERS.EMAIL),
		).In(RowValue{USERS.NAME, USERS.EMAIL})
		tt.wantQuery = "name || ? || email IN (name, email)"
		tt.wantArgs = []interface{}{": "}
		tt.wantParams = map[string][]int{"delim": {0}}
		assert(t, tt)
	})

	t.Run("StringField Eq", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewStringField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Eq(field)
		tt.wantQuery = "tbl.field = tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("StringField Ne", func(t *testing.T) {
		t.Parallel()
		var tt TT
		field := NewStringField("field", TableInfo{TableName: "tbl"})
		tt.item = field.Ne(field)
		tt.wantQuery = "tbl.field <> tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("StringField EqString", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewStringField("field", TableInfo{TableName: "tbl"}).EqString("abc")
		tt.wantQuery = "tbl.field = ?"
		tt.wantArgs = []interface{}{"abc"}
		assert(t, tt)
	})

	t.Run("StringField NeString", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewStringField("field", TableInfo{TableName: "tbl"}).NeString("abc")
		tt.wantQuery = "tbl.field <> ?"
		tt.wantArgs = []interface{}{"abc"}
		assert(t, tt)
	})

	t.Run("StringField SetString", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewStringField("field", TableInfo{TableName: "tbl"}).SetString("abc")
		tt.wantQuery = "tbl.field = ?"
		tt.wantArgs = []interface{}{"abc"}
		assert(t, tt)
	})
}
