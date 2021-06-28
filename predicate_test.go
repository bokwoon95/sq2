package sq

import (
	"bytes"
	"testing"
)

func TestVariadicPredicate(t *testing.T) {
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
		dialect                 string
		predicate               Predicate
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
		err := tt.predicate.AppendSQLExclude(tt.dialect, buf, &gotArgs, gotParams, tt.excludedTableQualifiers)
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

	assertError := func(t *testing.T, tt TT) {
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, params := []interface{}{}, map[string][]int{}
		err := tt.predicate.AppendSQLExclude(tt.dialect, buf, &gotArgs, params, tt.excludedTableQualifiers)
		if err == nil {
			t.Fatal(Callers(), "expected error but got nil")
		}
	}

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.predicate = VariadicPredicate{}
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("nil predicate", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.predicate = And(Predicate(nil))
		assertError(t, tt)
	})

	t.Run("1 predicate", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.predicate = And(Not(Eq(USERS.NAME, 21)))
		tt.wantQuery = "NOT name = ?"
		tt.wantArgs = []interface{}{21}
		assert(t, tt)
	})

	t.Run("nested variadic predicate", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.predicate = And(And(And(And(Eq(USERS.NAME, 21)))))
		tt.wantQuery = "(name = ?)"
		tt.wantArgs = []interface{}{21}
		assert(t, tt)
	})

	t.Run("multiple predicates", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.predicate = Or(
			IsNull(USERS.NAME),
			IsNotNull(USERS.NAME),
			Not(Eq(USERS.AGE, USERS.AGE)),
			Not(And(Eq(USERS.AGE, USERS.AGE))),
			Not(And(
				Eq(USERS.USER_ID, 1),
				Ne(USERS.USER_ID, 2),
				Gt(USERS.USER_ID, 3),
				Ge(USERS.USER_ID, 4),
				Lt(USERS.USER_ID, 5),
				Le(USERS.USER_ID, 6),
			)),
		)
		tt.wantQuery = "(name IS NULL" +
			" OR name IS NOT NULL" +
			" OR NOT age = age" +
			" OR NOT age = age" +
			" OR NOT (" +
			"user_id = ?" +
			" AND user_id <> ?" +
			" AND user_id > ?" +
			" AND user_id >= ?" +
			" AND user_id < ?" +
			" AND user_id <= ?" +
			"))"
		tt.wantArgs = []interface{}{1, 2, 3, 4, 5, 6}
		assert(t, tt)
	})

	t.Run("multiple predicates with nil", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.predicate = Or(
			Predicate(nil),
			Not(And(
				Predicate(nil),
				Predicate(nil),
			)),
		)
		assertError(t, tt)
	})
}
