package sq

import (
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
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
		tt.item = Postgres.
			DeleteFrom(ACTOR).
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

	t.Run("delete with join", func(t *testing.T) {
		t.Parallel()
		var tt TT
		FILM, LANGUAGE, INVENTORY := xNEW_FILM("f"), xNEW_LANGUAGE("l"), xNEW_INVENTORY("i")
		lang := NewCTE("lang", nil, Postgres.
			Select(LANGUAGE.LANGUAGE_ID, LANGUAGE.NAME).
			From(LANGUAGE).
			Where(LANGUAGE.NAME.IsNotNull()),
		)
		tt.item = Postgres.
			DeleteWith(lang).
			DeleteFrom(FILM).
			Using(lang).
			Join(INVENTORY, INVENTORY.FILM_ID.Eq(FILM.FILM_ID)).
			Where(
				lang.Field("language_id").Eq(FILM.LANGUAGE_ID),
				lang.Field("name").In([]string{"English", "Italian"}),
				INVENTORY.LAST_UPDATE.IsNotNull(),
			).
			Returning(FILM.FILM_ID)
		tt.wantQuery = "WITH lang AS (" +
			"SELECT l.language_id, l.name FROM language AS l WHERE l.name IS NOT NULL" +
			")" +
			" DELETE FROM film AS f" +
			" USING lang" +
			" JOIN inventory AS i ON i.film_id = f.film_id" +
			" WHERE lang.language_id = f.language_id AND lang.name IN ($1, $2) AND i.last_update IS NOT NULL" +
			" RETURNING f.film_id"
		tt.wantArgs = []interface{}{"English", "Italian"}
		assert(t, tt)
	})
}

func TestPostgresSakilaDelete(t *testing.T) {
	if testing.Short() {
		return
	}
	tx, err := postgresDB.Begin()
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	defer tx.Rollback()

	// delete address with address_id 617
	ADDRESS := xNEW_ADDRESS("")
	CITY := xNEW_CITY("")
	COUNTRY := xNEW_COUNTRY("")
	rowsAffected, _, err := Exec(Log(tx), Postgres.DeleteFrom(ADDRESS).Where(ADDRESS.ADDRESS_ID.EqInt(617)), ErowsAffected)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 1 {
		t.Fatalf(testutil.Callers()+"expected 1 row to be affected but got %d", rowsAffected)
	}

	// make sure address was deleted
	exists, err := FetchExists(Log(tx), Postgres.From(ADDRESS).Where(ADDRESS.ADDRESS_ID.EqInt(617)))
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if exists {
		t.Fatal(testutil.Callers(), "address_id 617 was not successfully deleted")
	}

	// delete all addresses with country 'Singapore'
	var addressIDs []int
	_, err = Fetch(Log(tx), Postgres.
		DeleteFrom(ADDRESS).
		Using(CITY).
		Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
		Where(
			CITY.CITY_ID.Eq(ADDRESS.CITY_ID),
			COUNTRY.COUNTRY.EqString("Singapore"),
		),
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
	exists, err = FetchExists(Log(tx), Postgres.
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
