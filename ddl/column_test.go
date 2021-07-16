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

	var _ sq.Field = Column{}

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

	t.Run("(dialect == postgres) IDENTITY column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
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

	t.Run("(dialect == postgres) GENERATED column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
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

	t.Run("(dialect == postgres) AUTOINCREMENT column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName:    "actor_id",
				ColumnType:    "INT",
				Autoincrement: true,
			},
		}
		tt.dialect = sq.DialectPostgres
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) GENERATED VIRTUAL column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName:    "full_name",
				ColumnType:    "TEXT",
				GeneratedExpr: "first_name || ' ' || last_name",
			},
		}
		tt.dialect = sq.DialectPostgres
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect != postgres) add column if not exists", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			AddIfNotExists: true,
			Column: Column{
				ColumnName: "first_name",
				ColumnType: "TEXT",
			},
		}
		tt.dialect = sq.DialectSQLite
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect != mysql) column with ON UPDATE CURRENT_TIMESTAMP", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName:               "last_update",
				ColumnType:               "TIMESTAMPTZ",
				OnUpdateCurrentTimestamp: true,
			},
		}
		tt.dialect = sq.DialectPostgres
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == sqlite) PRIMARY KEY column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName:   "actor_id",
				ColumnType:   "INTEGER",
				IsPrimaryKey: true,
			},
		}
		tt.dialect = sq.DialectSQLite
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == sqlite) UNIQUE column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName: "actor_id",
				ColumnType: "INTEGER",
				IsUnique:   true,
			},
		}
		tt.dialect = sq.DialectSQLite
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == sqlite) NOT NULL column without a DEFAULT value", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName: "actor_id",
				ColumnType: "INTEGER",
				IsNotNull:  true,
			},
		}
		tt.dialect = sq.DialectSQLite
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == sqlite) NOT NULL column with an expression as DEFAULT value", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName:    "actor_id",
				ColumnType:    "INTEGER",
				IsNotNull:     true,
				ColumnDefault: "(1 + 2 + 3)",
			},
		}
		tt.dialect = sq.DialectSQLite
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == sqlite) FOREIGN KEY column with non-null DEFAULT value", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName:    "country_id",
				ColumnType:    "INT",
				ColumnDefault: "22",
			},
			ReferencesTable:  "country",
			ReferencesColumn: "country_id",
		}
		tt.dialect = sq.DialectSQLite
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == sqlite) GENERATED STORED column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName:          "full_name",
				ColumnType:          "TEXT",
				GeneratedExpr:       "first_name || ' ' || last_name",
				GeneratedExprStored: true,
			},
		}
		tt.dialect = sq.DialectSQLite
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == sqlite) IDENTITY column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName: "actor_id",
				ColumnType: "INT",
				Identity:   BY_DEFAULT_AS_IDENTITY,
			},
		}
		tt.dialect = sq.DialectSQLite
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == sqlite) AUTOINCREMENT column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName:    "actor_id",
				ColumnType:    "INT",
				Autoincrement: true,
			},
		}
		tt.dialect = sq.DialectSQLite
		tt.wantQuery = "ADD COLUMN actor_id INT AUTOINCREMENT"
		assert(t, tt)
	})

	t.Run("(dialect == sqlite) column with COLLATE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName:    "first_name",
				ColumnType:    "TEXT",
				CollationName: "nocase",
			},
		}
		tt.dialect = sq.DialectSQLite
		tt.wantQuery = "ADD COLUMN first_name TEXT COLLATE nocase"
		assert(t, tt)
	})

	t.Run("(dialect == sqlite) GENERATED VIRTUAL column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName:    "full_name",
				ColumnType:    "TEXT",
				GeneratedExpr: "first_name || ' ' || last_name",
			},
		}
		tt.dialect = sq.DialectSQLite
		tt.wantQuery = "ADD COLUMN full_name TEXT GENERATED ALWAYS AS (first_name || ' ' || last_name) VIRTUAL"
		assert(t, tt)
	})

	t.Run("(dialect == sqlite) column with constraints", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName: "country_id",
				ColumnType: "INT",
			},
			CheckExprs: []string{
				"country_id > 0",
				"country_id IS NOT NULL",
			},
			ReferencesTable:  "country",
			ReferencesColumn: "country_id",
		}
		tt.dialect = sq.DialectSQLite
		tt.wantQuery = "ADD COLUMN country_id INT CHECK (country_id > 0) CHECK (country_id IS NOT NULL) REFERENCES country (country_id)"
		assert(t, tt)
	})

	t.Run("(dialect == mysql) AUTOINCREMENT column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName:    "actor_id",
				ColumnType:    "INT",
				Autoincrement: true,
			},
		}
		tt.dialect = sq.DialectMySQL
		tt.wantQuery = "ADD COLUMN actor_id INT AUTO_INCREMENT"
		assert(t, tt)
	})

	t.Run("(dialect == mysql) column with ON UPDATE CURRENT_TIMESTAMP", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.item = AddColumnCommand{
			Column: Column{
				ColumnName:               "last_update",
				ColumnType:               "DATETIME",
				ColumnDefault:            "CURRENT_TIMESTAMP",
				OnUpdateCurrentTimestamp: true,
			},
		}
		tt.dialect = sq.DialectMySQL
		tt.wantQuery = "ADD COLUMN last_update DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"
		assert(t, tt)
	})
}

