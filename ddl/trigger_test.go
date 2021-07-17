package ddl

import (
	"testing"

	"github.com/bokwoon95/sq"
)

func Test_Trigger(t *testing.T) {
	type TT struct {
		dialect         string
		item            Trigger
		wantTableSchema string
		wantTableName   string
		wantTriggerName string
	}

	assert := func(t *testing.T, tt TT) {
		err := tt.item.populateTriggerInfo(tt.dialect)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		sql, _, _, err := sq.ToSQL(tt.dialect, CreateTriggerCommand{tt.item})
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(sql, tt.item.SQL); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.item.TableSchema, tt.wantTableSchema); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.item.TableName, tt.wantTableName); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.item.TriggerName, tt.wantTriggerName); diff != "" {
			t.Error(testcallers(), diff)
		}
	}

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE TRIGGER actor_last_update_before_update_trg BEFORE UPDATE ON public.actor FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`
		tt.wantTableSchema = "public"
		tt.wantTableName = "actor"
		tt.wantTriggerName = "actor_last_update_before_update_trg"
		assert(t, tt)
	})

	t.Run("(dialect == sqlite)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item.SQL = `
CREATE TRIGGER actor_last_update_after_update_trg AFTER UPDATE ON actor BEGIN
    UPDATE actor SET last_update = DATETIME('now') WHERE actor_id = NEW.actor_id;
END;`
		tt.wantTableSchema = ""
		tt.wantTableName = "actor"
		tt.wantTriggerName = "actor_last_update_after_update_trg"
		assert(t, tt)
	})

	t.Run("(dialect == sqlite) IF NOT EXISTS", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item.SQL = `
CREATE TRIGGER IF NOT EXISTS actor_last_update_after_update_trg AFTER UPDATE ON actor BEGIN
    UPDATE actor SET last_update = DATETIME('now') WHERE actor_id = NEW.actor_id;
END;`
		tt.wantTableSchema = ""
		tt.wantTableName = "actor"
		tt.wantTriggerName = "actor_last_update_after_update_trg"
		assert(t, tt)
	})

	t.Run("junk trigger", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `the quick brown fox jumped over the lazy dog`
		err := tt.item.populateTriggerInfo(tt.dialect)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})
}
