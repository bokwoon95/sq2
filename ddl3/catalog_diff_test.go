package ddl3

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/bokwoon95/sq"
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
	buf.WriteString("]")
	return buf.String()
}

func Test_Z(t *testing.T) {
	const dialect = sq.DialectSQLite
	wantCatalog, err := NewCatalog(dialect, WithTables(
		NEW_ACTOR(dialect, ""),
		NEW_CATEGORY(dialect, ""),
		NEW_COUNTRY(dialect, ""),
		NEW_CITY(dialect, ""),
		NEW_ADDRESS(dialect, ""),
		NEW_LANGUAGE(dialect, ""),
		NEW_FILM(dialect, ""),
		NEW_FILM_TEXT(dialect, ""),
		NEW_FILM_ACTOR(dialect, ""),
		NEW_FILM_CATEGORY(dialect, ""),
		NEW_STAFF(dialect, ""),
		NEW_STORE(dialect, ""),
		NEW_CUSTOMER(dialect, ""),
		NEW_INVENTORY(dialect, ""),
		NEW_RENTAL(dialect, ""),
		NEW_PAYMENT(dialect, ""),
		NEW_DUMMY_TABLE(dialect, ""),
		NEW_DUMMY_TABLE_2(dialect, ""),
	))
	if err != nil {
		t.Fatal(Callers(), err)
	}
	catalogDiff, err := DiffCatalog(Catalog{Dialect: dialect}, wantCatalog)
	if err != nil {
		t.Fatal(Callers(), err)
	}
	cmdset := catalogDiff.Commands()
	err = cmdset.WriteOut(os.Stdout)
	if err != nil {
		t.Fatal(Callers(), err)
	}
}
