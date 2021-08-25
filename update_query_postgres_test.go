package sq

import (
	"sort"
	"strings"
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
)

func Test_PostgresUpdateQuery(t *testing.T) {
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
			Update(ACTOR).
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Set(ACTOR.ACTOR_ID.SetInt64(1)).
			From(ACTOR).
			Join(ACTOR, Eq(1, 1)).
			LeftJoin(ACTOR, Eq(1, 1)).
			RightJoin(ACTOR, Eq(1, 1)).
			FullJoin(ACTOR, Eq(1, 1)).
			CrossJoin(ACTOR).
			CustomJoin("CROSS JOIN LATERAL", ACTOR).
			Where(ACTOR.ACTOR_ID.EqInt64(1))
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" UPDATE actor AS a" +
			" SET actor_id = $1" +
			" FROM actor AS a" +
			" JOIN actor AS a ON $2 = $3" +
			" LEFT JOIN actor AS a ON $4 = $5" +
			" RIGHT JOIN actor AS a ON $6 = $7" +
			" FULL JOIN actor AS a ON $8 = $9" +
			" CROSS JOIN actor AS a" +
			" CROSS JOIN LATERAL actor AS a" +
			" WHERE a.actor_id = $10"
		tt.wantArgs = []interface{}{int64(1), 1, 1, 1, 1, 1, 1, 1, 1, int64(1)}
		assert(t, tt)
	})

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = Postgres.
			UpdateWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Update(ACTOR).
			Setx(func(c *Column) error {
				c.SetInt64(ACTOR.ACTOR_ID, 1)
				return nil
			}).
			Where(ACTOR.ACTOR_ID.EqInt64(1)).
			Returning(ACTOR.ACTOR_ID)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" UPDATE actor AS a" +
			" SET actor_id = $1" +
			" WHERE a.actor_id = $2" +
			" RETURNING a.actor_id"
		tt.wantArgs = []interface{}{int64(1), int64(1)}
		assert(t, tt)
	})
}

func TestPostgresSakilaUpdate(t *testing.T) {
	if testing.Short() {
		return
	}
	tx, err := postgresDB.Begin()
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	defer tx.Rollback()

	// update description for film with film_id 1
	FILM := xNEW_FILM("")
	rowsAffected, _, err := Exec(Log(tx), Postgres.
		Update(FILM).
		Set(FILM.DESCRIPTION.SetString("this is a film with film_id 1")).
		Where(FILM.FILM_ID.EqInt(1)),
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 1 {
		t.Fatalf(testutil.Callers()+"expected 1 row to be affected but got %d", rowsAffected)
	}

	// make sure description is updated
	var description string
	_, err = Fetch(Log(tx), Postgres.From(FILM).Where(FILM.FILM_ID.EqInt(1)), func(row *Row) {
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
	var filmIDs []int
	_, err = Fetch(Log(tx), Postgres.
		Update(FILM).
		Set(FILM.DESCRIPTION.Set(Fieldf("{} || {}", FILM.DESCRIPTION, " starring THORA TEMPLE"))).
		From(FILM_ACTOR).
		Join(ACTOR, ACTOR.ACTOR_ID.Eq(FILM_ACTOR.ACTOR_ID)).
		Where(
			FILM_ACTOR.FILM_ID.Eq(FILM.FILM_ID),
			RowValue{ACTOR.FIRST_NAME, ACTOR.LAST_NAME}.Eq(RowValue{"THORA", "TEMPLE"}),
		),
		func(row *Row) {
			filmID := row.Int(FILM.FILM_ID)
			row.Process(func() { filmIDs = append(filmIDs, filmID) })
		},
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	sort.Ints(filmIDs)
	if diff := testutil.Diff(filmIDs, thoraTempleFilmIDs()); diff != "" {
		t.Fatal(testutil.Callers(), diff)
	}

	// make sure the film descriptions are updated
	var descriptions []string
	_, err = Fetch(Log(tx), Postgres.From(FILM).Where(FILM.FILM_ID.In(filmIDs)), func(row *Row) {
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