func Test_AlterColumnCommand(t *testing.T) {
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

	t.Run("dialect == sqlite", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item = AlterColumnCommand{}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("dialect == mysql", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = AlterColumnCommand{
			Column: Column{
				ColumnName:    "replacement_cost",
				ColumnType:    "DECIMAL(5,2)",
				IsNotNull:     true,
				ColumnDefault: "19.99",
			},
		}
		tt.wantQuery = "MODIFY COLUMN replacement_cost DECIMAL(5,2) NOT NULL DEFAULT 19.99"
		assert(t, tt)
	})

	t.Run("(dialect == mysql) IDENTITY column", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = AlterColumnCommand{
			Column: Column{
				ColumnName: "actor_id",
				ColumnType: "INT",
				Identity:   BY_DEFAULT_AS_IDENTITY,
			},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) set column attributes", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AlterColumnCommand{
			Column: Column{
				ColumnName:    "replacement_cost",
				ColumnType:    "TEXT",
				IsNotNull:     true,
				ColumnDefault: "'19.99'",
				CollationName: "C",
				Identity:      BY_DEFAULT_AS_IDENTITY,
			},
			UsingExpr: "replacement_cost::TEXT",
		}
		tt.wantQuery = `ALTER COLUMN replacement_cost SET DATA TYPE TEXT COLLATE "C" USING replacement_cost::TEXT
,ALTER COLUMN replacement_cost SET NOT NULL
,ALTER COLUMN replacement_cost SET DEFAULT '19.99'
,ALTER COLUMN replacement_cost ADD GENERATED BY DEFAULT AS IDENTITY`
		assert(t, tt)
	})

	t.Run("(dialect == postgres) drop column attributes", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = AlterColumnCommand{
			Column: Column{
				ColumnName: "replacement_cost",
			},
			DropDefault:          true,
			DropNotNull:          true,
			DropExpr:             true,
			DropExprIfExists:     true,
			DropIdentity:         true,
			DropIdentityIfExists: true,
		}
		tt.wantQuery = `ALTER COLUMN replacement_cost DROP NOT NULL
,ALTER COLUMN replacement_cost DROP DEFAULT
,ALTER COLUMN replacement_cost DROP IDENTITY IF EXISTS
,ALTER COLUMN replacement_cost DROP EXPRESSION IF EXISTS`
		assert(t, tt)
	})

	t.Run("unrecognized dialect", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = "abcdefg"
		tt.item = AlterColumnCommand{
			Column: Column{
				ColumnName: "replacement_cost",
				ColumnType: "TEXT",
			},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})
}

func Test_DropColumnCommand(t *testing.T) {
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

	t.Run("(dialect != postgres) DROP", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item = DropColumnCommand{
			ColumnName: "actor_id",
		}
		tt.wantQuery = "DROP COLUMN actor_id"
		assert(t, tt)
	})

	t.Run("(dialect != postgres) DROP IF EXISTS", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item = DropColumnCommand{
			DropIfExists: true,
			ColumnName:   "actor_id",
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect != postgres) DROP ... CASCADE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item = DropColumnCommand{
			ColumnName:  "actor_id",
			DropCascade: true,
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) DROP IF EXISTS CASCADE", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = DropColumnCommand{
			DropIfExists: true,
			ColumnName:   "actor_id",
			DropCascade:  true,
		}
		tt.wantQuery = "DROP COLUMN IF EXISTS actor_id CASCADE"
		assert(t, tt)
	})
}

func Test_RenameColumnCommand(t *testing.T) {
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

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = RenameColumnCommand{
			ColumnName:   "actor_id",
			RenameToName: "actor ID",
		}
		tt.wantQuery = `RENAME COLUMN actor_id TO "actor ID"`
		assert(t, tt)
	})
}
