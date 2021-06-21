package ddl

import (
	"bytes"
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
			if constraint.ConstraintType != FOREIGN_KEY {
				continue
			}
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
		// fmt.Println(gotDDL)
	})
	t.Run("ACTOR Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_ACTOR(sq.DialectPostgres, ""))
		is.Equal(ACTOR_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("ACTOR MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_ACTOR(sq.DialectMySQL, ""))
		is.Equal(ACTOR_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("CATEGORY SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_CATEGORY(sq.DialectSQLite, ""))
		is.Equal(CATEGORY_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("CATEGORY Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_CATEGORY(sq.DialectPostgres, ""))
		is.Equal(CATEGORY_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("CATEGORY MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_CATEGORY(sq.DialectMySQL, ""))
		is.Equal(CATEGORY_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("COUNTRY SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_COUNTRY(sq.DialectSQLite, ""))
		is.Equal(COUNTRY_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("COUNTRY Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_COUNTRY(sq.DialectPostgres, ""))
		is.Equal(COUNTRY_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("COUNTRY MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_COUNTRY(sq.DialectMySQL, ""))
		is.Equal(COUNTRY_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("CITY SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_CITY(sq.DialectSQLite, ""))
		is.Equal(CITY_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("CITY Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_CITY(sq.DialectPostgres, ""))
		is.Equal(CITY_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("CITY MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_CITY(sq.DialectMySQL, ""))
		is.Equal(CITY_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("ADDRESS SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_ADDRESS(sq.DialectSQLite, ""))
		is.Equal(ADDRESS_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("ADDRESS Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_ADDRESS(sq.DialectPostgres, ""))
		is.Equal(ADDRESS_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("ADDRESS MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_ADDRESS(sq.DialectMySQL, ""))
		is.Equal(ADDRESS_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("LANGUAGE SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_LANGUAGE(sq.DialectSQLite, ""))
		is.Equal(LANGUAGE_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("LANGUAGE Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_LANGUAGE(sq.DialectPostgres, ""))
		is.Equal(LANGUAGE_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("LANGUAGE MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_LANGUAGE(sq.DialectMySQL, ""))
		is.Equal(LANGUAGE_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("FILM SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_FILM(sq.DialectSQLite, ""))
		is.Equal(FILM_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("FILM Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_FILM(sq.DialectPostgres, ""))
		is.Equal(FILM_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("FILM MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_FILM(sq.DialectMySQL, ""))
		is.Equal(FILM_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("FILM_TEXT SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_FILM_TEXT(sq.DialectSQLite, ""))
		is.Equal(FILM_TEXT_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("FILM_TEXT MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_FILM_TEXT(sq.DialectMySQL, ""))
		is.Equal(FILM_TEXT_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("FILM_ACTOR SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_FILM_ACTOR(sq.DialectSQLite, ""))
		is.Equal(FILM_ACTOR_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("FILM_ACTOR Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_FILM_ACTOR(sq.DialectPostgres, ""))
		is.Equal(FILM_ACTOR_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("FILM_ACTOR MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_FILM_ACTOR(sq.DialectMySQL, ""))
		is.Equal(FILM_ACTOR_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("FILM_CATEGORY SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_FILM_CATEGORY(sq.DialectSQLite, ""))
		is.Equal(FILM_CATEGORY_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("FILM_CATEGORY Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_FILM_CATEGORY(sq.DialectPostgres, ""))
		is.Equal(FILM_CATEGORY_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("FILM_CATEGORY MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_FILM_CATEGORY(sq.DialectMySQL, ""))
		is.Equal(FILM_CATEGORY_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("STAFF SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_STAFF(sq.DialectSQLite, ""))
		is.Equal(STAFF_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("STAFF Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_STAFF(sq.DialectPostgres, ""))
		is.Equal(STAFF_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("STAFF MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_STAFF(sq.DialectMySQL, ""))
		is.Equal(STAFF_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("STORE SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_STORE(sq.DialectSQLite, ""))
		is.Equal(STORE_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("STORE Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_STORE(sq.DialectPostgres, ""))
		is.Equal(STORE_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("STORE MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_STORE(sq.DialectMySQL, ""))
		is.Equal(STORE_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("CUSTOMER SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_CUSTOMER(sq.DialectSQLite, ""))
		is.Equal(CUSTOMER_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("CUSTOMER Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_CUSTOMER(sq.DialectPostgres, ""))
		is.Equal(CUSTOMER_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("CUSTOMER MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_CUSTOMER(sq.DialectMySQL, ""))
		is.Equal(CUSTOMER_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("INVENTORY SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_INVENTORY(sq.DialectSQLite, ""))
		is.Equal(INVENTORY_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("INVENTORY Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_INVENTORY(sq.DialectPostgres, ""))
		is.Equal(INVENTORY_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("INVENTORY MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_INVENTORY(sq.DialectMySQL, ""))
		is.Equal(INVENTORY_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("RENTAL SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_RENTAL(sq.DialectSQLite, ""))
		is.Equal(RENTAL_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("RENTAL Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_RENTAL(sq.DialectPostgres, ""))
		is.Equal(RENTAL_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("RENTAL MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_RENTAL(sq.DialectMySQL, ""))
		is.Equal(RENTAL_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("PAYMENT SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_PAYMENT(sq.DialectSQLite, ""))
		is.Equal(PAYMENT_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("PAYMENT Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_PAYMENT(sq.DialectPostgres, ""))
		is.Equal(PAYMENT_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("PAYMENT MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_PAYMENT(sq.DialectMySQL, ""))
		is.Equal(PAYMENT_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("DUMMY_TABLE SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_DUMMY_TABLE(sq.DialectSQLite, ""))
		is.Equal(DUMMY_TABLE_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("DUMMY_TABLE Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_DUMMY_TABLE(sq.DialectPostgres, ""))
		is.Equal(DUMMY_TABLE_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("DUMMY_TABLE MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_DUMMY_TABLE(sq.DialectMySQL, ""))
		is.Equal(DUMMY_TABLE_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})

	t.Run("DUMMY_TABLE_2 SQLite", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectSQLite, NEW_DUMMY_TABLE_2(sq.DialectSQLite, ""))
		is.Equal(DUMMY_TABLE_2_SQLite, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("DUMMY_TABLE_2 Postgres", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectPostgres, NEW_DUMMY_TABLE_2(sq.DialectPostgres, ""))
		is.Equal(DUMMY_TABLE_2_Postgres, gotDDL)
		// fmt.Println(gotDDL)
	})
	t.Run("DUMMY_TABLE_2 MySQL", func(t *testing.T) {
		is := testutil.New(t, testutil.Parallel)
		gotDDL := makeDDL(is, sq.DialectMySQL, NEW_DUMMY_TABLE_2(sq.DialectMySQL, ""))
		is.Equal(DUMMY_TABLE_2_MySQL, gotDDL)
		// fmt.Println(gotDDL)
	})
}
