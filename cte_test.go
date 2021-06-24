package sq

import (
	"bytes"
	"testing"
)

func TestCTE(t *testing.T) {
	type TT struct {
		dialect   string
		item      SQLAppender
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, gotParams := []interface{}{}, map[string][]int{}
		err := tt.item.AppendSQL(tt.dialect, buf, &gotArgs, gotParams)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(tt.wantQuery, buf.String()); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.wantArgs, gotArgs); diff != "" {
			t.Error(testcallers(), diff)
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
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})
}
