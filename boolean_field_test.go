package sq

import (
	"bytes"
	"testing"

	"github.com/bokwoon95/testutil"
)

func Test_BooleanField(t *testing.T) {
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
		err := tt.item.AppendSQLExclude(tt.dialect, buf, &gotArgs, gotParams, tt.excludedTableQualifiers)
		is.NoErr(err)
		is.Equal(tt.wantQuery, buf.String())
	}

	t.Run("BooleanField", func(t *testing.T) {
		var tt TT
		tt.item = NewBooleanField("field", GenericTable{TableName: "tbl"})
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("BooleanField with alias", func(t *testing.T) {
		var tt TT
		tt.item = NewBooleanField("field", GenericTable{TableName: "tbl"}).As("f")
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("BooleanField ASC NULLS LAST", func(t *testing.T) {
		var tt TT
		tt.item = NewBooleanField("field", GenericTable{TableName: "tbl"}).Asc().NullsLast()
		tt.wantQuery = "tbl.field ASC NULLS LAST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("BooleanField DESC NULLS FIRST", func(t *testing.T) {
		var tt TT
		tt.item = NewBooleanField("field", GenericTable{TableName: "tbl"}).Desc().NullsFirst()
		tt.wantQuery = "tbl.field DESC NULLS FIRST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("BooleanField NOT", func(t *testing.T) {
		var tt TT
		tt.item = NewBooleanField("field", GenericTable{TableName: "tbl"}).Not()
		tt.wantQuery = "NOT tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("BooleanField IS NULL", func(t *testing.T) {
		var tt TT
		tt.item = NewBooleanField("field", GenericTable{TableName: "tbl"}).IsNull()
		tt.wantQuery = "tbl.field IS NULL"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("BooleanField IS NOT NULL", func(t *testing.T) {
		var tt TT
		tt.item = NewBooleanField("field", GenericTable{TableName: "tbl"}).IsNotNull()
		tt.wantQuery = "tbl.field IS NOT NULL"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("BooleanField Eq", func(t *testing.T) {
		var tt TT
		f := NewBooleanField("field", GenericTable{TableName: "tbl"})
		tt.item = f.Eq(f)
		tt.wantQuery = "tbl.field = tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("BooleanField Ne", func(t *testing.T) {
		var tt TT
		f := NewBooleanField("field", GenericTable{TableName: "tbl"})
		tt.item = f.Ne(f)
		tt.wantQuery = "tbl.field <> tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("BooleanField EqBool", func(t *testing.T) {
		var tt TT
		tt.item = NewBooleanField("field", GenericTable{TableName: "tbl"}).EqBool(true)
		tt.wantQuery = "tbl.field = ?"
		tt.wantArgs = []interface{}{true}
		assert(t, tt)
	})

	t.Run("BooleanField NeBool", func(t *testing.T) {
		var tt TT
		tt.item = NewBooleanField("field", GenericTable{TableName: "tbl"}).NeBool(true)
		tt.wantQuery = "tbl.field <> ?"
		tt.wantArgs = []interface{}{true}
		assert(t, tt)
	})

	t.Run("BooleanField SetBool", func(t *testing.T) {
		var tt TT
		tt.item = NewBooleanField("field", GenericTable{TableName: "tbl"}).SetBool(true)
		tt.wantQuery = "tbl.field = ?"
		tt.wantArgs = []interface{}{true}
		assert(t, tt)
	})
}
