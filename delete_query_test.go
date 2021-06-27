package sq

import (
	"database/sql"
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
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectSQLite
		q.FromTables = append(q.FromTables, ACTOR)
		q.UsingTable = ACTOR
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("UsingTable faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.FromTables = append(q.FromTables, ACTOR)
		q.UsingTable = FaultySQL{}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("JoinTables not supported by sqlite", func(t *testing.T) {
		t.Parallel()
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectSQLite
		q.FromTables = append(q.FromTables, ACTOR)
		q.JoinTables = append(q.JoinTables, Join(ACTOR))
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("JoinTables without UsingTable", func(t *testing.T) {
		t.Parallel()
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.FromTables = append(q.FromTables, ACTOR)
		q.JoinTables = append(q.JoinTables, Join(ACTOR))
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("JoinTables faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.FromTables = append(q.FromTables, ACTOR)
		q.UsingTable = ACTOR
		q.JoinTables = append(q.JoinTables, Join(FaultySQL{}, Eq(1, 1)))
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("WherePredicate faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.FromTables = append(q.FromTables, ACTOR)
		q.WherePredicate = And(FaultySQL{})
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("OrderByFields not mysql", func(t *testing.T) {
		t.Parallel()
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectPostgres
		q.FromTables = append(q.FromTables, ACTOR)
		q.OrderByFields = Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("OrderByFields with multi-table DELETE", func(t *testing.T) {
		t.Parallel()
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectMySQL
		q.FromTables = append(q.FromTables, ACTOR)
		q.UsingTable = ACTOR
		q.JoinTables = append(q.JoinTables, Join(ACTOR, Eq(1, 1)))
		q.OrderByFields = Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("OrderByFields faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectMySQL
		q.FromTables = append(q.FromTables, ACTOR)
		q.OrderByFields = Fields{FaultySQL{}}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("RowLimit not mysql", func(t *testing.T) {
		t.Parallel()
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectPostgres
		q.FromTables = append(q.FromTables, ACTOR)
		q.RowLimit = sql.NullInt64{Valid: true, Int64: 10}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("RowLimit with multi-table DELETE", func(t *testing.T) {
		t.Parallel()
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectMySQL
		q.FromTables = append(q.FromTables, ACTOR)
		q.UsingTable = ACTOR
		q.JoinTables = append(q.JoinTables, Join(ACTOR, Eq(1, 1)))
		q.RowLimit = sql.NullInt64{Valid: true, Int64: 10}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("ReturningFields not postgres or sqlite", func(t *testing.T) {
		t.Parallel()
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectMySQL
		q.FromTables = append(q.FromTables, ACTOR)
		q.ReturningFields = Fields{ACTOR.ACTOR_ID}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("ReturningFields faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := NEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectPostgres
		q.FromTables = append(q.FromTables, ACTOR)
		q.ReturningFields = Fields{FaultySQL{}}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})
}
