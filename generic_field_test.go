package sq

import (
	"bytes"
	"testing"

	"github.com/bokwoon95/testutil"
)

func Test_GenericField(t *testing.T) {
	type TT struct {
		dialect                 string
		item                    SQLExcludeAppender
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		is := testutil.New(t, testutil.Parallel)
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, gotParams := []interface{}{}, map[string][]int{}
		tt.item.AppendSQLExclude(tt.dialect, buf, &gotArgs, gotParams, tt.excludedTableQualifiers)
		is.Equal(tt.wantQuery, buf.String())
		is.Equal(tt.wantArgs, gotArgs)
	}

	t.Run("GenericField table qualified", func(t *testing.T) {
		var tt TT
		tt.item = GenericField{
			TableName: "users",
			FieldName: "user_id",
		}
		tt.wantQuery = "users.user_id"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("GenericField table alias qualified", func(t *testing.T) {
		var tt TT
		tt.item = GenericField{
			TableName: "users", TableAlias: "u",
			FieldName: "user_id",
		}
		tt.wantQuery = "u.user_id"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("GenericField alias", func(t *testing.T) {
		var tt TT
		tt.item = GenericField{
			TableName: "users", TableAlias: "u",
			FieldName: "user_id", FieldAlias: "uid",
		}
		tt.wantQuery = "u.user_id"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("GenericField excludedTableQualifiers (name)", func(t *testing.T) {
		var tt TT
		tt.item = GenericField{
			TableName: "users",
			FieldName: "user_id",
		}
		tt.excludedTableQualifiers = []string{"users"}
		tt.wantQuery = "user_id"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("GenericField excludedTableQualifiers (alias)", func(t *testing.T) {
		var tt TT
		tt.item = GenericField{
			TableName: "users", TableAlias: "u",
			FieldName: "user_id",
		}
		tt.excludedTableQualifiers = []string{"u"}
		tt.wantQuery = "user_id"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("GenericField ASC NULLS LAST", func(t *testing.T) {
		var tt TT
		tt.item = GenericField{
			TableName: "users",
			FieldName: "user_id",
		}.Asc().NullsLast()
		tt.wantQuery = "users.user_id ASC NULLS LAST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("GenericField DESC NULLS FIRST", func(t *testing.T) {
		var tt TT
		tt.item = GenericField{
			TableName: "users",
			FieldName: "user_id",
		}.Desc().NullsFirst()
		tt.wantQuery = "users.user_id DESC NULLS FIRST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	BAD_TABLE := struct {
		WHITESPACE  GenericField
		UPPERCASE   GenericField
		UNDERSCORE  GenericField
		NUMBER      GenericField
		SPECIALCHAR GenericField
		GOOD_COLUMN GenericField
	}{
		WHITESPACE:  GenericField{TableName: "bad table", FieldName: "some shitty column name with spaces"},
		UPPERCASE:   GenericField{TableName: "bad table", FieldName: "uppercASE"},
		UNDERSCORE:  GenericField{TableName: "bad table", FieldName: "_start_with_underscore"},
		NUMBER:      GenericField{TableName: "bad table", FieldName: "123start_with_number"},
		SPECIALCHAR: GenericField{TableName: "bad table", FieldName: "!@#$%^&*"},
		GOOD_COLUMN: GenericField{TableName: "bad table", FieldName: "s1mple_unquoted"},
	}

	t.Run("default quoted identifiers", func(t *testing.T) {
		var tt TT
		tt.item = Fields{
			BAD_TABLE.WHITESPACE,
			BAD_TABLE.UPPERCASE,
			BAD_TABLE.UNDERSCORE,
			BAD_TABLE.NUMBER,
			BAD_TABLE.SPECIALCHAR,
			BAD_TABLE.GOOD_COLUMN,
		}
		tt.wantQuery = `"bad table"."some shitty column name with spaces"` +
			`, "bad table"."uppercASE"` +
			`, "bad table"."_start_with_underscore"` +
			`, "bad table"."123start_with_number"` +
			`, "bad table"."!@#$%^&*"` +
			`, "bad table".s1mple_unquoted`
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("mysql quoted identifiers", func(t *testing.T) {
		var tt TT
		tt.dialect = DialectMySQL
		tt.item = Fields{
			BAD_TABLE.WHITESPACE,
			BAD_TABLE.UPPERCASE,
			BAD_TABLE.UNDERSCORE,
			BAD_TABLE.NUMBER,
			BAD_TABLE.SPECIALCHAR,
			BAD_TABLE.GOOD_COLUMN,
		}
		tt.wantQuery = "`bad table`.`some shitty column name with spaces`" +
			", `bad table`.`uppercASE`" +
			", `bad table`.`_start_with_underscore`" +
			", `bad table`.`123start_with_number`" +
			", `bad table`.`!@#$%^&*`" +
			", `bad table`.s1mple_unquoted"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("mssql quoted identifiers", func(t *testing.T) {
		var tt TT
		tt.dialect = DialectSQLServer
		tt.item = Fields{
			BAD_TABLE.WHITESPACE,
			BAD_TABLE.UPPERCASE,
			BAD_TABLE.UNDERSCORE,
			BAD_TABLE.NUMBER,
			BAD_TABLE.SPECIALCHAR,
			BAD_TABLE.GOOD_COLUMN,
		}
		tt.wantQuery = "[bad table].[some shitty column name with spaces]" +
			", [bad table].[uppercASE]" +
			", [bad table].[_start_with_underscore]" +
			", [bad table].[123start_with_number]" +
			", [bad table].[!@#$%^&*]" +
			", [bad table].s1mple_unquoted"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})
}
