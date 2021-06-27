package sq_test

import (
	"testing"

	. "github.com/bokwoon95/sq"
)

func Test_MySQLDeleteQuery(t *testing.T) {
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
		ACTOR := NEW_ACTOR("")
		tt.item = MySQL.
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
			" DELETE FROM actor" +
			" USING actor" +
			" JOIN actor ON ? = ?" +
			" LEFT JOIN actor ON ? = ?" +
			" RIGHT JOIN actor ON ? = ?" +
			" FULL JOIN actor ON ? = ?" +
			" CROSS JOIN actor" +
			" NATURAL JOIN actor"
		tt.wantArgs = []interface{}{1, 1, 1, 1, 1, 1, 1, 1}
		assert(t, tt)
	})

	t.Run("delete with join", func(t *testing.T) {
		t.Parallel()
		var tt TT
		FILM, LANGUAGE, INVENTORY := NEW_FILM("f"), NEW_LANGUAGE("l"), NEW_INVENTORY("i")
		lang := NewCTE("lang", nil, MySQL.
			Select(LANGUAGE.LANGUAGE_ID, LANGUAGE.NAME).
			From(LANGUAGE).
			Where(LANGUAGE.NAME.IsNotNull()),
		)
		tt.item = MySQL.
			DeleteWith(lang).
			DeleteFrom(FILM).
			Using(FILM).
			Join(lang, lang.Field("language_id").Eq(FILM.LANGUAGE_ID)).
			Join(INVENTORY, INVENTORY.FILM_ID.Eq(FILM.FILM_ID)).
			Where(
				lang.Field("name").In([]string{"English", "Italian"}),
				INVENTORY.LAST_UPDATE.IsNotNull(),
			)
		tt.wantQuery = "WITH lang AS (" +
			"SELECT l.language_id, l.name FROM language AS l WHERE l.name IS NOT NULL" +
			")" +
			" DELETE FROM f" +
			" USING film AS f" +
			" JOIN lang ON lang.language_id = f.language_id" +
			" JOIN inventory AS i ON i.film_id = f.film_id" +
			" WHERE lang.name IN (?, ?) AND i.last_update IS NOT NULL"
		tt.wantArgs = []interface{}{"English", "Italian"}
		assert(t, tt)
	})
}
