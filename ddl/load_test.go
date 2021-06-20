package ddl

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/bokwoon95/sq"
	"github.com/bokwoon95/testutil"
)

func makeDDL(is testutil.I, dialect string, table sq.Table) (ddl string) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	m := NewMetadata(dialect)
	err := m.LoadTable(table)
	is.NoErr(err)
	tbl := m.Schemas[0].Tables[0]
	str, err := CreateTable(dialect, tbl)
	is.NoErr(err)
	buf.WriteString(str)
	if dialect != sq.DialectSQLite {
		for _, constraint := range tbl.Constraints {
			buf.WriteString("\n")
			str, err := CreateConstraint(dialect, constraint)
			is.NoErr(err)
			buf.WriteString(str)
		}
	}
	for _, index := range tbl.Indices {
		buf.WriteString("\n")
		str, err := CreateIndex(dialect, index)
		is.NoErr(err)
		buf.WriteString(str)
	}
	return buf.String()
}

func Test_LoadTable(t *testing.T) {
	t.Run("ACTOR SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_ACTOR(sq.DialectSQLite, ""))
		is.Equal(ACTOR_SQLite, gotDDL)
	})
	t.Run("ACTOR Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_ACTOR(sq.DialectPostgres, ""))
		is.Equal(ACTOR_Postgres, gotDDL)
	})
	t.Run("ACTOR MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_ACTOR(sq.DialectMySQL, ""))
		is.Equal(ACTOR_MySQL, gotDDL)
	})

	t.Run("CATEGORY SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_CATEGORY(sq.DialectSQLite, ""))
		fmt.Println(gotDDL)
	})
	t.Run("CATEGORY Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_CATEGORY(sq.DialectPostgres, ""))
		fmt.Println(gotDDL)
	})
	t.Run("CATEGORY MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_CATEGORY(sq.DialectMySQL, ""))
		fmt.Println(gotDDL)
	})
}

// NEW_ACTOR(dialect, ""),
// NEW_CATEGORY(dialect, ""),
// NEW_COUNTRY(dialect, ""),
// NEW_CITY(dialect, ""),
// NEW_ADDRESS(dialect, ""),
// NEW_LANGUAGE(dialect, ""),
// NEW_FILM(dialect, ""),
// NEW_FILM_TEXT(dialect, ""),
// NEW_FILM_ACTOR(dialect, ""),
// NEW_FILM_CATEGORY(dialect, ""),
// NEW_STAFF(dialect, ""),
// NEW_STORE(dialect, ""),
// NEW_CUSTOMER(dialect, ""),
// NEW_INVENTORY(dialect, ""),
// NEW_RENTAL(dialect, ""),
// NEW_PAYMENT(dialect, ""),
// NEW_DUMMY_TABLE(dialect, ""),
// NEW_DUMMY_TABLE_2(dialect, ""),
