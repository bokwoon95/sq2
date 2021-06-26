package sq_test

import (
	"testing"

	. "github.com/bokwoon95/sq"
)

func Test_PostgresDeleteQuery(t *testing.T) {
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

	t.Run("joins", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = Postgres.
			DeleteFrom(ACTOR).
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Using(ACTOR).
			Join(ACTOR, Eq(1, 1)).
			LeftJoin(ACTOR, Eq(1, 1)).
			RightJoin(ACTOR, Eq(1, 1)).
			FullJoin(ACTOR, Eq(1, 1)).
			CrossJoin(ACTOR).
			CustomJoin("NATURAL JOIN", ACTOR)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" DELETE FROM actor AS a" +
			" USING actor AS a" +
			" JOIN actor AS a ON $1 = $2" +
			" LEFT JOIN actor AS a ON $3 = $4" +
			" RIGHT JOIN actor AS a ON $5 = $6" +
			" FULL JOIN actor AS a ON $7 = $8" +
			" CROSS JOIN actor AS a" +
			" NATURAL JOIN actor AS a"
		tt.wantArgs = []interface{}{1, 1, 1, 1, 1, 1, 1, 1}
		assert(t, tt)
	})

	// t.Run("delete with join", func(t *testing.T) {
	// 	t.Parallel()
	// 	var tt TT
	// 	FILM, LANGUAGE := NEW_FILM("f"), NEW_LANGUAGE("l")
	// 	lang := NewCTE("lang", nil, SQLite.
	// 		Select(LANGUAGE.LANGUAGE_ID, LANGUAGE.NAME).
	// 		From(LANGUAGE).
	// 		Where(LANGUAGE.NAME.IsNotNull()),
	// 	)
	// 	tt.item = Postgres.
	// 		DeleteWith(lang).
	// 		DeleteFrom(FILM).
	// 		Using(lang).
	// 		Where()
	// 	tt.wantQuery = "WITH lang AS (" +
	// 		"SELECT l.language_id, l.name FROM language AS l WHERE l.name IS NOT NULL" +
	// 		")" +
	// 		" DELETE FROM film AS f1" +
	// 		" WHERE EXISTS (" +
	// 		"SELECT 1" +
	// 		" FROM film AS f2" +
	// 		" JOIN lang ON lang.language_id = f2.language_id AND f1.film_id = f2.film_id" +
	// 		" WHERE lang.name IN ($1, $2)" +
	// 		")"
	// 	tt.wantArgs = []interface{}{"English", "Italian"}
	// 	assert(t, tt)
	// })
}
