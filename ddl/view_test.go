package ddl

import "testing"

func Test_View(t *testing.T) {
	t.Run("Index", func(t *testing.T) {
		t.Parallel()
		var view View
		assertPosition := func(name string, wantPosition int) {
			gotPosition := view.CachedIndexPosition(name)
			if diff := testdiff(gotPosition, wantPosition); diff != "" {
				t.Error(testcallers(), diff)
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
			if diff := testdiff(gotPosition, wantPosition); diff != "" {
				t.Error(testcallers(), diff)
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
		view.Triggers[2].TriggerName = "J"
		assertPosition("G", "H", "I", -1)
		view.triggerCache = nil
		view.RefreshTriggerCache()
		assertPosition("A", "B", "C", 0)
		assertPosition("D", "E", "F", 1)
		assertPosition("G", "H", "J", 2)
	})
}
