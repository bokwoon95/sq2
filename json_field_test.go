package sq

import (
	"bytes"
	"testing"

	"github.com/bokwoon95/testutil"
)

func Test_JSONField(t *testing.T) {
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

	t.Run("JSONField", func(t *testing.T) {
		var tt TT
		tt.item = NewJSONField("field", GenericTable{TableName: "tbl"})
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("JSONField with alias", func(t *testing.T) {
		var tt TT
		tt.item = NewJSONField("field", GenericTable{TableName: "tbl"}).As("f")
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("JSONField ASC NULLS LAST", func(t *testing.T) {
		var tt TT
		tt.item = NewJSONField("field", GenericTable{TableName: "tbl"}).Asc().NullsLast()
		tt.wantQuery = "tbl.field ASC NULLS LAST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("JSONField DESC NULLS FIRST", func(t *testing.T) {
		var tt TT
		tt.item = NewJSONField("field", GenericTable{TableName: "tbl"}).Desc().NullsFirst()
		tt.wantQuery = "tbl.field DESC NULLS FIRST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})
}
