package ddl

import (
	"fmt"
	"testing"

	"github.com/bokwoon95/sq"
	"github.com/bokwoon95/testutil"
)

func Test_LoadTable_MySQL(t *testing.T) {
	const dialect = sq.DialectMySQL
	is := testutil.New(t)
	tables := []sq.Table{
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
	}
	m := NewMetadata(dialect)
	for _, table := range tables {
		err := m.LoadTable(table)
		is.NoErr(err)
	}
	is.True(len(m.Schemas) > 0)
	for i, table := range m.Schemas[0].Tables {
		if i > 0 {
			fmt.Println()
		}
		str, err := CreateTable(dialect, table, IncludeConstraints|IncludeIndices)
		is.NoErr(err)
		fmt.Println(str)
	}
}
