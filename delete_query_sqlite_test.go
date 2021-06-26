package sq_test

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	. "github.com/bokwoon95/sq"
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

func Test_SQLiteDeleteQuery(t *testing.T) {
	type TT struct {
		dialect   string
		item      Query
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := ToSQL(tt.dialect, tt.item)
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

	t.Run("delete", func(t *testing.T) {
		t.Parallel()
		var tt TT
		FILM1, FILM2, LANGUAGE := NEW_FILM("f1"), NEW_FILM("f2"), NEW_LANGUAGE("l")
		lang := NewCTE("lang", nil, SQLite.
			Select(LANGUAGE.LANGUAGE_ID, LANGUAGE.NAME).
			From(LANGUAGE).
			Where(LANGUAGE.NAME.IsNotNull()),
		)
		tt.item = SQLite.
			DeleteWith(lang).
			DeleteFrom(FILM1).
			Where(Exists(SQLite.
				SelectOne().
				From(FILM2).
				Join(lang, lang.Field("language_id").Eq(FILM2.LANGUAGE_ID)).
				Where(
					FILM1.FILM_ID.Eq(FILM2.FILM_ID),
					lang.Field("name").In([]string{"English", "Italian"}),
				),
			)).
			OrderBy(FILM1.FILM_ID).
			Limit(10)
		tt.wantQuery = "WITH lang AS (" +
			"SELECT l.language_id, l.name FROM language AS l WHERE l.name IS NOT NULL" +
			")" +
			" DELETE FROM film AS f1" +
			" WHERE EXISTS (" +
			"SELECT 1" +
			" FROM film AS f2 JOIN lang ON lang.language_id = f2.language_id" +
			" WHERE f1.film_id = f2.film_id AND lang.name IN ($1, $2)" +
			")" +
			" ORDER BY f1.film_id" +
			" LIMIT $3"
		tt.wantArgs = []interface{}{"English", "Italian", int64(10)}
		assert(t, tt)
	})
}
