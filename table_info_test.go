package sq

import (
	"bytes"
	"testing"

	"github.com/bokwoon95/testutil"
)

func Test_TableInfo(t *testing.T) {
	type TT struct {
		dialect   string
		tbl       TableInfo
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		is := testutil.New(t, testutil.Parallel)
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, gotParams := []interface{}{}, map[string][]int{}
		err := tt.tbl.AppendSQL(tt.dialect, buf, &gotArgs, gotParams)
		is.NoErr(err)
		is.Equal(tt.wantQuery, buf.String())
		is.Equal(tt.wantArgs, gotArgs)
		is.Equal(tt.tbl.TableSchema, tt.tbl.GetSchema())
		is.Equal(tt.tbl.TableName, tt.tbl.GetName())
		is.Equal(tt.tbl.TableAlias, tt.tbl.GetAlias())
	}

	t.Run("with schema", func(t *testing.T) {
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
		var tt TT
		tt.tbl = TableInfo{
			TableName: "users",
		}
		tt.wantQuery = "users"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})
}
