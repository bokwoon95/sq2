package sq

import (
	"errors"
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
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
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(tt.wantQuery, gotQuery); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
		if diff := testutil.Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
		if tt.wantParams != nil {
			if diff := testutil.Diff(gotParams, tt.wantParams); diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		}
	}

	t.Run("basic subquery", func(t *testing.T) {
		t.Parallel()
		var tt TT
		RENTAL, STAFF := xNEW_RENTAL(""), xNEW_STAFF("s")
		subquery_rental := NewSubquery("subquery_rental", Postgres(nil).
			Select(
				RENTAL.STAFF_ID,
				Fieldf("COUNT({})", RENTAL.RENTAL_ID).As("rental_count"),
			).
			From(RENTAL).
			GroupBy(RENTAL.STAFF_ID),
		)
		tt.item = Postgres(nil).
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
		tt.item = SQLite(nil).From(NewSubquery("subquery", nil)).Select(Literal("*"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("subquery query GetFetchableFields error", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite(nil).From(NewSubquery("subquery", Queryf("SELECT 1"))).Select(Literal("*"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("subquery query no fields", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite(nil).From(NewSubquery("subquery", MySQL.Select())).Select(Literal("*"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("subquery query field no name", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite(nil).From(NewSubquery("subquery", MySQL.Select(Value(1)))).Select(Literal("*"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("subquery query faulty sql", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite(nil).
			From(NewSubquery("subquery", MySQL.Select(Value(1).As("field")).Where(FaultySQL{}))).
			Select(Literal("*"))
		_, _, _, err := ToSQL("", tt.item)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testutil.Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("subquery no alias, dialect == postgres || dialect == mysql", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Postgres(nil).From(NewSubquery("", Postgres(nil).Select(Value(1).As("n")))).Select(Literal("*"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testutil.Callers(), "expected error but got nil")
		}
	})
}

func Test_SubqueryField(t *testing.T) {
	type TT struct {
		dialect                 string
		item                    SQLExcludeAppender
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
		wantParams              map[string][]int
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, gotParams, err := ToSQLExclude(tt.dialect, tt.item, tt.excludedTableQualifiers)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(tt.wantQuery, gotQuery); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
		if diff := testutil.Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
		if tt.wantParams != nil {
			if diff := testutil.Diff(gotParams, tt.wantParams); diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		}
	}

	t.Run("subquery no alias, dialect == postgres || dialect == mysql", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = DialectPostgres
		q := NewSubquery("", Postgres(nil).Select(Value(1).As("field")))
		tt.item = q.Field("field")
		_, _, _, err := ToSQLExclude(tt.dialect, tt.item, tt.excludedTableQualifiers)
		if err == nil {
			t.Fatal(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("propagate Subquery stickyErr", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", nil)
		tt.item = q.Field("field")
		_, _, _, err := ToSQLExclude(tt.dialect, tt.item, tt.excludedTableQualifiers)
		if err == nil {
			t.Fatal(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("Subquery field not exists", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("nonexistent_field")
		_, _, _, err := ToSQLExclude(tt.dialect, tt.item, tt.excludedTableQualifiers)
		if err == nil {
			t.Fatal(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("SubqueryField alias", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("field").As("f")
		tt.wantQuery = "subquery.field"
		assert(t, tt)
	})

	t.Run("SubqueryField ASC NULLS LAST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("field").Asc().NullsLast()
		tt.wantQuery = "subquery.field ASC NULLS LAST"
		assert(t, tt)
	})

	t.Run("SubqueryField DESC NULLS FIRST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("field").Desc().NullsFirst()
		tt.wantQuery = "subquery.field DESC NULLS FIRST"
		assert(t, tt)
	})

	t.Run("SubqueryField IS NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("field").IsNull()
		tt.wantQuery = "subquery.field IS NULL"
		assert(t, tt)
	})

	t.Run("SubqueryField IS NOT NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("field").IsNotNull()
		tt.wantQuery = "subquery.field IS NOT NULL"
		assert(t, tt)
	})

	t.Run("SubqueryField IN (slice)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("field").In([]int{5, 6, 7})
		tt.wantQuery = "subquery.field IN (?, ?, ?)"
		tt.wantArgs = []interface{}{5, 6, 7}
		assert(t, tt)
	})

	t.Run("SubqueryField Eq", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("field").Eq(123)
		tt.wantQuery = "subquery.field = ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("SubqueryField Ne", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("field").Ne(123)
		tt.wantQuery = "subquery.field <> ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("SubqueryField Gt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("field").Gt(123)
		tt.wantQuery = "subquery.field > ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("SubqueryField Ge", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("field").Ge(123)
		tt.wantQuery = "subquery.field >= ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("SubqueryField Lt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("field").Lt(123)
		tt.wantQuery = "subquery.field < ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("SubqueryField Le", func(t *testing.T) {
		t.Parallel()
		var tt TT
		q := NewSubquery("subquery", MySQL.Select(Value(1).As("field")))
		tt.item = q.Field("field").Le(123)
		tt.wantQuery = "subquery.field <= ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})
}
