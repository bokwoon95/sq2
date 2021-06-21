package sq

import (
	"testing"

	"github.com/bokwoon95/testutil"
)

func TestVariadicQuery(t *testing.T) {
	type TT struct {
		item       toSQLer
		wantQuery  string
		wantArgs   []interface{}
		wantParams map[string][]int
	}

	const dialect = DialectMySQL

	assert := func(t *testing.T, tt TT) {
		is := testutil.New(t, testutil.Parallel)
		gotQuery, gotArgs, gotParams, err := tt.item.ToSQL()
		is.NoErr(err)
		is.Equal(tt.wantQuery, gotQuery)
		is.Equal(tt.wantArgs, gotArgs)
		if tt.wantParams != nil {
			is.Equal(tt.wantParams, gotParams)
		}
	}

	t.Run("empty", func(t *testing.T) {
		var tt TT
		tt.item = VariadicQuery{}
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
