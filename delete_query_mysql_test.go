package sq

import (
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
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
		ACTOR := xNEW_ACTOR("")
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
		FILM, LANGUAGE, INVENTORY := xNEW_FILM("f"), xNEW_LANGUAGE("l"), xNEW_INVENTORY("i")
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

	t.Run("Multi-table DELETE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ADDRESS, CITY, COUNTRY := xNEW_ADDRESS("a"), xNEW_CITY("ci"), xNEW_COUNTRY("co")
		tt.item = MySQL.
			DeleteFrom(ADDRESS, CITY, COUNTRY).
			Using(ADDRESS).
			Join(CITY, CITY.CITY_ID.Eq(ADDRESS.CITY_ID)).
			Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
			Where(COUNTRY.COUNTRY_ID.EqInt(112))
		tt.wantQuery = "DELETE FROM a, ci, co" +
			" USING address AS a" +
			" JOIN city AS ci ON ci.city_id = a.city_id" +
			" JOIN country AS co ON co.country_id = ci.country_id" +
			" WHERE co.country_id = ?"
		tt.wantArgs = []interface{}{112}
		assert(t, tt)
	})

	t.Run("ORDER BY LIMIT", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = MySQL.DeleteFrom(ACTOR).OrderBy(ACTOR.ACTOR_ID).Limit(0)
		tt.wantQuery = "DELETE FROM actor AS a ORDER BY a.actor_id LIMIT ?"
		tt.wantArgs = []interface{}{int64(0)}
		assert(t, tt)
	})
}

func TestMySQLSakilaDelete(t *testing.T) {
	if testing.Short() {
		return
	}
	tx, err := mysqlDB.Begin()
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	defer tx.Rollback()

	// delete address with address_id 617
	ADDRESS := xNEW_ADDRESS("")
	CITY := xNEW_CITY("")
	COUNTRY := xNEW_COUNTRY("")
	rowsAffected, _, err := Exec(Log(tx), MySQL.DeleteFrom(ADDRESS).Where(ADDRESS.ADDRESS_ID.EqInt(617)))
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 1 {
		t.Fatalf(testutil.Callers()+"expected 1 row to be affected but got %d", rowsAffected)
	}

	// make sure address was deleted
	exists, err := FetchExists(Log(tx), MySQL.From(ADDRESS).Where(ADDRESS.ADDRESS_ID.EqInt(617)))
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if exists {
		t.Fatal(testutil.Callers(), "address_id 617 was not successfully deleted")
	}

	// delete all addresses with country 'Singapore'
	rowsAffected, _, err = Exec(Log(tx), MySQL.
		DeleteFrom(ADDRESS).
		Using(ADDRESS).
		Join(CITY, CITY.CITY_ID.Eq(ADDRESS.CITY_ID)).
		Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
		Where(COUNTRY.COUNTRY.EqString("Singapore")),
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 9 {
		t.Fatalf(testutil.Callers()+" expected 9 rows to be affected, got %d", rowsAffected)
	}

	// make sure addresses were deleted
	exists, err = FetchExists(Log(tx), MySQL.
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

	// delete with ORDER BY and LIMIT
	rowsAffected, _, err = Exec(Log(tx), MySQL.
		DeleteFrom(ADDRESS).
		Where(Exists(MySQL.
			SelectOne().
			From(CITY).
			Where(
				CITY.CITY_ID.Eq(ADDRESS.CITY_ID),
				CITY.CITY.EqString("Oslo"),
			),
		)).
		OrderBy(ADDRESS.ADDRESS_ID).
		Limit(1),
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 1 {
		t.Fatalf(testutil.Callers()+" expected 1 row to be affected, got %d", rowsAffected)
	}

	// make sure the other two Oslo addresses were unaffected by the delete
	var addressNames []string
	Fetch(Log(tx), MySQL.
		From(ADDRESS).
		Join(CITY, CITY.CITY_ID.Eq(ADDRESS.CITY_ID)).
		Where(CITY.CITY.EqString("Oslo")).
		OrderBy(ADDRESS.ADDRESS_ID),
		func(row *Row) {
			addressName := row.String(ADDRESS.ADDRESS)
			row.Process(func() { addressNames = append(addressNames, addressName) })
		},
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if diff := testutil.Diff(addressNames, []string{"187 Shadowmar Drive", "5034 Camden Street"}); diff != "" {
		t.Fatal(testutil.Callers(), diff)
	}
}
