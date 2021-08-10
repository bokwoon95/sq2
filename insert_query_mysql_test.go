package sq

import (
	"testing"
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
		ACTOR := xNEW_ACTOR("")
		tt.item = MySQL.
			InsertInto(ACTOR).
			Columns(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
			Values("bob", "the builder").
			Values("alice", "in wonderland").
			As("NEW", "fname", "lname")
		tt.wantQuery = "INSERT INTO actor (first_name, last_name)" +
			" VALUES (?, ?), (?, ?) AS NEW (fname, lname)"
		tt.wantArgs = []interface{}{"bob", "the builder", "alice", "in wonderland"}
		assert(t, tt)
	})

	t.Run("INSERT ignore duplicates (INSERT IGNORE)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("")
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
		ACTOR := xNEW_ACTOR("")
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
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("")
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
			As("NEW").
			OnDuplicateKeyUpdate(
				AssignAlias(ACTOR.FIRST_NAME, "NEW"),
				AssignAlias(ACTOR.LAST_NAME, "NEW"),
			)
		tt.wantQuery = "INSERT INTO actor (actor_id, first_name, last_name)" +
			" VALUES (?, ?, ?), (?, ?, ?) AS NEW" +
			" ON DUPLICATE KEY UPDATE first_name = NEW.first_name, last_name = NEW.last_name"
		tt.wantArgs = []interface{}{int64(1), "bob", "the builder", int64(2), "alice", "in wonderland"}
		assert(t, tt)
	})

	// TODO: mysql docs say to use a derived table to avoid using VALUES() for
	// ON DUPLICATE KEY UPDATE. Can I just directly alias my SELECT-ed fields
	// instead? Need to test it out. Not important though because I'm not going
	// to be using it here.
	t.Run("INSERT from SELECT", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR1, ACTOR2 := xNEW_ACTOR(""), xNEW_ACTOR("a2")
		tt.item = MySQL.
			InsertInto(ACTOR1).
			Columns(ACTOR1.FIRST_NAME, ACTOR1.LAST_NAME).
			Select(MySQL.
				Select(ACTOR2.FIRST_NAME, ACTOR2.LAST_NAME).
				From(ACTOR2).
				Where(ACTOR2.ACTOR_ID.In([]int64{1, 2})),
			)
		tt.wantQuery = "INSERT INTO actor (first_name, last_name)" +
			" SELECT a2.first_name, a2.last_name" +
			" FROM actor AS a2" +
			" WHERE a2.actor_id IN (?, ?)"
		tt.wantArgs = []interface{}{int64(1), int64(2)}
		assert(t, tt)
	})
}
