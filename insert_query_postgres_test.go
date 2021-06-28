package sq_test

import (
	"testing"

	. "github.com/bokwoon95/sq"
)

func Test_PostgresInsertQuery(t *testing.T) {
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

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = Postgres.
			InsertInto(ACTOR).
			InsertInto(ACTOR).
			Columns(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
			Values("bob", "the builder").
			Values("alice", "in wonderland").
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1")))
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" INSERT INTO actor AS a (first_name, last_name)" +
			" VALUES ($1, $2), ($3, $4)"
		tt.wantArgs = []interface{}{"bob", "the builder", "alice", "in wonderland"}
		assert(t, tt)
	})

	t.Run("INSERT with RETURNING", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = Postgres.
			InsertWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			InsertInto(ACTOR).
			Valuesx(func(c *Column) error {
				// bob
				c.SetString(ACTOR.FIRST_NAME, "bob")
				c.SetString(ACTOR.LAST_NAME, "the builder")
				// alice
				c.SetString(ACTOR.FIRST_NAME, "alice")
				c.SetString(ACTOR.LAST_NAME, "in wonderland")
				return nil
			}).
			Returning(ACTOR.ACTOR_ID)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" INSERT INTO actor AS a (first_name, last_name)" +
			" VALUES ($1, $2), ($3, $4)" +
			" RETURNING a.actor_id"
		tt.wantArgs = []interface{}{"bob", "the builder", "alice", "in wonderland"}
		assert(t, tt)
	})

	t.Run("INSERT ignore duplicates", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = Postgres.
			InsertWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			InsertInto(ACTOR).
			Valuesx(func(c *Column) error {
				// bob
				c.SetInt64(ACTOR.ACTOR_ID, 1)
				c.SetString(ACTOR.FIRST_NAME, "bob")
				c.SetString(ACTOR.LAST_NAME, "the builder")
				// alice
				c.SetInt64(ACTOR.ACTOR_ID, 2)
				c.SetString(ACTOR.FIRST_NAME, "alice")
				c.SetString(ACTOR.LAST_NAME, "in wonderland")
				return nil
			}).
			OnConflict(ACTOR.ACTOR_ID).
			Where(ACTOR.ACTOR_ID.IsNotNull(), ACTOR.FIRST_NAME.NeString("")).
			DoNothing()
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" INSERT INTO actor AS a (actor_id, first_name, last_name)" +
			" VALUES ($1, $2, $3), ($4, $5, $6)" +
			" ON CONFLICT (actor_id)" +
			" WHERE actor_id IS NOT NULL AND first_name <> $7" +
			" DO NOTHING"
		tt.wantArgs = []interface{}{int64(1), "bob", "the builder", int64(2), "alice", "in wonderland", ""}
		assert(t, tt)
	})

	t.Run("upsert", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = Postgres.
			InsertWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			InsertInto(ACTOR).
			Valuesx(func(c *Column) error {
				// bob
				c.SetInt64(ACTOR.ACTOR_ID, 1)
				c.SetString(ACTOR.FIRST_NAME, "bob")
				c.SetString(ACTOR.LAST_NAME, "the builder")
				// alice
				c.SetInt64(ACTOR.ACTOR_ID, 2)
				c.SetString(ACTOR.FIRST_NAME, "alice")
				c.SetString(ACTOR.LAST_NAME, "in wonderland")
				return nil
			}).
			OnConflictOnConstraint("actor_actor_id_pkey").
			DoUpdateSet(AssignExcluded(ACTOR.FIRST_NAME), AssignExcluded(ACTOR.LAST_NAME)).
			Where(ACTOR.LAST_UPDATE.IsNotNull(), ACTOR.LAST_NAME.NeString(""))
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" INSERT INTO actor AS a (actor_id, first_name, last_name)" +
			" VALUES ($1, $2, $3), ($4, $5, $6)" +
			" ON CONFLICT ON CONSTRAINT actor_actor_id_pkey" +
			" DO UPDATE SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name" +
			" WHERE a.last_update IS NOT NULL AND a.last_name <> $7"
		tt.wantArgs = []interface{}{int64(1), "bob", "the builder", int64(2), "alice", "in wonderland", ""}
		assert(t, tt)
	})

	t.Run("INSERT from SELECT", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR1, ACTOR2 := NEW_ACTOR("a1"), NEW_ACTOR("a2")
		tt.item = Postgres.
			InsertWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			InsertInto(ACTOR1).
			Columns(ACTOR1.FIRST_NAME, ACTOR1.LAST_NAME).
			Select(Postgres.
				Select(ACTOR2.FIRST_NAME, ACTOR2.LAST_NAME).
				From(ACTOR2).
				Where(ACTOR2.ACTOR_ID.In([]int64{1, 2})),
			)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" INSERT INTO actor AS a1 (first_name, last_name)" +
			" SELECT a2.first_name, a2.last_name" +
			" FROM actor AS a2" +
			" WHERE a2.actor_id IN ($1, $2)"
		tt.wantArgs = []interface{}{int64(1), int64(2)}
		assert(t, tt)
	})
}
