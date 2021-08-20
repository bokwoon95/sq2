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
		tt.item = SQLite.
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
		lang := NewCTE("lang", nil, SQLite.
			Select(LANGUAGE.LANGUAGE_ID, LANGUAGE.NAME).
			From(LANGUAGE).
			Where(LANGUAGE.NAME.IsNotNull()),
		)
		tt.item = SQLite.
			DeleteWith(lang).
			DeleteFrom(FILM1).
			Where(Exists(SQLite.
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
		tt.item = SQLite.DeleteFrom(ACTOR).OrderBy(ACTOR.ACTOR_ID).Limit(0).Offset(10)
		tt.wantQuery = "DELETE FROM actor AS a ORDER BY a.actor_id LIMIT $1 OFFSET $2"
		tt.wantArgs = []interface{}{int64(0), int64(10)}
		assert(t, tt)
	})
}

func TestSQLiteSakilaDelete(t *testing.T) {
	if testing.Short() {
		return
	}
	tx, err := sqliteDB.Begin()
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	defer tx.Rollback()

	// delete address with address_id 617
	ADDRESS := xNEW_ADDRESS("")
	CITY := xNEW_CITY("")
	COUNTRY := xNEW_COUNTRY("")
	rowsAffected, _, err := Exec(Log(tx), SQLite.DeleteFrom(ADDRESS).Where(ADDRESS.ADDRESS_ID.EqInt(617)), ErowsAffected)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 1 {
		t.Fatalf(testutil.Callers()+"expected 1 row to be affected but got %d", rowsAffected)
	}

	// make sure address was deleted
	exists, err := FetchExists(Log(tx), SQLite.From(ADDRESS).Where(ADDRESS.ADDRESS_ID.EqInt(617)))
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if exists {
		t.Fatal(testutil.Callers(), "address_id 617 was not successfully deleted")
	}

	// delete all addresses with country 'Singapore'
	var addressIDs []int
	_, err = Fetch(Log(tx), SQLite.
		DeleteFrom(ADDRESS).
		Where(Exists(SQLite.
			SelectOne().
			From(CITY).
			Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
			Where(
				CITY.CITY_ID.Eq(ADDRESS.CITY_ID),
				COUNTRY.COUNTRY.EqString("Singapore"),
			),
		)),
		func(row *Row) {
			addressID := row.Int(ADDRESS.ADDRESS_ID)
			row.Process(func() { addressIDs = append(addressIDs, addressID) })
		},
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if diff := testutil.Diff(addressIDs, []int{624, 625, 626, 627, 628, 629, 630, 631, 632}); diff != "" {
		t.Fatal(testutil.Callers(), "diff")
	}

	// make sure addresses were deleted
	exists, err = FetchExists(Log(tx), SQLite.
		From(ADDRESS).
		Join(CITY, CITY.CITY_ID.Eq(ADDRESS.CITY_ID)).
		Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
		Where(COUNTRY.COUNTRY.EqString("Singapore")),
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if exists {
		t.Fatal(testutil.Callers(), "addresses with country 'Singapore' were not successfully deleted")
	}
}
