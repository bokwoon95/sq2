package sq

import (
	"database/sql"
	"errors"
	"testing"
)

func Test_UpdateQuery(t *testing.T) {
	t.Run("ColumnMapper return error", func(t *testing.T) {
		t.Parallel()
		var ErrColumnMapper = errors.New("some error")
		var q UpdateQuery
		q.ColumnMapper = func(c *Column) error { return ErrColumnMapper }
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrColumnMapper) {
			t.Errorf(Callers()+" expected ErrColumnMapper but got %#v", err)
		}
	})

	t.Run("CTE faulty sql", func(t *testing.T) {
		t.Parallel()
		var q UpdateQuery
		q.CTEs = CTEs{NewCTE("cte", []string{"n"}, FaultySQL{})}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("nil table provided to UPDATE", func(t *testing.T) {
		t.Parallel()
		var q UpdateQuery
		q.UpdateTable = nil
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("UpdateTable faulty sql", func(t *testing.T) {
		t.Parallel()
		var q UpdateQuery
		q.UpdateTable = FaultySQL{}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("missing Assignments", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.UpdateTable = ACTOR
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("Assignments faulty sql, dialect != mysql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectPostgres
		q.UpdateTable = ACTOR
		q.Assignments = Assignments{Assign(FaultySQL{}, FaultySQL{})}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("FromTable dialect == mysql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectMySQL
		q.UpdateTable = ACTOR
		q.Assignments = Assignments{Assign(FaultySQL{}, FaultySQL{})}
		q.FromTable = ACTOR
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("FromTable faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectPostgres
		q.UpdateTable = ACTOR
		q.Assignments = Assignments{Assign(ACTOR.ACTOR_ID, 1)}
		q.FromTable = FaultySQL{}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("JoinTables without FromTable, dialect != mysql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectPostgres
		q.UpdateTable = ACTOR
		q.Assignments = Assignments{Assign(ACTOR.ACTOR_ID, 1)}
		q.JoinTables = JoinTables{Join(ACTOR, Eq(1, 1))}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("JoinTables faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectPostgres
		q.UpdateTable = ACTOR
		q.Assignments = Assignments{Assign(ACTOR.ACTOR_ID, 1)}
		q.FromTable = ACTOR
		q.JoinTables = JoinTables{Join(FaultySQL{}, Eq(1, 1))}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("Assignments faulty sql, dialect == mysql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectMySQL
		q.UpdateTable = ACTOR
		q.Assignments = Assignments{Assign(FaultySQL{}, FaultySQL{})}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("WherePredicate faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.UpdateTable = ACTOR
		q.Assignments = Assignments{Assign(ACTOR.ACTOR_ID, 1)}
		q.WherePredicate = And(FaultySQL{})
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("OrderByFields dialect != mysql && dialect != sqlite", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectPostgres
		q.UpdateTable = ACTOR
		q.Assignments = Assignments{Assign(ACTOR.ACTOR_ID, 1)}
		q.OrderByFields = Fields{ACTOR.ACTOR_ID}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("OrderByFields faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectMySQL
		q.UpdateTable = ACTOR
		q.Assignments = Assignments{Assign(ACTOR.ACTOR_ID, 1)}
		q.OrderByFields = Fields{FaultySQL{}}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("RowLimit dialect != mysql && dialect != sqlite", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectPostgres
		q.UpdateTable = ACTOR
		q.Assignments = Assignments{Assign(ACTOR.ACTOR_ID, 1)}
		q.RowLimit = sql.NullInt64{Valid: true, Int64: 10}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("RowOffset dialect != sqlite", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectMySQL
		q.UpdateTable = ACTOR
		q.Assignments = Assignments{Assign(ACTOR.ACTOR_ID, 1)}
		q.RowOffset = sql.NullInt64{Valid: true, Int64: 20}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("ReturningFields dialect != postgres && dialect != sqlite", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectMySQL
		q.UpdateTable = ACTOR
		q.Assignments = Assignments{Assign(ACTOR.ACTOR_ID, 1)}
		q.ReturningFields = AliasFields{ACTOR.ACTOR_ID}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("FetchableFields dialect == postgres", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectPostgres
		query, err := q.SetFetchableFields(Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME})
		if err != nil {
			t.Fatalf(Callers()+" expected nil error, got %#v", err)
		}
		q = query.(UpdateQuery)
		fields, err := q.GetFetchableFields()
		if err != nil {
			t.Fatalf(Callers()+" expected nil error, got %#v", err)
		}
		diff := Diff(fields, []Field{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME})
		if diff != "" {
			t.Error(Callers(), diff)
		}
	})

	t.Run("FetchableFields dialect != postgres", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q UpdateQuery
		q.Dialect = DialectMySQL
		_, err := q.SetFetchableFields(Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME})
		if !errors.Is(err, ErrNonFetchableQuery) {
			t.Error(Callers()+" expected ErrNonFetchableQuery, got %#v", err)
		}
		_, err = q.GetFetchableFields()
		if !errors.Is(err, ErrNonFetchableQuery) {
			t.Error(Callers()+" expected ErrNonFetchableQuery, got %#v", err)
		}
	})
}
