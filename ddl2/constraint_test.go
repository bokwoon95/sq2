package ddl2

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

	t.Run("(dialect == postgres) UNIQUE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName: "film_actor_actor_id_film_id_key",
				ConstraintType: UNIQUE,
				Columns:        []string{"actor_id", "film_id"},
			},
		}
		tt.wantQuery = "ADD CONSTRAINT film_actor_actor_id_film_id_key UNIQUE (actor_id, film_id)"
		assert(t, tt)
	})

	t.Run("(dialect == postgres) quoted identifier", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName: "bad constraint name",
				ConstraintType: UNIQUE,
			},
			IndexName: "bad index name",
		}
		tt.wantQuery = `ADD CONSTRAINT "bad constraint name" UNIQUE USING INDEX "bad index name"`
		assert(t, tt)
	})

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
		tt.wantQuery = "ADD CONSTRAINT city_country_id_fkey FOREIGN KEY (country_id) REFERENCES db.country (country_id) MATCH FULL ON UPDATE CASCADE ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED"
		assert(t, tt)
	})

	t.Run("(dialect == postgres) EXCLUDE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName:     "customer_email_excl",
				ConstraintType:     EXCLUDE,
				ExclusionIndex:     "GIST",
				Columns:            []string{"email", ""},
				Exprs:              []string{"", "LOWER(email)"},
				ExclusionOperators: []string{"ILIKE", "="},
				Predicate:          "LENGTH(email) > 3",
				IsDeferrable:       true,
			},
		}
		tt.wantQuery = "ADD CONSTRAINT customer_email_excl EXCLUDE USING GIST (email WITH ILIKE, LOWER(email) WITH =) WHERE (LENGTH(email) > 3) DEFERRABLE INITIALLY IMMEDIATE"
		assert(t, tt)
	})

	t.Run("(dialect == mysql) PRIMARY KEY", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName: "customer_customer_id_pkey",
				ConstraintType: PRIMARY_KEY,
				Columns:        []string{"customer_id"},
			},
		}
		tt.wantQuery = "ADD PRIMARY KEY (customer_id)"
		assert(t, tt)
	})

	t.Run("(dialect == mysql) CHECK ... NOT ENFORCED", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName: "film_release_year_check",
				ConstraintType: CHECK,
				CheckExpr:      "release_year >= 1901 AND release_year <= 2155",
			},
			IsNotValid: true,
		}
		tt.wantQuery = "ADD CONSTRAINT film_release_year_check CHECK (release_year >= 1901 AND release_year <= 2155) NOT ENFORCED"
		assert(t, tt)
	})

	t.Run("(dialect == sqlite)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item = AddConstraintCommand{}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect != postgres) ADD CONSTRAINT ... USING INDEX", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName:      "film_actor_actor_id_film_id_key",
				ConstraintType:      UNIQUE,
				IsDeferrable:        true,
				IsInitiallyDeferred: true,
			},
			IndexName: "film_actor_actor_id_film_id_idx",
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) ADD CHECK CONSTRAINT ... USING INDEX", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName: "film_release_year_check",
				ConstraintType: CHECK,
				CheckExpr:      "release_year >= 1901 AND release_year <= 2155",
			},
			IndexName: "film_actor_actor_id_film_id_idx",
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect != postgres) EXCLUDE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName:     "customer_email_excl",
				ConstraintType:     EXCLUDE,
				ExclusionIndex:     "GIST",
				Columns:            []string{"email", ""},
				Exprs:              []string{"", "LOWER(email)"},
				ExclusionOperators: []string{"ILIKE", "="},
				Predicate:          "LENGTH(email) > 3",
				IsDeferrable:       true,
			},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) EXCLUDE without specifying index", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName:     "customer_email_excl",
				ConstraintType:     EXCLUDE,
				Columns:            []string{"email", ""},
				Exprs:              []string{"", "LOWER(email)"},
				ExclusionOperators: []string{"ILIKE", "="},
				Predicate:          "LENGTH(email) > 3",
				IsDeferrable:       true,
			},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) EXCLUDE with empty column and empty expression", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName:     "customer_email_excl",
				ConstraintType:     EXCLUDE,
				ExclusionIndex:     "GIST",
				Columns:            []string{""},
				Exprs:              []string{""},
				ExclusionOperators: []string{"="},
			},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) EXCLUDE with empty operator", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName:     "customer_email_excl",
				ConstraintType:     EXCLUDE,
				ExclusionIndex:          "GIST",
				Columns:            []string{"email"},
				ExclusionOperators: []string{""},
			},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) PRIMARY KEY ... NOT VALID", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName: "customer_customer_id",
				ConstraintType: PRIMARY_KEY,
				Columns:        []string{"customer_id"},
			},
			IsNotValid: true,
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == mysql) PRIMARY KEY ... NOT ENFORCED", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName: "customer_customer_id",
				ConstraintType: PRIMARY_KEY,
				Columns:        []string{"customer_id"},
			},
			IsNotValid: true,
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) CHECK CONSTRAINT DEFERRABLE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName: "film_release_year_check",
				ConstraintType: CHECK,
				CheckExpr:      "release_year >= 1901 AND release_year <= 2155",
				IsDeferrable:   true,
			},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == mysql) CHECK CONSTRAINT DEFERRABLE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = AddConstraintCommand{
			Constraint: Constraint{
				ConstraintName: "film_release_year_check",
				ConstraintType: CHECK,
				CheckExpr:      "release_year >= 1901 AND release_year <= 2155",
				IsDeferrable:   true,
			},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})
}

