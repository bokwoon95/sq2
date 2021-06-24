package sq

import (
	"bytes"
	"testing"
)

func Test_BlobField(t *testing.T) {
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
		if diff := testdiff(tt.wantQuery, buf.String()); diff != "" {
			t.Fatal(testcallers(), diff)
		}
		if diff := testdiff(tt.wantArgs, gotArgs); diff != "" {
			t.Fatal(testcallers(), diff)
		}
	}

	t.Run("BlobField", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewBlobField("field", TableInfo{TableName: "tbl"})
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("BlobField with alias", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewBlobField("field", TableInfo{TableName: "tbl"}).As("f")
		tt.wantQuery = "tbl.field"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("ASC NULLS FIRST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewBlobField("field", TableInfo{TableName: "tbl"}).Asc().NullsFirst()
		tt.wantQuery = "tbl.field ASC NULLS FIRST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("DESC NULLS LAST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewBlobField("field", TableInfo{TableName: "tbl"}).Desc().NullsLast()
		tt.wantQuery = "tbl.field DESC NULLS LAST"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("BlobField IS NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewBlobField("field", TableInfo{TableName: "tbl"}).IsNull()
		tt.wantQuery = "tbl.field IS NULL"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("BlobField IS NOT NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewBlobField("field", TableInfo{TableName: "tbl"}).IsNotNull()
		tt.wantQuery = "tbl.field IS NOT NULL"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("SetBlob", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewBlobField("field", TableInfo{TableName: "tbl"}).SetBlob([]byte{'a', 'b', 'c', 'd'})
		tt.excludedTableQualifiers = []string{"tbl"}
		tt.wantQuery = "field = ?"
		tt.wantArgs = []interface{}{[]byte{'a', 'b', 'c', 'd'}}
		assert(t, tt)
	})
}
