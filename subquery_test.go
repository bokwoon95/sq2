package sq

import (
	"testing"
)

func TestSubquery(t *testing.T) {
	type TT struct {
		dialect    string
		item       SQLAppender
		wantQuery  string
		wantArgs   []interface{}
		wantParams map[string][]int
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, gotParams, err := ToSQL("", tt.item)
		if err != nil {
			t.Fatal(Callers(), err)
		}
		if diff := Diff(tt.wantQuery, gotQuery); diff != "" {
			t.Error(Callers(), diff)
		}
		if diff := Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(Callers(), diff)
		}
		if tt.wantParams != nil {
			if diff := Diff(gotParams, tt.wantParams); diff != "" {
				t.Error(Callers(), diff)
			}
		}
	}

	t.Run("basic subquery", func(t *testing.T) {
		t.Parallel()
		var tt TT
		RENTAL, STAFF := NEW_RENTAL(""), NEW_STAFF("s")
		subquery_rental := NewSubquery("subquery_rental", Postgres.
			Select(
				RENTAL.STAFF_ID,
				Fieldf("COUNT({})", RENTAL.RENTAL_ID).As("rental_count"),
			).
			From(RENTAL).
			GroupBy(RENTAL.STAFF_ID),
		)
		tt.item = Postgres.
			Select(
				STAFF.STAFF_ID,
				STAFF.FIRST_NAME,
				STAFF.LAST_NAME,
				subquery_rental.Field("rental_count"),
			).
			From(STAFF).
			Join(subquery_rental, subquery_rental.Field("staff_id").Eq(STAFF.STAFF_ID))
		tt.wantQuery = "SELECT s.staff_id, s.first_name, s.last_name, subquery_rental.rental_count" +
			" FROM staff AS s" +
			" JOIN (" +
			"SELECT rental.staff_id, COUNT(rental.rental_id) AS rental_count" +
			" FROM rental" +
			" GROUP BY rental.staff_id" +
			") AS subquery_rental ON subquery_rental.staff_id = s.staff_id"
		assert(t, tt)
	})

	t.Run("subquery nil query", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite.From(NewSubquery("subquery", nil)).Select(FieldLiteral("*"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(Callers(), "expected error but got nil")
		}
	})

	t.Run("subquery query GetFetchableFields error", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite.From(NewSubquery("subquery", Queryf("SELECT 1"))).Select(FieldLiteral("*"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(Callers(), "expected error but got nil")
		}
	})

	t.Run("subquery query no fields", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite.From(NewSubquery("subquery", SQLite.Select())).Select(FieldLiteral("*"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(Callers(), "expected error but got nil")
		}
	})

	t.Run("subquery no alias, dialect == postgres || dialect == mysql", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Postgres.From(NewSubquery("", Postgres.Select(FieldValue(1).As("n")))).Select(FieldLiteral("*"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(Callers(), "expected error but got nil")
		}
	})
}
