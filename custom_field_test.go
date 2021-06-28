package sq

import (
	"bytes"
	"testing"
)

func Test_CustomField(t *testing.T) {
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
			t.Fatal(Callers(), err)
		}
		if diff := Diff(buf.String(), tt.wantQuery); diff != "" {
			t.Error(Callers(), diff)
		}
		if diff := Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(Callers(), diff)
		}
	}

	t.Run("FieldValue", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = FieldValue("abcd")
		tt.wantQuery = "?"
		tt.wantArgs = []interface{}{"abcd"}
		assert(t, tt)
	})

	t.Run("Fieldf", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("lorem ipsum {} {}", 1, "a")
		tt.wantQuery = "lorem ipsum ? ?"
		tt.wantArgs = []interface{}{1, "a"}
		assert(t, tt)
	})

	t.Run("CustomField alias", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").As("ggggggg")
		tt.wantQuery = "my_field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("CustomField ASC NULLS LAST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").Asc().NullsLast()
		tt.wantQuery = "my_field ASC NULLS LAST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("CustomField DESC NULLS FIRST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").Desc().NullsFirst()
		tt.wantQuery = "my_field DESC NULLS FIRST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("CustomField IS NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").IsNull()
		tt.wantQuery = "my_field IS NULL"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("CustomField IS NOT NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").IsNotNull()
		tt.wantQuery = "my_field IS NOT NULL"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("CustomField IN (rowvalue)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").In(RowValue{USERS.USER_ID, USERS.NAME, USERS.EMAIL})
		tt.wantQuery = "my_field IN (user_id, name, email)"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("CustomField IN (slice)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").In([]int{5, 6, 7})
		tt.wantQuery = "my_field IN (?, ?, ?)"
		tt.wantArgs = []interface{}{5, 6, 7}
		assert(t, tt)
	})

	t.Run("CustomField Eq", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").Eq(123)
		tt.wantQuery = "my_field = ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("CustomField Ne", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").Ne(123)
		tt.wantQuery = "my_field <> ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("CustomField Gt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").Gt(123)
		tt.wantQuery = "my_field > ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("CustomField Ge", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").Ge(123)
		tt.wantQuery = "my_field >= ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("CustomField Lt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").Lt(123)
		tt.wantQuery = "my_field < ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("CustomField Le", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fieldf("my_field").Le(123)
		tt.wantQuery = "my_field <= ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("FieldLiteral", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = FieldLiteral("lorem ipsum dolor sit amet COUNT(*)")
		tt.wantQuery = "lorem ipsum dolor sit amet COUNT(*)"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})
}

func Test_Fields(t *testing.T) {
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
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := ToSQLExclude(tt.dialect, tt.item, tt.excludedTableQualifiers)
		if err != nil {
			t.Fatal(Callers(), err)
		}
		if diff := Diff(gotQuery, tt.wantQuery); diff != "" {
			t.Error(Callers(), diff)
		}
		if diff := Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(Callers(), diff)
		}
	}

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fields{}
		tt.wantQuery = ""
		assert(t, tt)
	})

	t.Run("Fields", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Fields{USERS.USER_ID, nil, FieldValue(456)}
		tt.wantQuery = "user_id, ?, ?"
		tt.wantArgs = []interface{}{nil, 456}
		assert(t, tt)
	})

	t.Run("AliasFields", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AliasFields{USERS.USER_ID.As("uid"), nil, FieldValue(456).As("some_number")}
		tt.wantQuery = "user_id AS uid, ?, ? AS some_number"
		tt.wantArgs = []interface{}{nil, 456}
		assert(t, tt)
	})
}
