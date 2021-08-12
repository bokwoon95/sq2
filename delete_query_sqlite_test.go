package sq

import (
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
)

func Test_SQLiteDeleteQuery(t *testing.T) {
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

	t.Run("filler", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = SQLite(nil).
			DeleteFrom(ACTOR).
			DeleteFrom(ACTOR).
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1")))
		tt.wantQuery = "WITH cte (n) AS (SELECT 1) DELETE FROM actor AS a"
		assert(t, tt)
	})

	t.Run("delete with join", func(t *testing.T) {
		t.Parallel()
		var tt TT
		FILM1, FILM2, LANGUAGE, INVENTORY := xNEW_FILM("f1"), xNEW_FILM("f2"), xNEW_LANGUAGE("l"), xNEW_INVENTORY("i")
		lang := NewCTE("lang", nil, SQLite(nil).
			Select(LANGUAGE.LANGUAGE_ID, LANGUAGE.NAME).
			From(LANGUAGE).
			Where(LANGUAGE.NAME.IsNotNull()),
		)
		tt.item = SQLite(nil).
			DeleteWith(lang).
			DeleteFrom(FILM1).
			Where(Exists(SQLite(nil).
				SelectOne().
				From(FILM2).
				Join(lang, lang.Field("language_id").Eq(FILM2.LANGUAGE_ID)).
				Join(INVENTORY, INVENTORY.FILM_ID.Eq(FILM2.FILM_ID)).
				Where(
					FILM1.FILM_ID.Eq(FILM2.FILM_ID),
					lang.Field("name").In([]string{"English", "Italian"}),
					INVENTORY.LAST_UPDATE.IsNotNull(),
				),
			)).
			Returning(FILM1.FILM_ID)
		tt.wantQuery = "WITH lang AS (" +
			"SELECT l.language_id, l.name FROM language AS l WHERE l.name IS NOT NULL" +
			")" +
			" DELETE FROM film AS f1" +
			" WHERE EXISTS (" +
			"SELECT 1" +
			" FROM film AS f2" +
			" JOIN lang ON lang.language_id = f2.language_id" +
			" JOIN inventory AS i ON i.film_id = f2.film_id" +
			" WHERE f1.film_id = f2.film_id AND lang.name IN ($1, $2) AND i.last_update IS NOT NULL" +
			")" +
			" RETURNING f1.film_id"
		tt.wantArgs = []interface{}{"English", "Italian"}
		assert(t, tt)
	})

	t.Run("ORDER BY LIMIT OFFSET", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = SQLite(nil).DeleteFrom(ACTOR).OrderBy(ACTOR.ACTOR_ID).Limit(0).Offset(10)
		tt.wantQuery = "DELETE FROM actor AS a ORDER BY a.actor_id LIMIT $1 OFFSET $2"
		tt.wantArgs = []interface{}{int64(0), int64(10)}
		assert(t, tt)
	})
}
