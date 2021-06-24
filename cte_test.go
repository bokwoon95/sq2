package sq

import (
	"bytes"
	"testing"

	"github.com/bokwoon95/testutil"
)

func TestCTE(t *testing.T) {
	type TT struct {
		dialect   string
		item      SQLAppender
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(is testutil.I, tt TT) {
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, gotParams := []interface{}{}, map[string][]int{}
		err := tt.item.AppendSQL(tt.dialect, buf, &gotArgs, gotParams)
		is.NoErr(err)
		is.Equal(tt.wantQuery, buf.String())
		is.Equal(tt.wantArgs, gotArgs)
	}

	// TODO: oh god change of heart change of heart, I want to copy jOOQ's syntax for CTEs a la cte.Field("field_name")

	t.Run("https://www.postgresqltutorial.com/postgresql-cte/", func(t *testing.T) {
		var tt TT
		is := testutil.New(t, testutil.Parallel)
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
		assert(is, tt)
	})
}
