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
		t.Fatal(testutil.Callers()+"expected 21 rows affected, got %d", rowsAffected)
	}

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
			t.Error(testutil.Callers()+"description '%s' does not have the correct suffix", description)
		}
	}
}
