package sq

import (
	"errors"
	"testing"
)

func Test_DeleteQuery(t *testing.T) {
	t.Run("CTE faulty sql", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		q.CTEs = CTEs{NewCTE("cte", []string{"n"}, FaultySQL{})}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("no table provided to DELETE", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})
}
