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
		wantReturnType     string
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
		if diff := testdiff(tt.item.ReturnType, tt.wantReturnType); diff != "" {
			t.Error(testcallers(), diff)
		}
	}

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION one() RETURNS integer AS $$ lorem ipsum $$ LANGUAGE plpgsql`
		tt.wantFunctionName = "one"
		tt.wantReturnType = "integer"
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION one(     ) RETURNS integer`
		tt.wantFunctionName = "one"
		tt.wantReturnType = "integer"
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION app.tf1 (integer, in numeric = 3.14) RETURNS integer LANGUAGE sql AS $$ lorem ipsum $$`
		tt.wantFunctionSchema = "app"
		tt.wantFunctionName = "tf1"
		tt.wantArgModes = []string{"", "in"}
		tt.wantArgNames = []string{"", ""}
		tt.wantArgTypes = []string{"integer", "numeric"}
		tt.wantReturnType = "integer"
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE OR REPLACE FUNCTION double_salary(emp) RETURNS numeric AS $$ lorem ipsum $$ LANGUAGE sql`
		tt.wantFunctionName = "double_salary"
		tt.wantArgModes = []string{""}
		tt.wantArgNames = []string{""}
		tt.wantArgTypes = []string{"emp"}
		tt.wantReturnType = "numeric"
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION sum_n_product (int, y int =25, OUT sum int DEFAULT 3, OUT product int=22)`
		tt.wantFunctionName = "sum_n_product"
		tt.wantArgModes = []string{"", "", "OUT", "OUT"}
		tt.wantArgNames = []string{"", "y", "sum", "product"}
		tt.wantArgTypes = []string{"int", "int", "int", "int"}
		tt.wantReturnType = ""
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION make_array(anyelement, anyelement) RETURNS anyarray AS LANGUAGE plpgsql $$ lorem ipsum $$`
		tt.wantFunctionName = "make_array"
		tt.wantArgModes = []string{"", ""}
		tt.wantArgNames = []string{"", ""}
		tt.wantArgTypes = []string{"anyelement", "anyelement"}
		tt.wantReturnType = "anyarray"
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `
CREATE OR REPLACE FUNCTION years_compare( IN year1 integer DEFAULT NULL,
                                          year2 IN integer DEFAULT NULL ) RETURNS BOOLEAN AS $$ lorem ipsum $$ language SQL`
		tt.wantFunctionName = "years_compare"
		tt.wantArgModes = []string{"IN", "IN"}
		tt.wantArgNames = []string{"year1", "year2"}
		tt.wantArgTypes = []string{"integer", "integer"}
		tt.wantReturnType = "BOOLEAN"
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `create function foo(bar varchar , baz varchar='qux') returns varchar AS $$ lorem ipsum $$ language sql`
		tt.wantFunctionName = "foo"
		tt.wantArgModes = []string{"", ""}
		tt.wantArgNames = []string{"bar", "baz"}
		tt.wantArgTypes = []string{"varchar", "varchar"}
		tt.wantReturnType = "varchar"
		assert(t, tt)
	})

	t.Run("(dialect == postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION get_count_of_earners(salary_val IN decimal, alphabets []TEXT='{"a", "b", "c"}', names VARIADIC [][]text = ARRAY[ARRAY['a', 'b'], ARRAY['c', 'd']]) RETURNS integer AS $$ lorem ipsum $$ language plpgsql`
		tt.wantFunctionName = "get_count_of_earners"
		tt.wantArgModes = []string{"IN", "", "VARIADIC"}
		tt.wantArgNames = []string{"salary_val", "alphabets", "names"}
		tt.wantArgTypes = []string{"decimal", "[]TEXT", "[][]text"}
		tt.wantReturnType = "integer"
		assert(t, tt)
	})

	t.Run("(dialect != postgres)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item.SQL = `CREATE FUNCTION hello (s CHAR(20)) RETURNS CHAR(50) DETERMINISTIC RETURN CONCAT('Hello, ',s,'!')`
		assert(t, tt)
	})

	t.Run("(dialect == postgres) no opening bracket", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION temp`
		err := tt.item.populateFunctionInfo(tt.dialect)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) no closing bracket", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION temp(`
		err := tt.item.populateFunctionInfo(tt.dialect)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) empty args", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION temp(,,,)`
		err := tt.item.populateFunctionInfo(tt.dialect)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) invalid args", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE FUNCTION temp(   DEFAULT 'test',='test')`
		err := tt.item.populateFunctionInfo(tt.dialect)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == postgres) invalid function", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectPostgres
		tt.item.SQL = `CREATE temp()`
		err := tt.item.populateFunctionInfo(tt.dialect)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})
}

func Test_DropFunctionCommand(t *testing.T) {
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
		tt.item = DropFunctionCommand{
			DropIfExists: true,
			Function: Function{
				FunctionSchema: "public",
				FunctionName:   "my_function",
				ArgNames:       []string{"IN", "IN"},
				ArgModes:       []string{"arg_str", "arg_num"},
				ArgTypes:       []string{"TEXT", "INT"},
			},
			DropCascade: true,
		}
		tt.wantQuery = `DROP FUNCTION IF EXISTS public.my_function(TEXT, INT) CASCADE`
		assert(t, tt)
	})

	t.Run("(dialect == sqlite)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item = DropFunctionCommand{
			Function: Function{FunctionName: "my_function"},
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})
}

func Test_RenameFunctionCommand(t *testing.T) {
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
		tt.item = RenameFunctionCommand{
			Function: Function{
				FunctionSchema: "public",
				FunctionName:   "my_function",
			},
			RenameToName: "my_new_function",
		}
		tt.wantQuery = `ALTER FUNCTION public.my_function() RENAME TO my_new_function`
		assert(t, tt)
	})

	t.Run("(dialect == sqlite)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectSQLite
		tt.item = RenameFunctionCommand{
			Function:     Function{FunctionName: "my_function"},
			RenameToName: "my_new_function",
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})

	t.Run("(dialect == mysql)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.item = RenameFunctionCommand{
			Function:     Function{FunctionName: "my_function"},
			RenameToName: "my_new_function",
		}
		_, _, _, err := sq.ToSQL(tt.dialect, tt.item)
		if err == nil {
			t.Fatal(testcallers(), "expected error but got nil")
		}
	})
}
