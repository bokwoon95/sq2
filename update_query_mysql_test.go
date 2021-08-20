package sq

import (
	"strings"
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
)

func Test_MySQLUpdateQuery(t *testing.T) {
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
		tt.item = MySQL.
			Update(ACTOR).
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Set(ACTOR.ACTOR_ID.SetInt64(1)).
			Join(ACTOR, Eq(1, 1)).
			LeftJoin(ACTOR, Eq(1, 1)).
			RightJoin(ACTOR, Eq(1, 1)).
			FullJoin(ACTOR, Eq(1, 1)).
			CrossJoin(ACTOR).
			CustomJoin("CROSS JOIN LATERAL", ACTOR).
			Where(ACTOR.ACTOR_ID.EqInt64(1))
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" UPDATE actor AS a" +
			" JOIN actor AS a ON ? = ?" +
			" LEFT JOIN actor AS a ON ? = ?" +
			" RIGHT JOIN actor AS a ON ? = ?" +
			" FULL JOIN actor AS a ON ? = ?" +
			" CROSS JOIN actor AS a" +
			" CROSS JOIN LATERAL actor AS a" +
			" SET a.actor_id = ?" +
			" WHERE a.actor_id = ?"
		tt.wantArgs = []interface{}{1, 1, 1, 1, 1, 1, 1, 1, int64(1), int64(1)}
		assert(t, tt)
	})

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = MySQL.
			UpdateWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Update(ACTOR).
			Setx(func(c *Column) error {
				c.SetInt64(ACTOR.ACTOR_ID, 1)
				return nil
			}).
			Where(ACTOR.ACTOR_ID.EqInt64(1)).
			OrderBy(ACTOR.ACTOR_ID).
			Limit(10)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" UPDATE actor AS a" +
			" SET a.actor_id = ?" +
			" WHERE a.actor_id = ?" +
			" ORDER BY a.actor_id" +
			" LIMIT ?"
		tt.wantArgs = []interface{}{int64(1), int64(1), int64(10)}
		assert(t, tt)
	})

	t.Run("Multi-table UPDATE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ADDRESS, CITY, COUNTRY := xNEW_ADDRESS("a"), xNEW_CITY("ci"), xNEW_COUNTRY("co")
		tt.item = MySQL.
			Update(ADDRESS).
			Join(CITY, CITY.CITY_ID.Eq(ADDRESS.CITY_ID)).
			Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
			Setx(func(c *Column) error {
				c.SetString(ADDRESS.ADDRESS, "3 CC Street (modified)")
				c.SetString(CITY.CITY, "City C-C (modified)")
				c.SetString(COUNTRY.COUNTRY, "Country C (modified)")
				return nil
			}).
			Where(ADDRESS.ADDRESS_ID.EqInt(632))
		tt.wantQuery = "UPDATE address AS a" +
			" JOIN city AS ci ON ci.city_id = a.city_id" +
			" JOIN country AS co ON co.country_id = ci.country_id" +
			" SET a.address = ?, ci.city = ?, co.country = ?" +
			" WHERE a.address_id = ?"
		tt.wantArgs = []interface{}{"3 CC Street (modified)", "City C-C (modified)", "Country C (modified)", 632}
		assert(t, tt)
	})
}

