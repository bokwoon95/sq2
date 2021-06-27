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
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("nil table provided to DELETE", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		q.FromTables = append(q.FromTables, nil)
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("FromTable faulty sql", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		q.FromTables = append(q.FromTables, FaultySQL{})
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("UsingTable not supported by sqlite", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		q.QueryDialect = DialectSQLite
		q.UsingTable = NEW_ACTOR("")
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("UsingTable faulty sql", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		q.FromTables = append(q.FromTables, NEW_ACTOR(""))
		q.UsingTable = FaultySQL{}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("JoinTables not supported by sqlite", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		q.QueryDialect = DialectSQLite
		q.JoinTables = append(q.JoinTables, Join(NEW_ACTOR("")))
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("JoinTables without UsingTable", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		q.FromTables = append(q.FromTables, NEW_ACTOR(""))
		q.JoinTables = append(q.JoinTables, Join(NEW_ACTOR("")))
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("JoinTables faulty sql", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		q.FromTables = append(q.FromTables, NEW_ACTOR(""))
		q.UsingTable = NEW_ACTOR("")
		q.JoinTables = append(q.JoinTables, Join(FaultySQL{}, Eq(1, 1)))
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})
}
