package ddl

import (
	"fmt"
	"testing"

	"github.com/bokwoon95/testutil"
)

func Test_LoadTable(t *testing.T) {
	is := testutil.New(t)
	const dialect = "postgres"
	ACTOR := NEW_ACTOR(dialect, "")
	m := NewMetadata(dialect)
	err := m.LoadTable(ACTOR)
	is.NoErr(err)
	is.True(len(m.Schemas) > 0 && len(m.Schemas[0].Tables) > 0)
	wantTable := ACTOR_TABLE(dialect)
	gotTable := m.Schemas[0].Tables[0]
	is.Equal(wantTable, gotTable)
	str, err := gotTable.ToSQL(dialect)
	is.NoErr(err)
	fmt.Println(str)
}
