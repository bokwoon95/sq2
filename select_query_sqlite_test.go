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
	db, err := sql.Open("sqlite3", "/Users/bokwoon/Documents/sq2/db.sqlite3")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}

	t.Run("Q1", func(t *testing.T) {
		t.Parallel()
		var answer1 []string
		ACTOR := xNEW_ACTOR("")
		_, err := Fetch(db, SQLite.
			SelectDistinct().
			From(ACTOR).
			OrderBy(ACTOR.LAST_NAME).
			Limit(5),
			func(row *Row) {
				lastName := row.String(ACTOR.LAST_NAME)
				row.Process(func() { answer1 = append(answer1, lastName) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(answer1, sakilaAnswer1()); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q2", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		answer2, err := FetchExists(Log(db), SQLite.
			From(ACTOR).
			Where(Or(
				ACTOR.FIRST_NAME.EqString("SCARLETT"),
				ACTOR.LAST_NAME.EqString("JOHANSSON"),
			)),
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(answer2, sakilaAnswer2()); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q3", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var answer3 int
		_, err := Fetch(db, SQLite.From(ACTOR), func(row *Row) {
			answer3 = row.Int(NumberFieldf("COUNT(DISTINCT {})", ACTOR.LAST_NAME))
			row.Close()
		})
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(answer3, sakilaAnswer3()); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q4", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var answer4 []Actor
		_, err := Fetch(db, SQLite.
			From(ACTOR).
			Where(ACTOR.LAST_NAME.LikeString("%GEN%")).
			OrderBy(ACTOR.ACTOR_ID),
			func(row *Row) {
				actor := Actor{
					ActorID:    row.Int(ACTOR.ACTOR_ID),
					FirstName:  row.String(ACTOR.FIRST_NAME),
					LastName:   row.String(ACTOR.LAST_NAME),
					LastUpdate: row.Time(ACTOR.LAST_UPDATE),
				}
				row.Process(func() { answer4 = append(answer4, actor) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(answer4, sakilaAnswer4()); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q5", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var answer5 []string
		_, err := Fetch(db, SQLite.
			From(ACTOR).
			GroupBy(ACTOR.LAST_NAME).
			Having(Fieldf("COUNT(*)").Eq(1)).
			OrderBy(ACTOR.LAST_NAME).
			Limit(5),
			func(row *Row) {
				lastName := row.String(ACTOR.LAST_NAME)
				row.Process(func() { answer5 = append(answer5, lastName) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(answer5, sakilaAnswer5()); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})
}
