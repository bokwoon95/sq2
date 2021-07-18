package ddl

import (
	"testing"

	"github.com/bokwoon95/sq"
)

func Test_Function(t *testing.T) {
	type TT struct {
		dialect            string
		item               Function
		wantFunctionSchema string
		wantFunctionName   string
		wantArgModes       []string
		wantArgNames       []string
		wantArgTypes       []string
	}

	assert := func(t *testing.T, tt TT) {
		err := tt.item.populateFunctionInfo(tt.dialect)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		sql, _, _, err := sq.ToSQL(tt.dialect, CreateFunctionCommand{tt.item})
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(sql, tt.item.SQL); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.item.FunctionSchema, tt.wantFunctionSchema); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.item.FunctionName, tt.wantFunctionName); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.item.ArgModes, tt.wantArgModes); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.item.ArgNames, tt.wantArgNames); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.item.ArgTypes, tt.wantArgTypes); diff != "" {
			t.Error(testcallers(), diff)
		}
	}

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION one() RETURNS integer`
		tt.wantFunctionName = "one"
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION app.tf1 (integer, in numeric = 3.14) RETURNS integer`
		tt.wantFunctionSchema = "app"
		tt.wantFunctionName = "tf1"
		tt.wantArgModes = []string{"", "in"}
		tt.wantArgNames = []string{"", ""}
		tt.wantArgTypes = []string{"integer", "numeric"}
		assert(t, tt)
	})
}
