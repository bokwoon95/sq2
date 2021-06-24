package ddl

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

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
	for i := 1; i < len(callsites)-1; i++ {
		buf.WriteString(callsites[i] + ":")
	}
	return buf.String()
}
