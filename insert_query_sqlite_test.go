package sq_test

import (
	"testing"

	. "github.com/bokwoon95/sq"
)

func Test_SQLiteInsertQuery(t *testing.T) {
	type TT struct {
		dialect   string
		item      Query
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := ToSQL(tt.dialect, tt.item)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(gotQuery, tt.wantQuery); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testcallers(), diff)
		}
	}

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = SQLite.
			InsertInto(ACTOR).
			InsertInto(ACTOR).
			Columns(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
			Values("bob", "the builder").
			Values("alice", "in wonderland").
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1")))
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" INSERT INTO actor AS a (first_name, last_name)" +
			" VALUES ($1, $2), ($3, $4)"
		tt.wantArgs = []interface{}{"bob", "the builder", "alice", "in wonderland"}
		assert(t, tt)
	})

	t.Run("INSERT with RETURNING", func(t *testing.T) {
	})

	t.Run("INSERT ignore duplicates", func(t *testing.T) {
	})

	t.Run("upsert", func(t *testing.T) {
	})

	t.Run("INSERT from SELECT", func(t *testing.T) {
	})
}
