package sq

import (
	"database/sql"
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
)

func Test_SQLiteSelectQuery(t *testing.T) {
	type TT struct {
		dialect   string
		item      Query
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := ToSQL(tt.dialect, tt.item)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotQuery, tt.wantQuery); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
		if diff := testutil.Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
	}

	t.Run("joins", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = SQLite.
			From(ACTOR).
			From(ACTOR).
			SelectOne().
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Join(ACTOR, Eq(1, 1)).
			LeftJoin(ACTOR, Eq(1, 1)).
			CrossJoin(ACTOR).
			CustomJoin(",", ACTOR)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" SELECT 1" +
			" FROM actor AS a" +
			" JOIN actor AS a ON $1 = $2" +
			" LEFT JOIN actor AS a ON $3 = $4" +
			" CROSS JOIN actor AS a" +
			" , actor AS a"
		tt.wantArgs = []interface{}{1, 1, 1, 1}
		assert(t, tt)
	})

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = SQLite.
			SelectDistinct(ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
			SelectDistinct(ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
			From(ACTOR).
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			GroupBy(ACTOR.FIRST_NAME).
			Having(ACTOR.FIRST_NAME.IsNotNull()).
			OrderBy(ACTOR.LAST_NAME).
			Limit(10).
			Offset(20)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" SELECT DISTINCT a.actor_id, a.first_name, a.last_name" +
			" FROM actor AS a" +
			" GROUP BY a.first_name" +
			" HAVING a.first_name IS NOT NULL" +
			" ORDER BY a.last_name" +
			" LIMIT $1" +
			" OFFSET $2"
		tt.wantArgs = []interface{}{int64(10), int64(20)}
		assert(t, tt)
	})
}

func Test_SQLiteTestSuite(t *testing.T) {
	answers := NewTestSuiteAnswers()

	t.Run("Q1", func(t *testing.T) {
		t.Parallel()
		db, err := sql.Open("sqlite3", "/Users/bokwoon/Documents/sq2/db.sqlite3")
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		var answer1 []string
		ACTOR := xNEW_ACTOR("")
		q := SQLiteEnv(nil).
			SelectDistinct().
			From(ACTOR).
			OrderBy(ACTOR.LAST_NAME).
			Limit(5)
		_, err = Fetch(db, q,
			func(row *Row) {
				lastName := row.String(ACTOR.LAST_NAME)
				row.Process(func() { answer1 = append(answer1, lastName) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(answers.Answer1, answer1); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q2", func(t *testing.T) {
		t.Parallel()
		db, err := sql.Open("sqlite3", "/Users/bokwoon/Documents/sq2/db.sqlite3")
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		var answer01 []string
		ACTOR := xNEW_ACTOR("")
		_, err = Fetch(db, SQLiteEnv(nil).
			SelectDistinct().
			From(ACTOR).
			OrderBy(ACTOR.LAST_NAME).
			Limit(5),
			func(row *Row) {
				lastName := row.String(ACTOR.LAST_NAME)
				row.Process(func() { answer01 = append(answer01, lastName) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(answers.Answer1, answer01); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})
}
