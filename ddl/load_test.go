package ddl

import (
	"bytes"
	"testing"

	"github.com/bokwoon95/sq"
)

func makeDDL(t *testing.T, dialect string, table sq.Table) (ddl string) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	m := NewMetadata(dialect)
	err := m.LoadTable(table)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	tbl := m.Schemas[0].Tables[0]
	str, err := CreateTable(dialect, tbl)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf.WriteString(str)
	if dialect != sq.DialectSQLite {
		for _, constraint := range tbl.Constraints {
			if constraint.ConstraintType != FOREIGN_KEY {
				continue
			}
			buf.WriteString("\n")
			str, err := CreateConstraint(dialect, constraint)
			if err != nil {
				t.Fatal(testcallers(), err)
			}
			buf.WriteString(str)
		}
	}
	for _, index := range tbl.Indices {
		buf.WriteString("\n")
		str, err := CreateIndex(dialect, index)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		buf.WriteString(str)
	}
	return buf.String()
}

func Test_LoadTable(t *testing.T) {
	t.Run("ACTOR SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_ACTOR(sq.DialectSQLite, ""))
		if diff := testdiff(ACTOR_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("ACTOR Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_ACTOR(sq.DialectPostgres, ""))
		if diff := testdiff(ACTOR_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("ACTOR MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_ACTOR(sq.DialectMySQL, ""))
		if diff := testdiff(ACTOR_MySQL, gotDDL); diff != "" {
		}
		// fmt.Println(gotDDL)
	})

	t.Run("CATEGORY SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_CATEGORY(sq.DialectSQLite, ""))
		if diff := testdiff(CATEGORY_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("CATEGORY Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_CATEGORY(sq.DialectPostgres, ""))
		if diff := testdiff(CATEGORY_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("CATEGORY MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_CATEGORY(sq.DialectMySQL, ""))
		if diff := testdiff(CATEGORY_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("COUNTRY SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_COUNTRY(sq.DialectSQLite, ""))
		if diff := testdiff(COUNTRY_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("COUNTRY Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_COUNTRY(sq.DialectPostgres, ""))
		if diff := testdiff(COUNTRY_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("COUNTRY MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_COUNTRY(sq.DialectMySQL, ""))
		if diff := testdiff(COUNTRY_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("CITY SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_CITY(sq.DialectSQLite, ""))
		if diff := testdiff(CITY_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("CITY Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_CITY(sq.DialectPostgres, ""))
		if diff := testdiff(CITY_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("CITY MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_CITY(sq.DialectMySQL, ""))
		if diff := testdiff(CITY_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("ADDRESS SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_ADDRESS(sq.DialectSQLite, ""))
		if diff := testdiff(ADDRESS_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("ADDRESS Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_ADDRESS(sq.DialectPostgres, ""))
		if diff := testdiff(ADDRESS_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("ADDRESS MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_ADDRESS(sq.DialectMySQL, ""))
		if diff := testdiff(ADDRESS_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("LANGUAGE SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_LANGUAGE(sq.DialectSQLite, ""))
		if diff := testdiff(LANGUAGE_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("LANGUAGE Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_LANGUAGE(sq.DialectPostgres, ""))
		if diff := testdiff(LANGUAGE_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("LANGUAGE MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_LANGUAGE(sq.DialectMySQL, ""))
		if diff := testdiff(LANGUAGE_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("FILM SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_FILM(sq.DialectSQLite, ""))
		if diff := testdiff(FILM_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("FILM Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_FILM(sq.DialectPostgres, ""))
		if diff := testdiff(FILM_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("FILM MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_FILM(sq.DialectMySQL, ""))
		if diff := testdiff(FILM_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("FILM_TEXT SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_FILM_TEXT(sq.DialectSQLite, ""))
		if diff := testdiff(FILM_TEXT_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("FILM_TEXT MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_FILM_TEXT(sq.DialectMySQL, ""))
		if diff := testdiff(FILM_TEXT_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("FILM_ACTOR SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_FILM_ACTOR(sq.DialectSQLite, ""))
		if diff := testdiff(FILM_ACTOR_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("FILM_ACTOR Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_FILM_ACTOR(sq.DialectPostgres, ""))
		if diff := testdiff(FILM_ACTOR_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("FILM_ACTOR MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_FILM_ACTOR(sq.DialectMySQL, ""))
		if diff := testdiff(FILM_ACTOR_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("FILM_CATEGORY SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_FILM_CATEGORY(sq.DialectSQLite, ""))
		if diff := testdiff(FILM_CATEGORY_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("FILM_CATEGORY Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_FILM_CATEGORY(sq.DialectPostgres, ""))
		if diff := testdiff(FILM_CATEGORY_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("FILM_CATEGORY MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_FILM_CATEGORY(sq.DialectMySQL, ""))
		if diff := testdiff(FILM_CATEGORY_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("STAFF SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_STAFF(sq.DialectSQLite, ""))
		if diff := testdiff(STAFF_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("STAFF Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_STAFF(sq.DialectPostgres, ""))
		if diff := testdiff(STAFF_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("STAFF MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_STAFF(sq.DialectMySQL, ""))
		if diff := testdiff(STAFF_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("STORE SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_STORE(sq.DialectSQLite, ""))
		if diff := testdiff(STORE_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("STORE Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_STORE(sq.DialectPostgres, ""))
		if diff := testdiff(STORE_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("STORE MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_STORE(sq.DialectMySQL, ""))
		if diff := testdiff(STORE_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("CUSTOMER SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_CUSTOMER(sq.DialectSQLite, ""))
		if diff := testdiff(CUSTOMER_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("CUSTOMER Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_CUSTOMER(sq.DialectPostgres, ""))
		if diff := testdiff(CUSTOMER_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("CUSTOMER MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_CUSTOMER(sq.DialectMySQL, ""))
		if diff := testdiff(CUSTOMER_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("INVENTORY SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_INVENTORY(sq.DialectSQLite, ""))
		if diff := testdiff(INVENTORY_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("INVENTORY Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_INVENTORY(sq.DialectPostgres, ""))
		if diff := testdiff(INVENTORY_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("INVENTORY MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_INVENTORY(sq.DialectMySQL, ""))
		if diff := testdiff(INVENTORY_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("RENTAL SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_RENTAL(sq.DialectSQLite, ""))
		if diff := testdiff(RENTAL_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("RENTAL Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_RENTAL(sq.DialectPostgres, ""))
		if diff := testdiff(RENTAL_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("RENTAL MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_RENTAL(sq.DialectMySQL, ""))
		if diff := testdiff(RENTAL_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("PAYMENT SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_PAYMENT(sq.DialectSQLite, ""))
		if diff := testdiff(PAYMENT_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("PAYMENT Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_PAYMENT(sq.DialectPostgres, ""))
		if diff := testdiff(PAYMENT_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("PAYMENT MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_PAYMENT(sq.DialectMySQL, ""))
		if diff := testdiff(PAYMENT_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("DUMMY_TABLE SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_DUMMY_TABLE(sq.DialectSQLite, ""))
		if diff := testdiff(DUMMY_TABLE_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("DUMMY_TABLE Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_DUMMY_TABLE(sq.DialectPostgres, ""))
		if diff := testdiff(DUMMY_TABLE_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("DUMMY_TABLE MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_DUMMY_TABLE(sq.DialectMySQL, ""))
		if diff := testdiff(DUMMY_TABLE_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})

	t.Run("DUMMY_TABLE_2 SQLite", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectSQLite, NEW_DUMMY_TABLE_2(sq.DialectSQLite, ""))
		if diff := testdiff(DUMMY_TABLE_2_SQLite, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("DUMMY_TABLE_2 Postgres", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectPostgres, NEW_DUMMY_TABLE_2(sq.DialectPostgres, ""))
		if diff := testdiff(DUMMY_TABLE_2_Postgres, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
	t.Run("DUMMY_TABLE_2 MySQL", func(t *testing.T) {
		t.Parallel()
		gotDDL := makeDDL(t, sq.DialectMySQL, NEW_DUMMY_TABLE_2(sq.DialectMySQL, ""))
		if diff := testdiff(DUMMY_TABLE_2_MySQL, gotDDL); diff != "" {
			t.Error(testcallers(), diff)
		}
		// fmt.Println(gotDDL)
	})
}
