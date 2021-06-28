package sq_test

import (
	"testing"

	. "github.com/bokwoon95/sq"
)

func Test_MySQLSelectQuery(t *testing.T) {
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

	t.Run("filler v1", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = MySQL.
			SelectWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Select(ACTOR.ACTOR_ID).
			From(ACTOR)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1) SELECT a.actor_id FROM actor AS a"
		assert(t, tt)
	})

	t.Run("filler v2", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = MySQL.
			SelectDistinct(ACTOR.ACTOR_ID).
			SelectDistinct(ACTOR.ACTOR_ID).
			From(ACTOR)
		tt.wantQuery = "SELECT DISTINCT a.actor_id FROM actor AS a"
		assert(t, tt)
	})

	t.Run("joins", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = MySQL.
			SelectOne().
			From(ACTOR).
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Join(ACTOR, Eq(1, 1)).
			LeftJoin(ACTOR, Eq(1, 1)).
			RightJoin(ACTOR, Eq(1, 1)).
			FullJoin(ACTOR, Eq(1, 1)).
			CrossJoin(ACTOR).
			CustomJoin("CROSS JOIN LATERAL", ACTOR)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" SELECT 1" +
			" FROM actor AS a" +
			" JOIN actor AS a ON ? = ?" +
			" LEFT JOIN actor AS a ON ? = ?" +
			" RIGHT JOIN actor AS a ON ? = ?" +
			" FULL JOIN actor AS a ON ? = ?" +
			" CROSS JOIN actor AS a" +
			" CROSS JOIN LATERAL actor AS a"
		tt.wantArgs = []interface{}{1, 1, 1, 1, 1, 1, 1, 1}
		assert(t, tt)
	})

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = MySQL.
			From(ACTOR).
			Select(ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			GroupBy(ACTOR.FIRST_NAME).
			Having(ACTOR.FIRST_NAME.IsNotNull()).
			OrderBy(ACTOR.LAST_NAME).
			Limit(10).
			Offset(20)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" SELECT a.actor_id, a.first_name, a.last_name" +
			" FROM actor AS a" +
			" GROUP BY a.first_name" +
			" HAVING a.first_name IS NOT NULL" +
			" ORDER BY a.last_name" +
			" LIMIT ?" +
			" OFFSET ?"
		tt.wantArgs = []interface{}{int64(10), int64(20)}
		assert(t, tt)
	})
}
