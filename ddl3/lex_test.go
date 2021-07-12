package ddl3

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func testdiff(lhs, rhs interface{}) string {
	diff := cmp.Diff(lhs, rhs, cmp.Exporter(func(typ reflect.Type) bool { return true }))
	if diff != "" {
		return "\n-lhs +rhs\n" + diff
	}
	return ""
}

func testcallers() string {
	/* https://talks.godoc.org/github.com/davecheney/go-1.9-release-party/presentation.slide#20
	 * "Users of runtime.Callers should avoid directly inspecting the resulting PC
	 * slice and instead use runtime.CallersFrames to get a complete view of the
	 * call stack, or runtime.Caller to get information about a single caller.
	 * This is because an individual element of the PC slice cannot account for
	 * inlined frames or other nuances of the call stack."
	 */
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

func Test_lexModifiers(t *testing.T) {
	type TT struct {
		config            string
		wantModifiers     [][2]string
		wantModifierIndex map[string]int
	}

	assert := func(t *testing.T, tt TT) {
		gotModifiers, gotModifierIndex, err := lexModifiers(tt.config)
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
		gotValue, gotModifiers, gotModifierIndex, err := lexValue(tt.config)
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
