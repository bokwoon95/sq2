package sq

import (
	"fmt"
	"testing"
)

func TestCTE(t *testing.T) {
	type TT struct {
		item       SQLAppender
		wantQuery  string
		wantArgs   []interface{}
		wantParams map[string][]int
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, gotParams, err := ToSQL("", tt.item)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(tt.wantQuery, gotQuery); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.wantArgs, gotArgs); diff != "" {
			t.Error(testcallers(), diff)
		}
		if tt.wantParams != nil {
			if diff := testdiff(tt.wantParams, gotParams); diff != "" {
				t.Error(testcallers(), diff)
			}
		}
	}

	t.Run("basic CTE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		RENTAL, STAFF := NEW_RENTAL(""), NEW_STAFF("s")
		// https://www.postgresqltutorial.com/postgresql-cte/
		cte_rental := NewCTE("cte_rental", nil, Postgres.
			Select(
				RENTAL.STAFF_ID,
				Fieldf("COUNT({})", RENTAL.RENTAL_ID).As("rental_count"),
			).
			From(RENTAL).
			GroupBy(RENTAL.STAFF_ID),
		)
		cte := cte_rental.As("cte")
		tt.item = Postgres.
			SelectWith(cte).
			Select(
				STAFF.STAFF_ID,
				STAFF.FIRST_NAME,
				STAFF.LAST_NAME,
				cte.Field("rental_count"),
			).
			From(STAFF).
			Join(cte, Eq(cte.Field("staff_id"), STAFF.STAFF_ID))
		tt.wantQuery = "WITH cte_rental AS (" +
			"SELECT rental.staff_id, COUNT(rental.rental_id) AS rental_count" +
			" FROM rental" +
			" GROUP BY rental.staff_id" +
			")" +
			" SELECT s.staff_id, s.first_name, s.last_name, cte.rental_count" +
			" FROM staff AS s" +
			" JOIN cte_rental AS cte ON cte.staff_id = s.staff_id"
		assert(t, tt)
	})

	t.Run("recursive CTE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite.
			SelectWith(NewRecursiveCTE("tens", []string{"n"}, UnionAll(
				Queryf("SELECT {ten}", Param("ten", 10)),
				Queryf("SELECT tens.n FROM tens WHERE tens.n + {ten} <= {hundred}", Param("ten", 10), Param("hundred", 100)),
			))).
			Select(Fieldf("n")).From(Tablef("tens"))
		tt.wantQuery = "WITH RECURSIVE tens (n) AS (" +
			"SELECT $1" +
			" UNION ALL" +
			" SELECT tens.n FROM tens WHERE tens.n + $1 <= $2" +
			")" +
			" SELECT n FROM tens"
		tt.wantArgs = []interface{}{10, 100}
		tt.wantParams = map[string][]int{"ten": {0}, "hundred": {1}}
		assert(t, tt)
	})

	t.Run("CTE no name", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite.SelectWith(NewCTE("", nil, nil)).Select(FieldLiteral("1"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("CTE nil query", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite.SelectWith(NewCTE("my_cte", nil, nil)).
			Select(FieldLiteral("1"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("CTE query GetFetchableFields error", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite.SelectWith(NewCTE("my_cte", nil, Queryf("SELECT 1"))).
			Select(FieldLiteral("1"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("CTE query no fields", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite.SelectWith(NewCTE("my_cte", nil, SQLite.Select())).
			Select(FieldLiteral("1"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("CTE query field no name", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = SQLite.SelectWith(NewCTE("my_cte", nil, SQLite.Select(Fieldf("bruh")))).
			Select(FieldLiteral("1"))
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("empty CTE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = CTE{}
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("aliased CTE with stickyErr", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = NewCTE("", nil, nil).As("aliased_cte")
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("aliased CTE with no name", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = CTE{}.As("aliased_cte")
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("aliased CTE with no query", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = CTE{cteName: "cte"}.As("aliased_cte")
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("aliased CTE with no fields", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = CTE{cteName: "cte", query: SQLite.Select()}.As("aliased_cte")
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("CTEs, some with no name", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = CTEs{
			NewCTE("cte", nil, SQLite.Select(FieldLiteral("1"))),
			CTE{},
			NewCTE("cte_2", nil, SQLite.Select(FieldLiteral("1"))),
		}
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("CTEs, some with no query", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = CTEs{
			NewCTE("cte", nil, SQLite.Select(FieldLiteral("1"))),
			CTE{cteName: "cte"},
			NewCTE("cte_2", nil, SQLite.Select(FieldLiteral("1"))),
		}
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})
}

// TODO: I need a query whose methods AppendSQL will always throw an error
