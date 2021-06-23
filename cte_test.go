package sq

import (
	"testing"
)

/*
WITH cte_rental AS (
    SELECT staff_id,
        COUNT(rental_id) rental_count
    FROM   rental
    GROUP  BY staff_id
)
SELECT s.staff_id,
    first_name,
    last_name,
    rental_count
FROM staff s
    INNER JOIN cte_rental USING (staff_id);
*/

func TestCTE(t *testing.T) {

	type TT struct {
		dialect   string
		item      SQLAppender
		wantQuery string
		wantArgs  []interface{}
	}

	// assert := func(t *testing.T, tt TT) {
	// 	is := testutil.New(t, testutil.Parallel)
	// 	buf := bufpool.Get().(*bytes.Buffer)
	// 	defer func() {
	// 		buf.Reset()
	// 		bufpool.Put(buf)
	// 	}()
	// 	gotArgs, gotParams := []interface{}{}, map[string][]int{}
	// 	tt.item.AppendSQL(tt.dialect, buf, &gotArgs, gotParams)
	// 	is.Equal(tt.wantQuery, buf.String())
	// 	is.Equal(tt.wantArgs, gotArgs)
	// }

	t.Run("https://www.postgresqltutorial.com/postgresql-cte/", func(t *testing.T) {
	})
}
