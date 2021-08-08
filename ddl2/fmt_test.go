package ddl

import (
	"testing"

	"github.com/bokwoon95/sq"
)

func Test_lexModifiers(t *testing.T) {
	type TT struct {
		config            string
		wantModifiers     [][2]string
		wantModifierIndex map[string]int
	}

	assert := func(t *testing.T, tt TT) {
		gotModifiers, gotModifierIndex, err := tokenizeModifiers(tt.config)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(tt.wantModifiers, gotModifiers); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.wantModifierIndex, gotModifierIndex); diff != "" {
			t.Error(testcallers(), diff)
		}
	}

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.config = ""
		tt.wantModifiers = nil
		tt.wantModifierIndex = map[string]int{}
		assert(t, tt)
	})

	t.Run("test1", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.config = "notnull unique index={. unique} name=testing references={inventory onupdate=cascade ondelete=restrict}"
		tt.wantModifiers = [][2]string{
			{"notnull", ""},
			{"unique", ""},
			{"index", ". unique"},
			{"name", "testing"},
			{"references", "inventory onupdate=cascade ondelete=restrict"},
		}
		tt.wantModifierIndex = map[string]int{
			"notnull":    0,
			"unique":     1,
			"index":      2,
			"name":       3,
			"references": 4,
		}
		assert(t, tt)
	})

	t.Run("test2", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.config = "cols=a,b,c index={. where={email LIKE '%gmail'}}"
		tt.wantModifiers = [][2]string{
			{"cols", "a,b,c"},
			{"index", ". where={email LIKE '%gmail'}"},
		}
		tt.wantModifierIndex = map[string]int{
			"cols":  0,
			"index": 1,
		}
		assert(t, tt)
	})
}

func Test_lexValue(t *testing.T) {
	type TT struct {
		config            string
		wantValue         string
		wantModifiers     [][2]string
		wantModifierIndex map[string]int
	}

	assert := func(t *testing.T, tt TT) {
		gotValue, gotModifiers, gotModifierIndex, err := tokenizeValue(tt.config)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		if diff := testdiff(tt.wantValue, gotValue); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.wantModifiers, gotModifiers); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(tt.wantModifierIndex, gotModifierIndex); diff != "" {
			t.Error(testcallers(), diff)
		}
	}

	t.Run("", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.config = ""
		tt.wantModifiers = nil
		tt.wantModifierIndex = map[string]int{}
		assert(t, tt)
	})

	t.Run("", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.config = "1 unique"
		tt.wantValue = "1"
		tt.wantModifiers = [][2]string{
			{"unique", ""},
		}
		tt.wantModifierIndex = map[string]int{
			"unique": 0,
		}
		assert(t, tt)
	})

	t.Run("", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.config = "{abcd efg} generated={first_name || ' ' || last_name} virtual name=gg=G"
		tt.wantValue = "abcd efg"
		tt.wantModifiers = [][2]string{
			{"generated", "first_name || ' ' || last_name"},
			{"virtual", ""},
			{"name", "gg=G"},
		}
		tt.wantModifierIndex = map[string]int{
			"generated": 0,
			"virtual":   1,
			"name":      2,
		}
		assert(t, tt)
	})

	t.Run("", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.config = "inventory cols=1,2,3,4 onupdate=cascade ondelete=restrict"
		tt.wantValue = "inventory"
		tt.wantModifiers = [][2]string{
			{"cols", "1,2,3,4"},
			{"onupdate", "cascade"},
			{"ondelete", "restrict"},
		}
		tt.wantModifierIndex = map[string]int{
			"cols":     0,
			"onupdate": 1,
			"ondelete": 2,
		}
		assert(t, tt)
	})
}

func Test_popWords(t *testing.T) {
	type TT struct {
		dialect   string
		s         string
		num       int
		wantWords []string
		wantRest  string
	}

	assert := func(t *testing.T, tt TT) {
		gotWords, gotRest, _ := popIdentifierTokens(tt.dialect, tt.s, tt.num)
		if diff := testdiff(gotWords, tt.wantWords); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(gotRest, tt.wantRest); diff != "" {
			t.Error(testcallers(), diff)
		}
	}

	t.Run("empty (popWord)", func(t *testing.T) {
		t.Parallel()
		gotWord, gotRest, _ := popIdentifierToken("", "")
		if diff := testdiff(gotWord, ""); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(gotRest, ""); diff != "" {
			t.Error(testcallers(), diff)
		}
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.s = ""
		tt.num = 3
		tt.wantWords = nil
		tt.wantRest = ""
		assert(t, tt)
	})

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.s = " the\n    quick\r\n\tbrown      fox  jumped over the lazy dog"
		tt.num = 4
		tt.wantWords = []string{"the", "quick", "brown", "fox"}
		tt.wantRest = "  jumped over the lazy dog"
		assert(t, tt)
	})

	t.Run("mysql quoted identifier", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.dialect = sq.DialectMySQL
		tt.s = "CREATE TRIGGER \"a \"\"b c\".`d e f``ghi` AS"
		tt.num = 4
		tt.wantWords = []string{"CREATE", "TRIGGER", "\"a \"\"b c\".`d e f``ghi`", "AS"}
		tt.wantRest = ""
		assert(t, tt)
	})

	t.Run("sqlserver quoted identifier", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.s = "CREATE TRIGGER \"a \"\"b c\".[d e f]]ghi] AS"
		tt.dialect = sq.DialectSQLServer
		tt.num = 4
		tt.wantWords = []string{"CREATE", "TRIGGER", "\"a \"\"b c\".[d e f]]ghi]", "AS"}
		tt.wantRest = ""
		assert(t, tt)
	})
}

func Test_splitArgs(t *testing.T) {
	t.Run("", func(t *testing.T) {
		gotArgs := splitArgs(`salary_val IN decimal, alphabets []TEXT='{"a", "b", "c"}', names VARIADIC [][]text = ARRAY[ARRAY['a', 'b'], ARRAY['c', 'd']]`)
		wantArgs := []string{
			"salary_val IN decimal",
			` alphabets []TEXT='{"a", "b", "c"}'`,
			" names VARIADIC [][]text = ARRAY[ARRAY['a', 'b'], ARRAY['c', 'd']]",
		}
		if diff := testdiff(gotArgs, wantArgs); diff != "" {
			t.Error(testcallers(), diff)
		}
	})
}