func Test_AlterConstraintCommnd(t *testing.T) {
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

	t.Run("(dialect == postgres) no-op", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AlterConstraintCommand{
			ConstraintName: "city_country_id_fkey",
		}
		tt.wantQuery = ""
		assert(t, tt)
	})

	t.Run("(dialect == postgres) quoted identifier", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AlterConstraintCommand{
			ConstraintName:  "bad constraint name",
			AlterDeferrable: true,
		}
		tt.wantQuery = `ALTER CONSTRAINT "bad constraint name" NOT DEFERRABLE`
		assert(t, tt)
	})

	t.Run("(dialect == postgres) ALTER CONSTRAINT DEFERRABLE INITIALLY DEFERRED", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AlterConstraintCommand{
			ConstraintName:      "city_country_id_fkey",
			AlterDeferrable:     true,
			IsDeferrable:        true,
			IsInitiallyDeferred: true,
		}
		tt.wantQuery = "ALTER CONSTRAINT city_country_id_fkey DEFERRABLE INITIALLY DEFERRED"
		assert(t, tt)
	})

	t.Run("(dialect == postgres) ALTER CONSTRAINT DEFERRABLE INITIALLY IMMEDIATE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AlterConstraintCommand{
			ConstraintName:  "city_country_id_fkey",
			AlterDeferrable: true,
			IsDeferrable:    true,
		}
		tt.wantQuery = "ALTER CONSTRAINT city_country_id_fkey DEFERRABLE INITIALLY IMMEDIATE"
		assert(t, tt)
	})

	t.Run("(dialect == postgres) ALTER CONSTRAINT NOT DEFERRABLE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AlterConstraintCommand{
			ConstraintName:  "city_country_id_fkey",
			AlterDeferrable: true,
		}
		tt.wantQuery = "ALTER CONSTRAINT city_country_id_fkey NOT DEFERRABLE"
		assert(t, tt)
	})

	t.Run("(dialect == mysql)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = AlterConstraintCommand{
			ConstraintName:  "city_country_id_fkey",
			AlterDeferrable: true,
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == sqlite)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item = AlterConstraintCommand{
			ConstraintName:  "city_country_id_fkey",
			AlterDeferrable: true,
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})
}

func Test_DropConstraintCommnd(t *testing.T) {
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

	t.Run("(dialect == postgres) quoted identifier", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = DropConstraintCommand{
			ConstraintName: "bad constraint name",
		}
		tt.wantQuery = `DROP CONSTRAINT "bad constraint name"`
		assert(t, tt)
	})

	t.Run("(dialect == postgres) DROP IF EXISTS CASCADE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = DropConstraintCommand{
			DropIfExists:   true,
			ConstraintName: "city_country_id_fkey",
			DropCascade:    true,
		}
		tt.wantQuery = "DROP CONSTRAINT IF EXISTS city_country_id_fkey CASCADE"
		assert(t, tt)
	})

	t.Run("(dialect == mysql) DROP", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = DropConstraintCommand{
			ConstraintName: "city_country_id_fkey",
		}
		tt.wantQuery = "DROP CONSTRAINT city_country_id_fkey"
		assert(t, tt)
	})

	t.Run("(dialect == sqlite)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item = DropConstraintCommand{
			ConstraintName: "city_country_id_fkey",
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == mysql) DROP IF EXISTS", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = DropConstraintCommand{
			DropIfExists:   true,
			ConstraintName: "city_country_id_fkey",
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == mysql) DROP CASCADE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = DropConstraintCommand{
			ConstraintName: "city_country_id_fkey",
			DropCascade:    true,
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})
}

func Test_RenameConstraintCommnd(t *testing.T) {
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

	t.Run("quoted identifier", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = RenameConstraintCommand{
			ConstraintName: "bad constraint name",
			RenameToName:   "4lso bad constraint name",
		}
		tt.wantQuery = `RENAME CONSTRAINT "bad constraint name" TO "4lso bad constraint name"`
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = RenameConstraintCommand{
			ConstraintName: "city_country_id_fkey",
			RenameToName:   "fk_city_country_id",
		}
		tt.wantQuery = "RENAME CONSTRAINT city_country_id_fkey TO fk_city_country_id"
		assert(t, tt)
	})

	t.Run("(dialect == sqlite)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item = RenameConstraintCommand{
			ConstraintName: "city_country_id_fkey",
			RenameToName:   "fk_city_country_id",
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})
}
