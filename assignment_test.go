package sq

import (
	"testing"
)

func Test_Assignment(t *testing.T) {
	USERS := struct {
		tmptable
		USER_ID tmpfield
		NAME    tmpfield
		EMAIL   tmpfield
		AGE     tmpfield
	}{
		tmptable: "users",
		USER_ID:  [2]string{"users", "user_id"},
		NAME:     [2]string{"users", "name"},
		EMAIL:    [2]string{"users", "email"},
		AGE:      [2]string{"users", "age"},
	}

	type TT struct {
		dialect                 string
		item                    Assignment
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := ToSQLExclude(tt.dialect, tt.item, tt.excludedTableQualifiers)
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

	t.Run("field assign field", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Assign(USERS.USER_ID, USERS.NAME)
		tt.excludedTableQualifiers = []string{"users"}
		tt.wantQuery = "user_id = name"
		assert(t, tt)
	})

	t.Run("field assign value", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Assign(USERS.USER_ID, 5)
		tt.excludedTableQualifiers = []string{"users"}
		tt.wantQuery = "user_id = ?"
		tt.wantArgs = []interface{}{5}
		assert(t, tt)
	})

	t.Run("field assign query", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Assign(USERS.USER_ID, SQLite.Select(USERS.USER_ID).From(USERS).Limit(1))
		tt.excludedTableQualifiers = []string{"users"}
		tt.wantQuery = "user_id = (SELECT users.user_id FROM users LIMIT ?)"
		tt.wantArgs = []interface{}{int64(1)}
		assert(t, tt)
	})

	t.Run("assign excluded", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AssignExcluded(USERS.USER_ID)
		tt.wantQuery = "user_id = EXCLUDED.user_id"
		assert(t, tt)
	})

	t.Run("assign values", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AssignValues(USERS.USER_ID)
		tt.wantQuery = "user_id = VALUES(user_id)"
		assert(t, tt)
	})

	t.Run("assign new", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AssignNew(USERS.USER_ID)
		tt.wantQuery = "user_id = NEW.user_id"
		assert(t, tt)
	})

	t.Run("self assign", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AssignSelf(USERS.USER_ID)
		tt.wantQuery = "user_id = user_id"
		assert(t, tt)
	})
}

func Test_Assignments(t *testing.T) {
	USERS := struct {
		tmptable
		USER_ID tmpfield
		NAME    tmpfield
		EMAIL   tmpfield
		AGE     tmpfield
	}{
		tmptable: "users",
		USER_ID:  [2]string{"", "user_id"},
		NAME:     [2]string{"", "name"},
		EMAIL:    [2]string{"", "email"},
		AGE:      [2]string{"", "age"},
	}

	type TT struct {
		dialect                 string
		item                    Assignments
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := ToSQLExclude(tt.dialect, tt.item, tt.excludedTableQualifiers)
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

	t.Run("multiple assignments", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Assignments{
			Assign(USERS.USER_ID, USERS.NAME),
			Assign(USERS.AGE, 123456),
			Assign(USERS.EMAIL, "bob@email.com"),
		}
		tt.wantQuery = "user_id = name, age = ?, email = ?"
		tt.wantArgs = []interface{}{123456, "bob@email.com"}
		assert(t, tt)
	})
}
