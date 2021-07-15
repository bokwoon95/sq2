package ddl

import (
	"testing"

	"github.com/bokwoon95/sq"
)

func Test_Column(t *testing.T) {
	type TT struct {
		dialect                 string
		item                    sq.SQLExcludeAppender
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := sq.ToSQLExclude(tt.dialect, tt.item, tt.excludedTableQualifiers)
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

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Column{
			TableName:  "actor",
			ColumnName: "actor_id",
		}
		tt.wantQuery = "actor.actor_id"
		assert(t, tt)
	})

	t.Run("with table alias", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Column{
			TableName:  "actor",
			TableAlias: "a",
			ColumnName: "actor_id",
		}
		tt.wantQuery = "a.actor_id"
		assert(t, tt)
	})

	t.Run("excludedTableQualifiers", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Column{
			TableName:  "actor",
			ColumnName: "actor_id",
		}
		tt.excludedTableQualifiers = []string{"actor"}
		tt.wantQuery = "actor_id"
		assert(t, tt)
	})

	t.Run("quoted identifiers", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = Column{
			TableName:  "table with spaces",
			ColumnName: "column with spaces",
		}
		tt.excludedTableQualifiers = []string{"actor"}
		tt.wantQuery = `"table with spaces"."column with spaces"`
		assert(t, tt)
	})
}

func Test_AddColumnCommand(t *testing.T) {
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

	t.Run("postgres identity column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = &AddColumnCommand{
			AddIfNotExists: true,
			Column: Column{
				ColumnName:    "actor_id",
				ColumnType:    "INT",
				IsNotNull:     true,
				Identity:      BY_DEFAULT_AS_IDENTITY,
				CollationName: "C",
			},
		}
		tt.dialect = sq.DialectPostgres
		tt.wantQuery = `ADD COLUMN IF NOT EXISTS actor_id INT NOT NULL GENERATED BY DEFAULT AS IDENTITY COLLATE "C"`
		assert(t, tt)
	})

	t.Run("postgres generated column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = &AddColumnCommand{
			Column: Column{
				ColumnName:          "full_name",
				ColumnType:          "TEXT",
				GeneratedExpr:       "first_name || ' ' || last_name",
				GeneratedExprStored: true,
			},
		}
		tt.dialect = sq.DialectPostgres
		tt.wantQuery = `ADD COLUMN full_name TEXT GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED`
		assert(t, tt)
	})

	t.Run("postgres autoincrement column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = &AddColumnCommand{
			Column: Column{
				ColumnName:    "actor_id",
				ColumnType:    "INT",
				Autoincrement: true,
			},
		}
		tt.dialect = sq.DialectPostgres
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
	})

	t.Run("postgres generated virtual column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = &AddColumnCommand{
			Column: Column{
				ColumnName:    "full_name",
				ColumnType:    "TEXT",
				GeneratedExpr: "first_name || ' ' || last_name",
			},
		}
		tt.dialect = sq.DialectPostgres
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
	})
}

// func Test_AlterColumnCommand(t *testing.T) {
// 	type TT struct {
// 		dialect   string
// 		item      Command
// 		wantQuery string
// 		wantArgs  []interface{}
// 	}
//
// 	assert := func(t *testing.T, tt TT) {
// 		gotQuery, gotArgs, _, err := sq.ToSQL(tt.dialect, tt.item)
// 		if err != nil {
// 			t.Fatal(testcallers(), err)
// 		}
// 		if diff := testdiff(gotQuery, tt.wantQuery); diff != "" {
// 			t.Error(testcallers(), diff)
// 		}
// 		if diff := testdiff(gotArgs, tt.wantArgs); diff != "" {
// 			t.Error(testcallers(), diff)
// 		}
// 	}
//
// 	t.Run("basic", func(t *testing.T) {
// 		t.Parallel()
// 		var tt TT
// 		tt.item = &AlterColumnCommand{}
// 		tt.wantQuery = "actor.actor_id"
// 		assert(t, tt)
// 	})
// }
//
// func Test_DropColumnCommand(t *testing.T) {
// 	type TT struct {
// 		dialect   string
// 		item      Command
// 		wantQuery string
// 		wantArgs  []interface{}
// 	}
//
// 	assert := func(t *testing.T, tt TT) {
// 		gotQuery, gotArgs, _, err := sq.ToSQL(tt.dialect, tt.item)
// 		if err != nil {
// 			t.Fatal(testcallers(), err)
// 		}
// 		if diff := testdiff(gotQuery, tt.wantQuery); diff != "" {
// 			t.Error(testcallers(), diff)
// 		}
// 		if diff := testdiff(gotArgs, tt.wantArgs); diff != "" {
// 			t.Error(testcallers(), diff)
// 		}
// 	}
//
// 	t.Run("basic", func(t *testing.T) {
// 		t.Parallel()
// 		var tt TT
// 		tt.item = &AddColumnCommand{}
// 		tt.wantQuery = "actor.actor_id"
// 		assert(t, tt)
// 	})
// }
//
// func Test_RenameColumnCommand(t *testing.T) {
// 	type TT struct {
// 		dialect   string
// 		item      Command
// 		wantQuery string
// 		wantArgs  []interface{}
// 	}
//
// 	assert := func(t *testing.T, tt TT) {
// 		gotQuery, gotArgs, _, err := sq.ToSQL(tt.dialect, tt.item)
// 		if err != nil {
// 			t.Fatal(testcallers(), err)
// 		}
// 		if diff := testdiff(gotQuery, tt.wantQuery); diff != "" {
// 			t.Error(testcallers(), diff)
// 		}
// 		if diff := testdiff(gotArgs, tt.wantArgs); diff != "" {
// 			t.Error(testcallers(), diff)
// 		}
// 	}
//
// 	t.Run("basic", func(t *testing.T) {
// 		t.Parallel()
// 		var tt TT
// 		tt.item = &AddColumnCommand{}
// 		tt.wantQuery = "actor.actor_id"
// 		assert(t, tt)
// 	})
// }
