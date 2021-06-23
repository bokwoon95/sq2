package sq

import (
	"bytes"
	"testing"

	"github.com/bokwoon95/testutil"
)

func TestVariadicQuery(t *testing.T) {
	type TT struct {
		dialect    string
		item       SQLAppender
		wantQuery  string
		wantArgs   []interface{}
		wantParams map[string][]int
	}

	const dialect = DialectMySQL

	assert := func(t *testing.T, tt TT) {
		is := testutil.New(t, testutil.Parallel)
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, gotParams := []interface{}{}, map[string][]int{}
		err := tt.item.AppendSQL(tt.dialect, buf, &gotArgs, gotParams)
		is.NoErr(err)
		is.Equal(tt.wantQuery, buf.String())
		is.Equal(tt.wantArgs, gotArgs)
		if tt.wantParams != nil {
			is.Equal(tt.wantParams, gotParams)
		}
	}

	t.Run("empty", func(t *testing.T) {
		var tt TT
		tt.item = VariadicQuery{}
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("nested single variadic queries", func(t *testing.T) {
		var tt TT
		tt.item = Union(UnionAll(Intersect(IntersectAll(Except(ExceptAll(Queryf(dialect, "SELECT {}", 1)))))))
		tt.wantQuery = "SELECT ?"
		tt.wantArgs = []interface{}{1}
		assert(t, tt)
	})

	t.Run("nested variadic queries", func(t *testing.T) {
		var tt TT
		tt.item = Intersect(
			Union(Union(Queryf(dialect, "SELECT {}", 1)), Queryf(dialect, "SELECT {}", "abc")),
			Union(Queryf(dialect, "SELECT {}", 3.14)),
		)
		tt.wantQuery = "(SELECT ? UNION SELECT ?) INTERSECT SELECT ?"
		tt.wantArgs = []interface{}{1, "abc", 3.14}
		assert(t, tt)
	})
}
