package ddl

import (
	"testing"

	"github.com/bokwoon95/testutil"
)

func Test_lexModifiers(t *testing.T) {
	type TT struct {
		config        string
		wantModifiers [][2]string
		wantNameIndex int
		wantColsIndex int
	}

	assert := func(t *testing.T, tt TT) {
		is := testutil.New(t, testutil.Parallel)
		gotModifiers, gotNameIndex, gotColsIndex, err := lexModifiers(tt.config)
		is.NoErr(err)
		is.Equal(tt.wantModifiers, gotModifiers)
		is.Equal(tt.wantNameIndex, gotNameIndex)
		is.Equal(tt.wantColsIndex, gotColsIndex)
	}

	t.Run("empty", func(t *testing.T) {
		var tt TT
		tt.config = ""
		tt.wantNameIndex, tt.wantColsIndex = -1, -1
		assert(t, tt)
	})

	t.Run("test1", func(t *testing.T) {
		var tt TT
		tt.config = "notnull unique index={. unique} name=testing references={inventory onupdate=cascade ondelete=restrict}"
		tt.wantModifiers = [][2]string{
			{"notnull", ""},
			{"unique", ""},
			{"index", ". unique"},
			{"name", "testing"},
			{"references", "inventory onupdate=cascade ondelete=restrict"},
		}
		tt.wantNameIndex, tt.wantColsIndex = 3, -1
		assert(t, tt)
	})

	t.Run("test2", func(t *testing.T) {
		var tt TT
		tt.config = "cols=a,b,c index={. where={email LIKE '%gmail'}}"
		tt.wantModifiers = [][2]string{
			{"cols", "a,b,c"},
			{"index", ". where={email LIKE '%gmail'}"},
		}
		tt.wantNameIndex, tt.wantColsIndex = -1, 0
		assert(t, tt)
	})
}

func Test_lexValue(t *testing.T) {
	type TT struct {
		config        string
		wantValue     string
		wantModifiers [][2]string
		wantNameIndex int
		wantColsIndex int
	}

	assert := func(t *testing.T, tt TT) {
		is := testutil.New(t, testutil.Parallel)
		gotValue, gotModifiers, gotNameIndex, gotColsIndex, err := lexValue(tt.config)
		is.NoErr(err)
		is.Equal(tt.wantValue, gotValue)
		is.Equal(tt.wantModifiers, gotModifiers)
		is.Equal(tt.wantNameIndex, gotNameIndex)
		is.Equal(tt.wantColsIndex, gotColsIndex)
	}

	t.Run("", func(t *testing.T) {
		var tt TT
		tt.config = ""
		tt.wantNameIndex, tt.wantColsIndex = -1, -1
		assert(t, tt)
	})

	t.Run("", func(t *testing.T) {
		var tt TT
		tt.config = "1 unique"
		tt.wantValue = "1"
		tt.wantModifiers = [][2]string{
			{"unique", ""},
		}
		tt.wantNameIndex, tt.wantColsIndex = -1, -1
		assert(t, tt)
	})

	t.Run("", func(t *testing.T) {
		var tt TT
		tt.config = "{abcd efg} generated={first_name || ' ' || last_name} virtual name=gg=G"
		tt.wantValue = "abcd efg"
		tt.wantModifiers = [][2]string{
			{"generated", "first_name || ' ' || last_name"},
			{"virtual", ""},
			{"name", "gg=G"},
		}
		tt.wantNameIndex, tt.wantColsIndex = 2, -1
		assert(t, tt)
	})

	t.Run("", func(t *testing.T) {
		var tt TT
		tt.config = "inventory cols=1,2,3,4 onupdate=cascade ondelete=restrict"
		tt.wantValue = "inventory"
		tt.wantModifiers = [][2]string{
			{"cols", "1,2,3,4"},
			{"onupdate", "cascade"},
			{"ondelete", "restrict"},
		}
		tt.wantNameIndex, tt.wantColsIndex = -1, 0
		assert(t, tt)
	})
}
