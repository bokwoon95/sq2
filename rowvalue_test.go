package sq

import (
	"bytes"
	"testing"

	"github.com/bokwoon95/testutil"
)

func TestRowValue(t *testing.T) {
	USERS := struct {
		tmptable
		USER_ID tmpfield
		NAME    tmpfield
		EMAIL   tmpfield
		AGE     tmpfield
	}{
		tmptable: "users",
		USER_ID:  [2]string{"", "user_id"},
		NAME:     [2]string{"", "name"},
		EMAIL:    [2]string{"", "email"},
		AGE:      [2]string{"", "age"},
	}

	type TT struct {
		dialect   string
		value     SQLAppender
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
		err := tt.value.AppendSQL(tt.dialect, buf, &gotArgs, gotParams)
		is.NoErr(err)
	}

	t.Run("rowvalue", func(t *testing.T) {
		var tt TT
		tt.value = RowValue{1, 2, USERS.USER_ID, 4}
		tt.wantQuery = "(?, ?, user_id, ?)"
		tt.wantArgs = []interface{}{1, 2, 4}
		assert(t, tt)
	})

	t.Run("rowvalues", func(t *testing.T) {
		var tt TT
		tt.value = RowValues{
			{1, 2, USERS.USER_ID, 4},
			{USERS.NAME, 2, 3, 4},
			{1, 2, 3, USERS.AGE},
		}
		tt.wantQuery = "(?, ?, user_id, ?), (name, ?, ?, ?), (?, ?, ?, age)"
		tt.wantArgs = []interface{}{1, 2, 4, 2, 3, 4, 1, 2, 3}
		assert(t, tt)
	})
}
