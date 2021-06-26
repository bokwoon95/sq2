package sq

import (
	"fmt"
	"testing"
)

func TestCTE(t *testing.T) {
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
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(tt.wantQuery, gotQuery); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testcallers(), diff)
		}
		if tt.wantParams != nil {
			if diff := testdiff(gotParams, tt.wantParams); diff != "" {
				t.Error(testcallers(), diff)
			}
		}
	}

	t.Run("basic CTE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		RENTAL, STAFF := NEW_RENTAL(""), NEW_STAFF("s")
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
			Join(cte, cte.Field("staff_id").Eq(STAFF.STAFF_ID))
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
		tt.dialect = DialectSQLite
		tt.item = SQLite.
			SelectWith(
				NewCTE("cte_1", nil, SQLite.Select(FieldValue(1).As("some_number"))),
				NewRecursiveCTE("tens", []string{"n"}, UnionAll(
					Queryf("SELECT {ten}", Param("ten", 10)),
					Queryf("SELECT tens.n FROM tens WHERE tens.n + {ten} <= {hundred}", Param("ten", 10), Param("hundred", 100)),
				)),
			).
			Select(Fieldf("n")).From(Tablef("tens"))
		tt.wantQuery = "WITH RECURSIVE" +
			" cte_1 AS (SELECT $1 AS some_number)" +
			", tens (n) AS (" +
			"SELECT $2" +
			" UNION ALL" +
			" SELECT tens.n FROM tens WHERE tens.n + $2 <= $3" +
			")" +
			" SELECT n FROM tens"
		tt.wantArgs = []interface{}{1, 10, 100}
		tt.wantParams = map[string][]int{"ten": {1}, "hundred": {2}}
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
		tt.item = SQLite.SelectWith(NewCTE("cte", nil, nil)).
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
		tt.item = SQLite.SelectWith(NewCTE("cte", nil, Queryf("SELECT 1"))).
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
		tt.item = SQLite.SelectWith(NewCTE("cte", nil, SQLite.Select())).
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
		tt.item = SQLite.SelectWith(NewCTE("cte", nil, SQLite.Select(Fieldf("bruh")))).
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

	t.Run("CTEs, variadic query returns error", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = CTEs{
			NewCTE("cte", nil, SQLite.Select(FieldLiteral("1"))),
			CTE{cteName: "faulty_cte", query: Union(FaultySQL{})},
			NewCTE("cte_2", nil, SQLite.Select(FieldLiteral("1"))),
		}
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("CTEs, query returns error", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = CTEs{
			NewCTE("cte", nil, SQLite.Select(FieldLiteral("1"))),
			CTE{cteName: "faulty_cte", query: FaultySQL{}},
			NewCTE("cte_2", nil, SQLite.Select(FieldLiteral("1"))),
		}
		_, _, _, err := ToSQL("", tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})
}

func Test_CTEField(t *testing.T) {
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
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(tt.wantQuery, gotQuery); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testcallers(), diff)
		}
		if tt.wantParams != nil {
			if diff := testdiff(gotParams, tt.wantParams); diff != "" {
				t.Error(testcallers(), diff)
			}
		}
	}

	t.Run("propagate CTE stickyErr", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", nil, nil)
		tt.item = cte.Field("field")
		_, _, _, err := ToSQLExclude(tt.dialect, tt.item, tt.excludedTableQualifiers)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("CTE field not exists", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		cte2 := cte.As("cte2")
		tt.item = cte2.Field("nonexistent_field")
		_, _, _, err := ToSQLExclude(tt.dialect, tt.item, tt.excludedTableQualifiers)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
		fmt.Println(testcallers(), err.Error())
	})

	t.Run("CTEField alias", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		tt.item = cte.Field("field").As("f")
		tt.wantQuery = "cte.field"
		assert(t, tt)
	})

	t.Run("CTEField ASC NULLS LAST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		tt.item = cte.Field("field").Asc().NullsLast()
		tt.wantQuery = "cte.field ASC NULLS LAST"
		assert(t, tt)
	})

	t.Run("CTEField DESC NULLS FIRST", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		tt.item = cte.Field("field").Desc().NullsFirst()
		tt.wantQuery = "cte.field DESC NULLS FIRST"
		assert(t, tt)
	})

	t.Run("CTEField IS NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		tt.item = cte.Field("field").IsNull()
		tt.wantQuery = "cte.field IS NULL"
		assert(t, tt)
	})

	t.Run("CTEField IS NOT NULL", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		tt.item = cte.Field("field").IsNotNull()
		tt.wantQuery = "cte.field IS NOT NULL"
		assert(t, tt)
	})

	t.Run("CTEField IN (slice)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		tt.item = cte.Field("field").In([]int{5, 6, 7})
		tt.wantQuery = "cte.field IN (?, ?, ?)"
		tt.wantArgs = []interface{}{5, 6, 7}
		assert(t, tt)
	})

	t.Run("CTEField Eq", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		tt.item = cte.Field("field").Eq(123)
		tt.wantQuery = "cte.field = ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("CTEField Ne", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		tt.item = cte.Field("field").Ne(123)
		tt.wantQuery = "cte.field <> ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("CTEField Gt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		tt.item = cte.Field("field").Gt(123)
		tt.wantQuery = "cte.field > ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("CTEField Ge", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		tt.item = cte.Field("field").Ge(123)
		tt.wantQuery = "cte.field >= ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("CTEField Lt", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		tt.item = cte.Field("field").Lt(123)
		tt.wantQuery = "cte.field < ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})

	t.Run("CTEField Le", func(t *testing.T) {
		t.Parallel()
		var tt TT
		cte := NewCTE("cte", []string{"field"}, Queryf("SELECT 1"))
		tt.item = cte.Field("field").Le(123)
		tt.wantQuery = "cte.field <= ?"
		tt.wantArgs = []interface{}{123}
		assert(t, tt)
	})
}
