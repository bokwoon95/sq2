package ddl

import (
	"testing"

	"github.com/bokwoon95/sq"
	"github.com/bokwoon95/sq/internal/testutil"
)

func Test_View(t *testing.T) {
	t.Run("Index", func(t *testing.T) {
		t.Parallel()
		var view View
		assertPosition := func(name string, wantPosition int) {
			gotPosition := view.CachedIndexPosition(name)
			if diff := testutil.Diff(gotPosition, wantPosition); diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		}
		view.AppendIndex(Index{IndexName: "A"})
		view.AppendIndex(Index{IndexName: "B"})
		view.AppendIndex(Index{IndexName: "C"})
		assertPosition("A", 0)
		assertPosition("B", 1)
		assertPosition("C", 2)
		assertPosition("", -1)
		assertPosition("does not exist", -1)
		view.Indexes[2].IndexName = "D"
		assertPosition("C", -1)
		view.indexCache = nil
		view.RefreshIndexCache()
		assertPosition("A", 0)
		assertPosition("B", 1)
		assertPosition("D", 2)
	})

	t.Run("Trigger", func(t *testing.T) {
		t.Parallel()
		var view View
		assertPosition := func(schema, tableName, triggerName string, wantPosition int) {
			gotPosition := view.CachedTriggerPosition(schema, tableName, triggerName)
			if diff := testutil.Diff(gotPosition, wantPosition); diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		}
		view.AppendTrigger(Trigger{TableSchema: "A", TableName: "B", TriggerName: "C"})
		view.AppendTrigger(Trigger{TableSchema: "D", TableName: "E", TriggerName: "F"})
		view.AppendTrigger(Trigger{TableSchema: "G", TableName: "H", TriggerName: "I"})
		assertPosition("A", "B", "C", 0)
		assertPosition("D", "E", "F", 1)
		assertPosition("G", "H", "I", 2)
		assertPosition("", "", "", -1)
		assertPosition("", "", "does not exist", -1)
		view.triggerCache[[3]string{"D", "E", "F"}] = 999
		view.Triggers[2].TriggerName = "J"
		assertPosition("D", "E", "F", -1)
		assertPosition("G", "H", "I", -1)
		view.triggerCache = nil
		view.RefreshTriggerCache()
		assertPosition("A", "B", "C", 0)
		assertPosition("D", "E", "F", 1)
		assertPosition("G", "H", "J", 2)
	})

	t.Run("createOrUpdateIndex", func(t *testing.T) {
		t.Parallel()
		var view View
		createOrUpdateIndex := func(indexName string, columns, exprs []string, wantPosition int) {
			gotPosition, err := view.createOrUpdateIndex(indexName, columns, exprs)
			if err != nil {
				t.Fatal(testutil.Callers(), err)
			}
			if diff := testutil.Diff(gotPosition, wantPosition); diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		}
		_, err := view.createOrUpdateIndex("", nil, nil)
		if err == nil {
			t.Error(testutil.Callers(), "expected error but got nil")
		}
		createOrUpdateIndex("A", nil, nil, 0)
		createOrUpdateIndex("B", nil, nil, 1)
		createOrUpdateIndex("C", nil, nil, 2)
		createOrUpdateIndex("A", []string{"a"}, nil, 0)
		createOrUpdateIndex("B", []string{"b"}, nil, 1)
		createOrUpdateIndex("C", []string{"c"}, nil, 2)
	})
}

func Test_CreateViewCommand(t *testing.T) {
	type TT struct {
		dialect   string
		item      Command
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := sq.ToSQL(tt.dialect, tt.item, nil)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotQuery, tt.wantQuery); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
		if diff := testutil.Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
	}

	t.Run("(dialect == postgres) create or replace", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = CreateViewCommand{
			CreateOrReplace: true,
			View: View{
				ViewSchema: "some table",
				ViewName:   "some view",
				SQL:        "SELECT 1",
			},
		}
		tt.wantQuery = `CREATE OR REPLACE VIEW "some table"."some view" AS SELECT 1`
		assert(t, tt)
	})

	t.Run("(dialect == postgres) create materialized view if not exists", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item = CreateViewCommand{
			CreateIfNotExists: true,
			View: View{
				IsMaterialized: true,
				ViewSchema:     "some table",
				ViewName:       "some view",
				SQL:            "SELECT 1",
			},
		}
		tt.wantQuery = `CREATE MATERIALIZED VIEW IF NOT EXISTS "some table"."some view" AS SELECT 1`
		assert(t, tt)
	})
}
