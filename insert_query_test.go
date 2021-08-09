package sq

import (
	"errors"
	"testing"
)

func Test_InsertQuery(t *testing.T) {
	t.Run("ColumnMapper return error", func(t *testing.T) {
		t.Parallel()
		var ErrColumnMapper = errors.New("some error")
		var q InsertQuery
		q.ColumnMapper = func(c *Column) error { return ErrColumnMapper }
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrColumnMapper) {
			t.Errorf(Callers()+" expected ErrColumnMapper but got %#v", err)
		}
	})

	t.Run("CTE faulty sql", func(t *testing.T) {
		t.Parallel()
		var q InsertQuery
		q.CTEs = CTEs{NewCTE("cte", []string{"n"}, FaultySQL{})}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("INSERT IGNORE dialect != mysql", func(t *testing.T) {
		t.Parallel()
		var q InsertQuery
		q.Dialect = DialectPostgres
		q.InsertIgnore = true
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("nil table provided to INSERT", func(t *testing.T) {
		t.Parallel()
		var q InsertQuery
		q.IntoTable = nil
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("IntoTable faulty sql", func(t *testing.T) {
		t.Parallel()
		var q InsertQuery
		q.IntoTable = FaultySQL{}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("IntoTable alias, dialect == mysql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("a")
		var q InsertQuery
		q.Dialect = DialectMySQL
		q.IntoTable = ACTOR
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("InsertColumns faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.IntoTable = ACTOR
		q.InsertColumns = Fields{FaultySQL{}}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("RowValues faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.IntoTable = ACTOR
		q.RowValues = RowValues{{FaultySQL{}}}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("RowAlias dialect != mysql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.Dialect = DialectPostgres
		q.IntoTable = ACTOR
		q.InsertColumns = Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME}
		q.RowValues = RowValues{{1, "bob", "the builder"}, {2, "alice", "in wonderland"}}
		q.RowAlias = "NEW"
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("SelectQuery faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.IntoTable = ACTOR
		q.SelectQuery = &SelectQuery{SelectFields: AliasFields{FaultySQL{}}}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("missing RowValues and SelectQuery", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.IntoTable = ACTOR
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("ConflictConstraint dialect != postgres", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.Dialect = DialectSQLite
		q.IntoTable = ACTOR
		q.InsertColumns = Fields{ACTOR.ACTOR_ID}
		q.RowValues = RowValues{{1}, {2}, {3}}
		q.ConflictConstraint = "actor_actor_id_pkey"
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("ConflictFields faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.Dialect = DialectPostgres
		q.IntoTable = ACTOR
		q.InsertColumns = Fields{ACTOR.ACTOR_ID}
		q.RowValues = RowValues{{1}, {2}, {3}}
		q.ConflictFields = Fields{FaultySQL{}}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("ConflictPredicate faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.Dialect = DialectPostgres
		q.IntoTable = ACTOR
		q.InsertColumns = Fields{ACTOR.ACTOR_ID}
		q.RowValues = RowValues{{1}, {2}, {3}}
		q.ConflictFields = Fields{ACTOR.ACTOR_ID}
		q.ConflictPredicate = And(FaultySQL{})
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("Resolution faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.Dialect = DialectPostgres
		q.IntoTable = ACTOR
		q.InsertColumns = Fields{ACTOR.ACTOR_ID}
		q.RowValues = RowValues{{1}, {2}, {3}}
		q.ConflictFields = Fields{ACTOR.ACTOR_ID}
		q.Resolution = Assignments{Assign(FaultySQL{}, FaultySQL{})}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("ResolutionPredicate faulty sql, dialect == mysql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.Dialect = DialectMySQL
		q.IntoTable = ACTOR
		q.InsertColumns = Fields{ACTOR.ACTOR_ID}
		q.RowValues = RowValues{{1}, {2}, {3}}
		q.Resolution = Assignments{Assign(FaultySQL{}, FaultySQL{})}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("ResolutionPredicate faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.Dialect = DialectPostgres
		q.IntoTable = ACTOR
		q.InsertColumns = Fields{ACTOR.ACTOR_ID}
		q.RowValues = RowValues{{1}, {2}, {3}}
		q.ConflictFields = Fields{ACTOR.ACTOR_ID}
		q.Resolution = Assignments{AssignExcluded(ACTOR.ACTOR_ID)}
		q.ResolutionPredicate = And(FaultySQL{})
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("ReturningFields dialect != postgres && dialect != sqlite", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.Dialect = DialectMySQL
		q.IntoTable = ACTOR
		q.InsertColumns = Fields{ACTOR.ACTOR_ID}
		q.RowValues = RowValues{{1}, {2}, {3}}
		q.ReturningFields = AliasFields{ACTOR.ACTOR_ID}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(Callers(), "expected error but got nil")
		}
	})

	t.Run("ReturningFields faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.Dialect = DialectPostgres
		q.IntoTable = ACTOR
		q.InsertColumns = Fields{ACTOR.ACTOR_ID}
		q.RowValues = RowValues{{1}, {2}, {3}}
		q.ReturningFields = AliasFields{FaultySQL{}}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(Callers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("FetchableFields dialect == postgres", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var q InsertQuery
		q.Dialect = DialectPostgres
		query, err := q.SetFetchableFields(Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME})
		if err != nil {
			t.Fatalf(Callers()+" expected nil error, got %#v", err)
		}
		q = query.(InsertQuery)
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
		var q InsertQuery
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
