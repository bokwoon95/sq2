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
		tt.item.AppendSQL(tt.dialect, buf, &gotArgs, gotParams)
		is.Equal(tt.wantQuery, buf.String())
		is.Equal(tt.wantArgs, gotArgs)
	}

	t.Run("https://www.postgresqltutorial.com/postgresql-cte/", func(t *testing.T) {
		var tt TT
		is := testutil.New(t, testutil.Parallel)
		RENTAL, STAFF := NEW_RENTAL(""), NEW_STAFF("s")
		cte_rental, err := NewCTE("cte_rental", nil, Postgres.
			Select(
				RENTAL.STAFF_ID,
				Fieldf("COUNT({})", RENTAL.RENTAL_ID).As("rental_count"),
			).
			From(RENTAL).
			GroupBy(RENTAL.STAFF_ID),
		)
		is.NoErr(err)
		tt.item = Postgres.
			SelectWith(cte_rental).
			Select(
				STAFF.STAFF_ID,
				STAFF.FIRST_NAME,
				STAFF.LAST_NAME,
				cte_rental["rental_count"],
			).
			From(STAFF).
			Join(cte_rental, Eq(cte_rental["staff_id"], STAFF.STAFF_ID))
		tt.wantQuery = "WITH cte_rental AS (" +
			"SELECT rental.staff_id, COUNT(rental.rental_id) AS rental_count" +
			" FROM rental" +
			" GROUP BY rental.staff_id" +
			")" +
			" SELECT s.staff_id, s.first_name, s.last_name, cte_rental.rental_count" +
			" FROM staff AS s" +
			" JOIN cte_rental ON cte_rental.staff_id = s.staff_id"
		tt.wantArgs = []interface{}{}
		assert(is, tt)
	})
}
