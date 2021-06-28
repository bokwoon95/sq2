package sq_test

import (
	"testing"

	. "github.com/bokwoon95/sq"
)

func Test_SQLiteUpdateQuery(t *testing.T) {
	type TT struct {
		dialect   string
		item      Query
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := ToSQL(tt.dialect, tt.item)
		if err != nil {
			t.Fatal(Callers(), err)
		}
		if diff := Diff(gotQuery, tt.wantQuery); diff != "" {
			t.Error(Callers(), diff)
		}
		if diff := Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(Callers(), diff)
		}
	}

	t.Run("joins", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = SQLite.
			Update(ACTOR).
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Set(ACTOR.ACTOR_ID.SetInt64(1)).
			From(ACTOR).
			Join(ACTOR, Eq(1, 1)).
			LeftJoin(ACTOR, Eq(1, 1)).
			CrossJoin(ACTOR).
			CustomJoin(",", ACTOR).
			Where(ACTOR.ACTOR_ID.EqInt64(1))
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" UPDATE actor AS a" +
			" SET actor_id = $1" +
			" FROM actor AS a" +
			" JOIN actor AS a ON $2 = $3" +
			" LEFT JOIN actor AS a ON $4 = $5" +
			" CROSS JOIN actor AS a" +
			" , actor AS a" +
			" WHERE a.actor_id = $6"
		tt.wantArgs = []interface{}{int64(1), 1, 1, 1, 1, int64(1)}
		assert(t, tt)
	})

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = SQLite.
			UpdateWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Update(ACTOR).
			Setx(func(c *Column) error {
				c.SetInt64(ACTOR.ACTOR_ID, 1)
				return nil
			}).
			Where(ACTOR.ACTOR_ID.EqInt64(1)).
			OrderBy(ACTOR.ACTOR_ID).
			Limit(10).
			Offset(20)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" UPDATE actor AS a" +
			" SET actor_id = $1" +
			" WHERE a.actor_id = $2" +
			" ORDER BY a.actor_id" +
			" LIMIT $3" +
			" OFFSET $4"
		tt.wantArgs = []interface{}{int64(1), int64(1), int64(10), int64(20)}
		assert(t, tt)
	})
}
