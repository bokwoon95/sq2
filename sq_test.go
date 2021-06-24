package sq

import (
	"bytes"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/bokwoon95/testutil"
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

func testdiff(lhs, rhs interface{}) string {
	diff := cmp.Diff(lhs, rhs, cmp.Exporter(func(typ reflect.Type) bool { return true }))
	if diff != "" {
		return "\n -lhs +rhs\n" + diff
	}
	return ""
}

func testcallers() string {
	var pc [50]uintptr
	// Skip two extra frames to account for this function
	// and runtime.Callers itself.
	n := runtime.Callers(2, pc[:])
	if n == 0 {
		panic("testutil: zero callers found")
	}
	var callsites []string
	frames := runtime.CallersFrames(pc[:n])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		callsites = append(callsites, filepath.Base(frame.File)+":"+strconv.Itoa(frame.Line))
	}
	buf := &strings.Builder{}
	for i := 1; i < len(callsites)-1; i++ {
		buf.WriteString(callsites[i] + ":")
	}
	return buf.String()
}

func Test_explodeSlice(t *testing.T) {
	type TT struct {
		dialect                 string
		slice                   interface{}
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		is := testutil.New(t, testutil.Parallel)
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, gotParams := []interface{}{}, map[string][]int{}
		is.True(isExplodableSlice(tt.slice))
		err := explodeSlice(tt.dialect, buf, &gotArgs, gotParams, tt.excludedTableQualifiers, tt.slice)
		is.NoErr(err)
		is.Equal(tt.wantQuery, buf.String())
		is.Equal(tt.wantArgs, gotArgs)
	}

	t.Run("tmpfield slice", func(t *testing.T) {
		var tt TT
		tt.slice = []tmpfield{{"111", "aaa"}, {"222", "bbb"}, {"333", "ccc"}}
		tt.excludedTableQualifiers = []string{"222"}
		tt.wantQuery = `"111".aaa, bbb, "333".ccc`
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("tmptable slice", func(t *testing.T) {
		var tt TT
		tt.slice = []tmptable{"111", "222", "333"}
		tt.excludedTableQualifiers = []string{"222"}
		tt.wantQuery = `"111", "222", "333"`
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})
}