func TestMySQLSakilaUpdate(t *testing.T) {
	if testing.Short() {
		return
	}
	tx, err := mysqlDB.Begin()
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	defer tx.Rollback()

	// update description for film with film_id 1
	FILM := xNEW_FILM("")
	rowsAffected, _, err := Exec(Log(tx), MySQL.
		Update(FILM).
		Set(FILM.DESCRIPTION.SetString("this is a film with film_id 1")).
		Where(FILM.FILM_ID.EqInt(1)),
		ErowsAffected,
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 1 {
		t.Fatalf(testutil.Callers()+"expected 1 row to be affected but got %d", rowsAffected)
	}

	// make sure description is updated
	var description string
	_, err = Fetch(Log(tx), MySQL.From(FILM).Where(FILM.FILM_ID.EqInt(1)), func(row *Row) {
		description = row.String(FILM.DESCRIPTION)
		row.Close()
	})
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if diff := testutil.Diff("this is a film with film_id 1", description); diff != "" {
		t.Fatal(testutil.Callers(), diff)
	}

	// update all films starring 'THORA TEMPLE' with the suffix ' starring THORA TEMPLE'
	FILM_ACTOR := xNEW_FILM_ACTOR("")
	ACTOR := xNEW_ACTOR("")
	rowsAffected, _, err = Exec(Log(tx), MySQL.
		Update(FILM).
		Set(FILM.DESCRIPTION.Set(Fieldf("CONCAT({}, {})", FILM.DESCRIPTION, " starring THORA TEMPLE"))).
		Join(FILM_ACTOR, FILM_ACTOR.FILM_ID.Eq(FILM.FILM_ID)).
		Join(ACTOR, ACTOR.ACTOR_ID.Eq(FILM_ACTOR.ACTOR_ID)).
		Where(RowValue{ACTOR.FIRST_NAME, ACTOR.LAST_NAME}.Eq(RowValue{"THORA", "TEMPLE"})),
		ErowsAffected,
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 21 {
		t.Fatalf(testutil.Callers()+"expected 21 rows affected, got %d", rowsAffected)
	}

	// TODO: update with LIMIT and ORDER BY. then make sure only a subset of rows were actually modified

	// make sure the film descriptions are updated
	var descriptions []string
	_, err = Fetch(Log(tx), MySQL.From(FILM).Where(FILM.FILM_ID.In(thoraTempleFilmIDs())), func(row *Row) {
		description := row.String(FILM.DESCRIPTION)
		row.Process(func() { descriptions = append(descriptions, description) })
	})
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	for _, description := range descriptions {
		if !strings.HasSuffix(description, " starring THORA TEMPLE") {
			t.Errorf(testutil.Callers()+"description '%s' does not have the correct suffix", description)
		}
	}

	// multi table update (without alias)
	ADDRESS := xNEW_ADDRESS("")
	CITY := xNEW_CITY("")
	COUNTRY := xNEW_COUNTRY("")
	rowsAffected, _, err = Exec(Log(tx), MySQL.
		Update(ADDRESS).
		Join(CITY, CITY.CITY_ID.Eq(ADDRESS.CITY_ID)).
		Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
		Set(
			ADDRESS.ADDRESS.Set(Fieldf("CONCAT({}, {})", ADDRESS.ADDRESS, " (modified)")),
			CITY.CITY.Set(Fieldf("CONCAT({}, {})", CITY.CITY, " (modified)")),
			COUNTRY.COUNTRY.Set(Fieldf("CONCAT({}, {})", COUNTRY.COUNTRY, " (modified)")),
		).
		Where(ADDRESS.ADDRESS_ID.EqInt(632)),
		ErowsAffected,
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 3 {
		t.Fatalf(testutil.Callers()+"expected 3 rows affected, got %d", rowsAffected)
	}

	// make sure the address, city, country names are updated
	var names []string
	_, err = Fetch(Log(tx), MySQL.
		From(ADDRESS).
		Join(CITY, CITY.CITY_ID.Eq(ADDRESS.CITY_ID)).
		Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
		Where(ADDRESS.ADDRESS_ID.EqInt(632)),
		func(row *Row) {
			address := row.String(ADDRESS.ADDRESS)
			city := row.String(CITY.CITY)
			country := row.String(COUNTRY.COUNTRY)
			row.Process(func() { names = append(names, address, city, country) })
		},
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	for _, name := range names {
		if !strings.HasSuffix(name, " (modified)") {
			t.Errorf(testutil.Callers()+"name '%s' does not have the correct suffix", name)
		}
	}

	// multi table update (with alias)
	a := xNEW_ADDRESS("a")
	ci := xNEW_CITY("ci")
	co := xNEW_COUNTRY("co")
	rowsAffected, _, err = Exec(Log(tx), MySQL.
		Update(a).
		Join(ci, ci.CITY_ID.Eq(a.CITY_ID)).
		Join(co, co.COUNTRY_ID.Eq(ci.COUNTRY_ID)).
		Set(
			a.ADDRESS.Set(Fieldf("TRIM(TRAILING {2} FROM {1})", a.ADDRESS, " (modified)")),
			ci.CITY.Set(Fieldf("TRIM(TRAILING {2} FROM {1})", ci.CITY, " (modified)")),
			co.COUNTRY.Set(Fieldf("TRIM(TRAILING {2} FROM {1})", co.COUNTRY, " (modified)")),
		).
		Where(a.ADDRESS_ID.EqInt(632)),
		ErowsAffected,
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 3 {
		t.Fatalf(testutil.Callers()+"expected 3 rows affected, got %d", rowsAffected)
	}

	// make sure the address, city, country names are updated
	names = names[:0]
	_, err = Fetch(Log(tx), MySQL.
		From(ADDRESS).
		Join(CITY, CITY.CITY_ID.Eq(ADDRESS.CITY_ID)).
		Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
		Where(ADDRESS.ADDRESS_ID.EqInt(632)),
		func(row *Row) {
			address := row.String(ADDRESS.ADDRESS)
			city := row.String(CITY.CITY)
			country := row.String(COUNTRY.COUNTRY)
			row.Process(func() { names = append(names, address, city, country) })
		},
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	for _, name := range names {
		if strings.HasSuffix(name, " (modified)") {
			t.Errorf(testutil.Callers()+"name '%s' did not have its suffix trimmed", name)
		}
	}
}
