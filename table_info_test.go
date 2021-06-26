package sq

import (
	"bytes"
	"testing"
)

func Test_TableInfo(t *testing.T) {
	type TT struct {
		dialect   string
		tbl       TableInfo
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, gotParams := []interface{}{}, map[string][]int{}
		err := tt.tbl.AppendSQL(tt.dialect, buf, &gotArgs, gotParams)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(buf.String(), tt.wantQuery); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.tbl.TableSchema, tt.tbl.GetSchema()); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.tbl.TableName, tt.tbl.GetName()); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.tbl.TableAlias, tt.tbl.GetAlias()); diff != "" {
			t.Error(testcallers(), diff)
		}
	}

	t.Run("with schema", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.tbl = TableInfo{
			TableSchema: "public",
			TableName:   "users",
		}
		tt.wantQuery = "public.users"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("without schema", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.tbl = TableInfo{
			TableName: "users",
		}
		tt.wantQuery = "users"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})
}
