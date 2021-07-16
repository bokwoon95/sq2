package ddl

import (
	"testing"

	"github.com/bokwoon95/sq"
)

func Test_AddConstraintCommnd(t *testing.T) {
	type TT struct {
		dialect   string
		item      Command
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(gotQuery, tt.wantQuery); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testcallers(), diff)
		}
	}

	t.Run("(dialect == postgres) UNIQUE USING INDEX DEFERRABLE INITIALLY DEFERRED", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName:      "film_actor_actor_id_film_id_key",
				ConstraintType:      UNIQUE,
				IsDeferrable:        true,
				IsInitiallyDeferred: true,
			},
			IndexName: "film_actor_actor_id_film_id_idx",
		}
		tt.dialect = sq.DialectPostgres
		tt.wantQuery = "ADD CONSTRAINT film_actor_actor_id_film_id_key UNIQUE USING INDEX film_actor_actor_id_film_id_idx DEFERRABLE INITIALLY DEFERRED"
		assert(t, tt)
	})

	t.Run("(dialect == postgres) PRIMARY KEY USING INDEX DEFERRABLE INITIALLY IMMEDIATE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName: "customer_customer_id_pkey",
				ConstraintType: PRIMARY_KEY,
				IsDeferrable:   true,
			},
			IndexName: "customer_customer_id_idx",
		}
		tt.dialect = sq.DialectPostgres
		tt.wantQuery = "ADD CONSTRAINT customer_customer_id_pkey PRIMARY KEY USING INDEX customer_customer_id_idx DEFERRABLE INITIALLY IMMEDIATE"
		assert(t, tt)
	})

	t.Run("(dialect == postgres) CHECK ... NOT VALID", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName: "film_release_year_check",
				ConstraintType: CHECK,
				CheckExpr:      "release_year >= 1901 AND release_year <= 2155",
			},
			IsNotValid: true,
		}
		tt.dialect = sq.DialectPostgres
		tt.wantQuery = "ADD CONSTRAINT film_release_year_check CHECK (release_year >= 1901 AND release_year <= 2155) NOT VALID"
		assert(t, tt)
	})

	t.Run("(dialect == postgres) FOREIGN KEY ... DEFERRABLE INITIALLY DEFERRED", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName:      "city_country_id_fkey",
				ConstraintType:      FOREIGN_KEY,
				Columns:             []string{"country_id"},
				ReferencesSchema:    "db",
				ReferencesTable:     "country",
				ReferencesColumns:   []string{"country_id"},
				MatchOption:         "MATCH FULL",
				UpdateRule:          CASCADE,
				DeleteRule:          RESTRICT,
				IsDeferrable:        true,
				IsInitiallyDeferred: true,
			},
		}
		tt.dialect = sq.DialectPostgres
		tt.wantQuery = "ADD CONSTRAINT city_country_id_fkey FOREIGN KEY (country_id) REFERENCES db.country (country_id) MATCH FULL ON UPDATE CASCADE ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED"
		assert(t, tt)
	})

	t.Run("(dialect == postgres) EXCLUDE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName: "customer_email_excl",
				ConstraintType: EXCLUDE,
				IndexType:      "GIST",
				Columns:        []string{"email", ""},
				Exprs:          []string{"", "LOWER(email)"},
				Operators:      []string{"ILIKE", "="},
				Predicate:      "LENGTH(email) > 3",
				IsDeferrable:   true,
			},
		}
		tt.dialect = sq.DialectPostgres
		tt.wantQuery = "ADD CONSTRAINT customer_email_excl EXCLUDE USING GIST (email WITH ILIKE, (LOWER(email)) WITH =) WHERE (LENGTH(email) > 3) DEFERRABLE INITIALLY IMMEDIATE"
		assert(t, tt)
	})
}
