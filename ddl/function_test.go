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

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE OR REPLACE FUNCTION double_salary(emp) RETURNS numeric`
		tt.wantFunctionName = "double_salary"
		tt.wantArgModes = []string{""}
		tt.wantArgNames = []string{""}
		tt.wantArgTypes = []string{"emp"}
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION sum_n_product (int, y int, OUT sum int DEFAULT 3, OUT product int=22)`
		tt.wantFunctionName = "sum_n_product"
		tt.wantArgModes = []string{"", "", "OUT", "OUT"}
		tt.wantArgNames = []string{"", "y", "sum", "product"}
		tt.wantArgTypes = []string{"int", "int", "int", "int"}
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION make_array(anyelement, anyelement) RETURNS anyarray`
		tt.wantFunctionName = "make_array"
		tt.wantArgModes = []string{"", ""}
		tt.wantArgNames = []string{"", ""}
		tt.wantArgTypes = []string{"anyelement", "anyelement"}
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `
CREATE OR REPLACE FUNCTION years_compare( IN year1 integer DEFAULT NULL,
                                          IN year2 integer DEFAULT NULL )`
		tt.wantFunctionName = "years_compare"
		tt.wantArgModes = []string{"IN", "IN"}
		tt.wantArgNames = []string{"year1", "year2"}
		tt.wantArgTypes = []string{"integer", "integer"}
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `create function foo(bar varchar , baz varchar='qux') returns varchar`
		tt.wantFunctionName = "foo"
		tt.wantArgModes = []string{"", ""}
		tt.wantArgNames = []string{"bar", "baz"}
		tt.wantArgTypes = []string{"varchar", "varchar"}
		assert(t, tt)
	})
}
