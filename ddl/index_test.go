package ddl

import (
	"testing"

	"github.com/bokwoon95/sq"
)

func Test_CreateIndexCommnd(t *testing.T) {
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

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = CreateIndexCommand{
			CreateConcurrently: true,
			CreateIfNotExists:  true,
			Index: Index{
				TableSchema:    "some table schema",
				TableName:      "some table name",
				IndexName:      "some index name",
				IndexType:      "HASH",
				IsUnique:       true,
				Columns:        []string{"column a", ""},
				Exprs:          []string{"", "UPPER(column_b)"},
				IncludeColumns: []string{"column_c", "column d"},
				Predicate:      "column_e IS NOT NULL",
			},
		}
		tt.wantQuery = `CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS` +
			` "some index name" ON "some table schema"."some table name" USING HASH ("column a", UPPER(column_b))` +
			` INCLUDE (column_c, "column d") WHERE column_e IS NOT NULL`
		assert(t, tt)
	})

	t.Run("(dialect == sqlite)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item = CreateIndexCommand{
			Index: Index{
				TableSchema: "my_schema",
				TableName:   "my_table",
				IndexName:   "my_index",
				Columns:     []string{"column a", ""},
				Exprs:       []string{"", "UPPER(column_b)"},
				Predicate:   "column_e IS NOT NULL",
			},
		}
		tt.wantQuery = `CREATE INDEX my_index ON my_schema.my_table ("column a", UPPER(column_b)) WHERE column_e IS NOT NULL`
		assert(t, tt)
	})

	t.Run("(dialect == mysql) FULLTEXT", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = CreateIndexCommand{
			Index: Index{
				TableSchema: "my_table",
				TableName:   "my_table",
				IndexName:   "my_index",
				IndexType:   "FULLTEXT",
				Columns:     []string{"my_column"},
			},
		}
		tt.wantQuery = `FULLTEXT INDEX my_index (my_column)`
		assert(t, tt)
	})

	t.Run("(dialect != postgres) CREATE INDEX CONCURRENTLY", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = CreateIndexCommand{
			CreateConcurrently: true,
			Index: Index{
				TableName: "my_table",
				IndexName: "my_index",
				Columns:   []string{"my_column"},
			},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect != postgres && dialect != mysql) index type", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item = CreateIndexCommand{
			Index: Index{
				TableName: "my_table",
				IndexName: "my_index",
				IndexType: "HASH",
				Columns:   []string{"my_column"},
			},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect != postgres) INCLUDE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = CreateIndexCommand{
			Index: Index{
				TableName:      "my_table",
				IndexName:      "my_index",
				Columns:        []string{"my_column"},
				IncludeColumns: []string{"other_column"},
			},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect != postgres && dialect != sqlite) WHERE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = CreateIndexCommand{
			Index: Index{
				TableName: "my_table",
				IndexName: "my_index",
				Columns:   []string{"my_column"},
				Predicate: "other_column IS NULL",
			},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})
}

func Test_DropIndexCommnd(t *testing.T) {
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

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = DropIndexCommand{
			DropConcurrently: true,
			DropIfExists:     true,
			TableSchema:      "some table schema",
			TableName:        "some table name",
			IndexName:        "some index name",
			DropCascade:      true,
		}
		tt.wantQuery = `DROP INDEX CONCURRENTLY IF EXISTS "some table schema"."some index name" CASCADE`
		assert(t, tt)
	})

	t.Run("(dialect == mysql)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = DropIndexCommand{
			TableSchema: "some table schema",
			TableName:   "some table name",
			IndexName:   "some index name",
		}
		tt.wantQuery = "DROP INDEX `some index name`"
		assert(t, tt)
	})

	t.Run("(dialect != postgres) DROP CONCURRENTLY", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = DropIndexCommand{
			DropConcurrently: true,
			TableSchema:      "some table schema",
			TableName:        "some table name",
			IndexName:        "some index name",
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect != postgres && dialect != sqlite) DROP CONCURRENTLY", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = DropIndexCommand{
			DropIfExists: true,
			TableSchema:  "some table schema",
			TableName:    "some table name",
			IndexName:    "some index name",
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect != postgres) DROP CASCADE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = DropIndexCommand{
			TableSchema: "some table schema",
			TableName:   "some table name",
			IndexName:   "some index name",
			DropCascade: true,
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})
}
