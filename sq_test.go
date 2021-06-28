package sq

import (
	"bytes"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type tmptable string

var _ Table = tmptable("")

func (t tmptable) GetAlias() string { return "" }

func (t tmptable) GetName() string { return string(t) }

func (t tmptable) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString(QuoteIdentifier(dialect, string(t)))
	return nil
}

type tmpfield [2]string

var _ Field = tmpfield{}

func (f tmpfield) GetAlias() string { return "" }

func (f tmpfield) GetName() string { return f[1] }

func (f tmpfield) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	tableQualifier := f[0]
	if tableQualifier != "" {
		for _, excludedTableQualifier := range excludedTableQualifiers {
			if tableQualifier == excludedTableQualifier {
				tableQualifier = ""
				break
			}
		}
	}
	if tableQualifier != "" {
		buf.WriteString(QuoteIdentifier(dialect, tableQualifier) + ".")
	}
	buf.WriteString(QuoteIdentifier(dialect, f[1]))
	return nil
}

func Diff(got, want interface{}) string {
	diff := cmp.Diff(got, want, cmp.Exporter(func(typ reflect.Type) bool { return true }))
	if diff != "" {
		return "\n-got +want\n" + diff
	}
	return ""
}

func Callers() string {
	var pc [50]uintptr
	// Skip two extra frames to account for this function
	// and runtime.Callers itself.
	n := runtime.Callers(2, pc[:])
	if n == 0 {
		panic("zero callers found")
	}
	var callsites []string
	frames := runtime.CallersFrames(pc[:n])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		callsites = append(callsites, filepath.Base(frame.File)+":"+strconv.Itoa(frame.Line))
	}
	buf := &strings.Builder{}
	last := len(callsites) - 2
	buf.WriteString("[")
	for i := last; i >= 0; i-- {
		if i < last {
			buf.WriteString(" -> ")
		}
		buf.WriteString(callsites[i])
	}
	buf.WriteString("]")
	return buf.String()
}

type FaultySQLError struct{}

func (e FaultySQLError) Error() string { return "sql broke" }

var ErrFaultySQL error = FaultySQLError{}

type FaultySQL struct{}

var (
	_ Query     = FaultySQL{}
	_ BaseTable = FaultySQL{}
	_ Field     = FaultySQL{}
	_ Predicate = FaultySQL{}
)

func (q FaultySQL) AppendSQL(string, *bytes.Buffer, *[]interface{}, map[string][]int) error {
	return ErrFaultySQL
}

func (q FaultySQL) SetFetchableFields([]Field) (Query, error) {
	return nil, ErrFaultySQL
}

func (q FaultySQL) GetFetchableFields() ([]Field, error) {
	return nil, ErrFaultySQL
}

func (q FaultySQL) GetDialect() string { return "" }

func (q FaultySQL) AppendSQLExclude(string, *bytes.Buffer, *[]interface{}, map[string][]int, []string) error {
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
			t.Fatalf("%s expected slice %#v to be explodable", Callers(), tt.slice)
		}
		err := explodeSlice(tt.dialect, buf, &gotArgs, gotParams, tt.excludedTableQualifiers, tt.slice)
		if err != nil {
			t.Fatal(Callers(), err)
		}
		if diff := Diff(buf.String(), tt.wantQuery); diff != "" {
			t.Error(Callers(), diff)
		}
		if diff := Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(Callers(), diff)
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
		tt.slice = []tmptable{"111", "222", "333"}
		tt.excludedTableQualifiers = []string{"222"}
		tt.wantQuery = `"111", "222", "333"`
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})
}
