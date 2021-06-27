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
		ACTOR := NEW_ACTOR("")
		tt.item = MySQL.
			InsertInto(ACTOR).
			Columns(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
			Values("bob", "the builder").
			Values("alice", "in wonderland")
		tt.wantQuery = "INSERT INTO actor (first_name, last_name)" +
			" VALUES (?, ?), (?, ?)"
		tt.wantArgs = []interface{}{"bob", "the builder", "alice", "in wonderland"}
		assert(t, tt)
	})

	t.Run("INSERT ignore duplicates (INSERT IGNORE)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("")
		tt.item = MySQL.
			InsertIgnoreInto(ACTOR).
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
			})
		tt.wantQuery = "INSERT IGNORE INTO actor (actor_id, first_name, last_name)" +
			" VALUES (?, ?, ?), (?, ?, ?)"
		tt.wantArgs = []interface{}{int64(1), "bob", "the builder", int64(2), "alice", "in wonderland"}
		assert(t, tt)
	})

	t.Run("INSERT ignore duplicates (ON DUPLICATE KEY UPDATE)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("")
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
			OnDuplicateKeyUpdate(
				AssignSelf(ACTOR.FIRST_NAME),
				AssignSelf(ACTOR.LAST_NAME),
			)
		tt.wantQuery = "INSERT INTO actor (actor_id, first_name, last_name)" +
			" VALUES (?, ?, ?), (?, ?, ?)" +
			" ON DUPLICATE KEY UPDATE first_name = first_name, last_name = last_name"
		tt.wantArgs = []interface{}{int64(1), "bob", "the builder", int64(2), "alice", "in wonderland"}
		assert(t, tt)
	})

	t.Run("upsert", func(t *testing.T) {
	})

	t.Run("INSERT from SELECT", func(t *testing.T) {
	})
}
