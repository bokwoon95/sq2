package ddl

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
)

func testdiff(got, want interface{}) string {
	diff := cmp.Diff(got, want, cmp.Exporter(func(typ reflect.Type) bool { return true }))
	if diff != "" {
		return "\n-got +want\n" + diff
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
