package sq

import (
	"errors"
	"testing"
)

func Test_SelectQuery(t *testing.T) {
	t.Run("DistinctOnFields dialect != postgres", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("a")
		var q SelectQuery
		q.Dialect = DialectMySQL
		q.DistinctOnFields = Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME}
		q.SelectFields = AliasFields{ACTOR.ACTOR_ID}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("Distinct + DistinctOnFields", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("a")
		var q SelectQuery
		q.Dialect = DialectPostgres
		q.Distinct = true
		q.DistinctOnFields = Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME}
		q.SelectFields = AliasFields{ACTOR.ACTOR_ID}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("DistinctOnFields faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("a")
		var q SelectQuery
		q.Dialect = DialectPostgres
		q.DistinctOnFields = Fields{FaultySQL{}}
		q.SelectFields = AliasFields{ACTOR.ACTOR_ID}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("SelectFields empty", func(t *testing.T) {
		t.Parallel()
		var q SelectQuery
		q.SelectFields = AliasFields{}
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("SelectFields faultySQL", func(t *testing.T) {
		t.Parallel()
		var q SelectQuery
		q.SelectFields = AliasFields{FaultySQL{}}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("FromTable faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("a")
		var q SelectQuery
		q.SelectFields = AliasFields{ACTOR.ACTOR_ID}
		q.FromTable = FaultySQL{}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("JoinTables without FromTable", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("a")
		var q SelectQuery
		q.SelectFields = AliasFields{ACTOR.ACTOR_ID}
		q.JoinTables = append(q.JoinTables, Join(ACTOR, Eq(1, 1)))
		_, _, _, err := ToSQL("", q)
		if err == nil {
			t.Error(testcallers(), "expected error but got nil")
		}
	})

	t.Run("JoinTables faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("a")
		var q SelectQuery
		q.SelectFields = AliasFields{ACTOR.ACTOR_ID}
		q.FromTable = ACTOR
		q.JoinTables = append(q.JoinTables, Join(FaultySQL{}, Eq(1, 1)))
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("WherePredicate faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("a")
		var q SelectQuery
		q.SelectFields = AliasFields{ACTOR.ACTOR_ID}
		q.FromTable = ACTOR
		q.WherePredicate = And(FaultySQL{})
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("GroupByFields faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("a")
		var q SelectQuery
		q.SelectFields = AliasFields{ACTOR.ACTOR_ID}
		q.FromTable = ACTOR
		q.GroupByFields = Fields{FaultySQL{}}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("HavingPredicate faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("a")
		var q SelectQuery
		q.SelectFields = AliasFields{ACTOR.ACTOR_ID}
		q.FromTable = ACTOR
		q.HavingPredicate = And(FaultySQL{})
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("OrderByFields faulty sql", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("a")
		var q SelectQuery
		q.SelectFields = AliasFields{ACTOR.ACTOR_ID}
		q.FromTable = ACTOR
		q.OrderByFields = Fields{FaultySQL{}}
		_, _, _, err := ToSQL("", q)
		if !errors.Is(err, ErrFaultySQL) {
			t.Errorf(testcallers()+" expected ErrFaultySQL but got %#v", err)
		}
	})

	t.Run("FetchableFields", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("a")
		var q SelectQuery
		query, err := q.SetFetchableFields(Fields{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME})
		if err != nil {
			t.Fatalf(testcallers()+" expected nil error, got %#v", err)
		}
		q = query.(SelectQuery)
		fields, err := q.GetFetchableFields()
		if err != nil {
			t.Fatalf(testcallers()+" expected nil error, got %#v", err)
		}
		diff := testdiff(fields, []Field{ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME})
		if diff != "" {
			t.Error(testcallers(), diff)
		}
	})
}
