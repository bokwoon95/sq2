package sq

import (
	"bytes"
	"sort"
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
)

type tmptable [2]string

var _ Table = tmptable{}

func (t tmptable) GetAlias() string { return "" }

func (t tmptable) GetName() string { return string(t[1]) }

func (t tmptable) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	if t[0] != "" {
		buf.WriteString(QuoteIdentifier(dialect, string(t[0])) + ".")
	}
	buf.WriteString(QuoteIdentifier(dialect, string(t[1])))
	return nil
}

type tmpfield [2]string

var _ Field = tmpfield{}

func (f tmpfield) GetAlias() string { return "" }

func (f tmpfield) GetName() string { return f[1] }

func (f tmpfield) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	tableQualifier := f[0]
	if tableQualifier != "" {
		i := sort.SearchStrings(excludedTableQualifiers, tableQualifier)
		if i < len(excludedTableQualifiers) && excludedTableQualifiers[i] == tableQualifier {
			tableQualifier = ""
		}
	}
	if tableQualifier != "" {
		buf.WriteString(QuoteIdentifier(dialect, tableQualifier) + ".")
	}
	buf.WriteString(QuoteIdentifier(dialect, f[1]))
	return nil
}

type FaultySQLError struct{}

func (e FaultySQLError) Error() string { return "sql broke" }

var ErrFaultySQL error = FaultySQLError{}

type FaultySQL struct{}

var (
	_ Query       = FaultySQL{}
	_ SchemaTable = FaultySQL{}
	_ Field       = FaultySQL{}
	_ Predicate   = FaultySQL{}
)

func (q FaultySQL) AppendSQL(string, *bytes.Buffer, *[]interface{}, map[string][]int, map[string]interface{}) error {
	return ErrFaultySQL
}

func (q FaultySQL) SetFetchableFields([]Field) (Query, error) {
	return nil, ErrFaultySQL
}

func (q FaultySQL) GetFetchableFields() ([]Field, error) {
	return nil, ErrFaultySQL
}

func (q FaultySQL) GetDialect() string { return "" }

func (q FaultySQL) AppendSQLExclude(string, *bytes.Buffer, *[]interface{}, map[string][]int, map[string]interface{}, []string) error {
	return ErrFaultySQL
}

func (q FaultySQL) GetAlias() string { return "" }

func (q FaultySQL) GetName() string { return "" }

func (q FaultySQL) GetSchema() string { return "" }

func (q FaultySQL) Not() Predicate { return q }

func Test_explodeSlice(t *testing.T) {
	type TT struct {
		dialect                 string
		slice                   interface{}
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, gotParams := []interface{}{}, map[string][]int{}
		if !isExplodableSlice(tt.slice) {
			t.Fatalf("%s expected slice %#v to be explodable", testutil.Callers(), tt.slice)
		}
		err := explodeSlice(tt.dialect, buf, &gotArgs, gotParams, tt.excludedTableQualifiers, tt.slice)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(buf.String(), tt.wantQuery); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
		if diff := testutil.Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
	}

	t.Run("tmpfield slice", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.slice = []tmpfield{{"111", "aaa"}, {"222", "bbb"}, {"333", "ccc"}}
		tt.excludedTableQualifiers = []string{"222"}
		tt.wantQuery = `"111".aaa, bbb, "333".ccc`
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("tmptable slice", func(t *testing.T) {
		t.Parallel()
		var tt TT
		tt.slice = []tmptable{[2]string{"", "111"}, [2]string{"", "222"}, [2]string{"", "333"}}
		tt.excludedTableQualifiers = []string{"222"}
		tt.wantQuery = `"111", "222", "333"`
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})
}
