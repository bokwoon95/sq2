package sq

import (
	"bytes"
	"testing"
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
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, gotParams := []interface{}{}, map[string][]int{}
		err := tt.value.AppendSQL(tt.dialect, buf, &gotArgs, gotParams)
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

	t.Run("rowvalue", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.value = RowValue{1, 2, USERS.USER_ID, 4}
		tt.wantQuery = "(?, ?, user_id, ?)"
		tt.wantArgs = []interface{}{1, 2, 4}
		assert(t, tt)
	})

	t.Run("rowvalues", func(t *testing.T) {
		t.Parallel()
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

	t.Run("rowvalue in rowvalues", func(t *testing.T) {
		t.Parallel()
		var tt TT
		predicate := RowValue{USERS.USER_ID, USERS.NAME}.In(RowValues{
			{1, "abc"},
			{2, "def"},
			{3, "ghi"},
		})
		tt.wantQuery = "(user_id, name) IN ((?, ?), (?, ?), (?, ?))"
		tt.wantArgs = []interface{}{1, "abc", 2, "def", 3, "ghi"}
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, gotParams := []interface{}{}, map[string][]int{}
		err := predicate.AppendSQLExclude(tt.dialect, buf, &gotArgs, gotParams, nil)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(buf.String(), tt.wantQuery); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testcallers(), diff)
		}
	})
}
