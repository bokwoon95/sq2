package testutil

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
)

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
	buf.WriteString("] ")
	return buf.String()
}
