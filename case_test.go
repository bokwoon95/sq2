package sq

import (
	"testing"
)

func Test_PredicateCases(t *testing.T) {
	type TT struct {
		dialect                 string
		item                    SQLExcludeAppender
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := ToSQLExclude(tt.dialect, tt.item, tt.excludedTableQualifiers)
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

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		_, _, _, err := ToSQLExclude("", PredicateCases{}, nil)
		if err == nil {
			t.Fatal(Callers(), "expected error but got nil")
		}
	})

	t.Run("1 case", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = CaseWhen(ACTOR.ACTOR_ID.IsNull(), 5)
		tt.wantQuery = "CASE WHEN a.actor_id IS NULL THEN ? END"
		tt.wantArgs = []interface{}{5}
		assert(t, tt)
	})

	t.Run("2 cases", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = CaseWhen(ACTOR.ACTOR_ID.IsNull(), 5).When(ACTOR.FIRST_NAME.EqString("abc"), ACTOR.LAST_NAME).As("alias")
		tt.wantQuery = "CASE WHEN a.actor_id IS NULL THEN ? WHEN a.first_name = ? THEN a.last_name END"
		tt.wantArgs = []interface{}{5, "abc"}
		assert(t, tt)
	})

	t.Run("2 cases, fallback", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = CaseWhen(ACTOR.ACTOR_ID.IsNull(), 5).When(ACTOR.FIRST_NAME.EqString("abc"), ACTOR.LAST_NAME).Else(6789)
		tt.wantQuery = "CASE WHEN a.actor_id IS NULL THEN ? WHEN a.first_name = ? THEN a.last_name ELSE ? END"
		tt.wantArgs = []interface{}{5, "abc", 6789}
		assert(t, tt)
	})
}

func Test_SimpleCases(t *testing.T) {
	type TT struct {
		dialect                 string
		item                    SQLExcludeAppender
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := ToSQLExclude(tt.dialect, tt.item, tt.excludedTableQualifiers)
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

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		_, _, _, err := ToSQLExclude("", SimpleCases{}, nil)
		if err == nil {
			t.Fatal(Callers(), "expected error but got nil")
		}
	})

	t.Run("expression only", func(t *testing.T) {
		t.Parallel()
		a := NEW_ACTOR("a")
		_, _, _, err := ToSQLExclude("", Case(a.ACTOR_ID), nil)
		if err == nil {
			t.Fatal(Callers(), "expected error but got nil")
		}
	})

	t.Run("expression, 1 case", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = Case(ACTOR.ACTOR_ID).When(99, 97)
		tt.wantQuery = "CASE a.actor_id WHEN ? THEN ? END"
		tt.wantArgs = []interface{}{99, 97}
		assert(t, tt)
	})

	t.Run("expression, 2 cases", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = Case(ACTOR.ACTOR_ID).When(99, 97).When(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).As("alias")
		tt.wantQuery = "CASE a.actor_id WHEN ? THEN ? WHEN a.first_name THEN a.last_name END"
		tt.wantArgs = []interface{}{99, 97}
		assert(t, tt)
	})

	t.Run("expression, 2 cases, fallback", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := NEW_ACTOR("a")
		tt.item = Case(ACTOR.ACTOR_ID).When(99, 97).When(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).Else("abcde")
		tt.wantQuery = "CASE a.actor_id WHEN ? THEN ? WHEN a.first_name THEN a.last_name ELSE ? END"
		tt.wantArgs = []interface{}{99, 97, "abcde"}
		assert(t, tt)
	})
}
