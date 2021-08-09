package sq

import (
	"testing"
)

func Test_SQLiteUpdateQuery(t *testing.T) {
	type TT struct {
		dialect   string
		item      Query
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := ToSQL(tt.dialect, tt.item)
		if err != nil {
			t.Fatal(Callers(), err)
		}
		if diff := Diff(gotQuery, tt.wantQuery); diff != "" {
			t.Error(Callers(), diff)
		}
		if diff := Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(Callers(), diff)
		}
	}

	t.Run("joins", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("")
		tt.item = SQLite.
			Update(ACTOR).
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Set(ACTOR.ACTOR_ID.SetInt64(1)).
			From(ACTOR).
			Join(ACTOR, Eq(1, 1)).
			LeftJoin(ACTOR, Eq(1, 1)).
			CrossJoin(ACTOR).
			CustomJoin(",", ACTOR).
			Where(ACTOR.ACTOR_ID.EqInt64(1))
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" UPDATE actor" +
			" SET actor_id = $1" +
			" FROM actor" +
			" JOIN actor ON $2 = $3" +
			" LEFT JOIN actor ON $4 = $5" +
			" CROSS JOIN actor" +
			" , actor" +
			" WHERE actor.actor_id = $6"
		tt.wantArgs = []interface{}{int64(1), 1, 1, 1, 1, int64(1)}
		assert(t, tt)
	})

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = SQLite.
			UpdateWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Update(ACTOR).
			Setx(func(c *Column) error {
				c.SetInt64(ACTOR.ACTOR_ID, 1)
				return nil
			}).
			Where(ACTOR.ACTOR_ID.EqInt64(1)).
			OrderBy(ACTOR.ACTOR_ID).
			Limit(10).
			Offset(20)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" UPDATE actor AS a" +
			" SET actor_id = $1" +
			" WHERE a.actor_id = $2" +
			" ORDER BY a.actor_id" +
			" LIMIT $3" +
			" OFFSET $4"
		tt.wantArgs = []interface{}{int64(1), int64(1), int64(10), int64(20)}
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
