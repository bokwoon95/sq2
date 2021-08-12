package sq

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
)

func Test_DeleteQuery(t *testing.T) {
	t.Run("CTE faulty sql", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		q.CTEs = CTEs{NewCTE("cte", []string{"n"}, FaultySQL{})}
		_, _, _, err := ToSQL("", q, nil)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testutil.Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("no table provided to DELETE", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		_, _, _, err := ToSQL("", q, nil)
		if err == nil {
			t.Error(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("nil table provided to DELETE", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		q.FromTables = append(q.FromTables, nil)
		_, _, _, err := ToSQL("", q, nil)
		if err == nil {
			t.Error(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("FromTable faulty sql", func(t *testing.T) {
		t.Parallel()
		var q DeleteQuery
		q.FromTables = append(q.FromTables, FaultySQL{})
		_, _, _, err := ToSQL("", q, nil)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testutil.Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("UsingTable dialect == sqlite", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectSQLite
		q.FromTables = append(q.FromTables, ACTOR)
		q.UsingTable = ACTOR
		_, _, _, err := ToSQL("", q, nil)
		if err == nil {
			t.Error(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("UsingTable faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.FromTables = append(q.FromTables, ACTOR)
		q.UsingTable = FaultySQL{}
		_, _, _, err := ToSQL("", q, nil)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testutil.Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("JoinTables dialect == sqlite", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectSQLite
		q.FromTables = append(q.FromTables, ACTOR)
		q.JoinTables = append(q.JoinTables, Join(ACTOR))
		_, _, _, err := ToSQL("", q, nil)
		if err == nil {
			t.Error(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("JoinTables without UsingTable", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.FromTables = append(q.FromTables, ACTOR)
		q.JoinTables = append(q.JoinTables, Join(ACTOR))
		_, _, _, err := ToSQL("", q, nil)
		if err == nil {
			t.Error(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("JoinTables faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.FromTables = append(q.FromTables, ACTOR)
		q.UsingTable = ACTOR
		q.JoinTables = append(q.JoinTables, Join(FaultySQL{}, Eq(1, 1)))
		_, _, _, err := ToSQL("", q, nil)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testutil.Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("WherePredicate faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.FromTables = append(q.FromTables, ACTOR)
		q.WherePredicate = And(FaultySQL{})
		_, _, _, err := ToSQL("", q, nil)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testutil.Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("OrderByFields dialect != mysql && dialect != sqlite", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectPostgres
		q.FromTables = append(q.FromTables, ACTOR)
		q.OrderByFields = Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME}
		_, _, _, err := ToSQL("", q, nil)
		if err == nil {
			t.Error(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("OrderByFields with multi-table DELETE", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectMySQL
		q.FromTables = append(q.FromTables, ACTOR)
		q.UsingTable = ACTOR
		q.JoinTables = append(q.JoinTables, Join(ACTOR, Eq(1, 1)))
		q.OrderByFields = Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME}
		_, _, _, err := ToSQL("", q, nil)
		if err == nil {
			t.Error(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("OrderByFields faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectMySQL
		q.FromTables = append(q.FromTables, ACTOR)
		q.OrderByFields = Fields{FaultySQL{}}
		_, _, _, err := ToSQL("", q, nil)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testutil.Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("RowLimit dialect != mysql && dialect != sqlite", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectPostgres
		q.FromTables = append(q.FromTables, ACTOR)
		q.RowLimit = sql.NullInt64{Valid: true, Int64: 10}
		_, _, _, err := ToSQL("", q, nil)
		if err == nil {
			t.Error(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("RowLimit with multi-table DELETE", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectMySQL
		q.FromTables = append(q.FromTables, ACTOR)
		q.UsingTable = ACTOR
		q.JoinTables = append(q.JoinTables, Join(ACTOR, Eq(1, 1)))
		q.RowLimit = sql.NullInt64{Valid: true, Int64: 10}
		_, _, _, err := ToSQL("", q, nil)
		if err == nil {
			t.Error(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("ReturningFields dialect != postgres && dialect != sqlite", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectMySQL
		q.FromTables = append(q.FromTables, ACTOR)
		q.ReturningFields = AliasFields{ACTOR.ACTOR_ID}
		_, _, _, err := ToSQL("", q, nil)
		if err == nil {
			t.Error(testutil.Callers(), "expected error but got nil")
		}
	})

	t.Run("ReturningFields faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectPostgres
		q.FromTables = append(q.FromTables, ACTOR)
		q.ReturningFields = AliasFields{FaultySQL{}}
		_, _, _, err := ToSQL("", q, nil)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testutil.Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("FetchableFields dialect == postgres", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectPostgres
		query, err := q.SetFetchableFields(Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME})
		if err != nil {
			t.Fatalf(testutil.Callers()+" expected nil error, got %#v", err)
		}
		q = query.(DeleteQuery)
		fields, err := q.GetFetchableFields()
		if err != nil {
			t.Fatalf(testutil.Callers()+" expected nil error, got %#v", err)
		}
		diff := testutil.Diff(fields, []Field{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME})
		if diff != "" {
			t.Error(testutil.Callers(), diff)
		}
	})

	t.Run("FetchableFields dialect != postgres", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q DeleteQuery
		q.Dialect = DialectMySQL
		_, err := q.SetFetchableFields(Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME})
		if !errors.Is(err, ErrNonFetchableQuery) {
			t.Error(testutil.Callers()+" expected ErrNonFetchableQuery, got %#v", err)
		}
		_, err = q.GetFetchableFields()
		if !errors.Is(err, ErrNonFetchableQuery) {
			t.Error(testutil.Callers()+" expected ErrNonFetchableQuery, got %#v", err)
		}
	})
}
