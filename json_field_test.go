package sq

import (
	"bytes"
	"testing"
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
	}

	t.Run("JSONField", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewJSONField("field", TableInfo{TableName: "tbl"})
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("JSONField with alias", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewJSONField("field", TableInfo{TableName: "tbl"}).As("f")
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("JSONField ASC NULLS LAST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewJSONField("field", TableInfo{TableName: "tbl"}).Asc().NullsLast()
		tt.wantQuery = "tbl.field ASC NULLS LAST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("JSONField DESC NULLS FIRST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewJSONField("field", TableInfo{TableName: "tbl"}).Desc().NullsFirst()
		tt.wantQuery = "tbl.field DESC NULLS FIRST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})
}
