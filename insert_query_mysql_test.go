package sq_test

import (
	"testing"

	. "github.com/bokwoon95/sq"
)

func Test_MySQLInsertQuery(t *testing.T) {
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

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = MySQL.
			InsertInto(ACTOR).
			InsertInto(ACTOR).
			Columns(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
			Values("bob", "the builder").
			Values("alice", "in wonderland")
		tt.wantQuery = "INSERT INTO actor AS a (first_name, last_name)" +
			" VALUES (?, ?), (?, ?)"
		tt.wantArgs = []interface{}{"bob", "the builder", "alice", "in wonderland"}
		assert(t, tt)
	})

	// TODO: time to bite the bullet and investigate
	// SQLServer/Oracle/Clickhouse what kind of non-standard crap they
	// can shove between 'INSERT' and 'INTO'. This will be the
	// mechanism to which I implement MySQL's INSERT IGNORE INTO and
	// SQLite's INSERT OR ABORT/FAIL/IGNORE/REPLACE/ROLLBACK
	t.Run("INSERT ignore duplicates", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = MySQL.
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
			OnDuplicateKeyUpdate()
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" INSERT INTO actor AS a (actor_id, first_name, last_name)" +
			" VALUES ($1, $2, $3), ($4, $5, $6)" +
			" ON CONFLICT (actor_id)" +
			" WHERE actor_id IS NOT NULL AND first_name <> $7" +
			" DO NOTHING"
		tt.wantArgs = []interface{}{int64(1), "bob", "the builder", int64(2), "alice", "in wonderland", ""}
		// assert(t, tt)
	})

	t.Run("upsert", func(t *testing.T) {
	})

	t.Run("INSERT from SELECT", func(t *testing.T) {
	})
}
