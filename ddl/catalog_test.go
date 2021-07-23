package ddl

import (
	"os"
	"testing"

	"github.com/bokwoon95/sq"
)

func Test_catalog_sqlite(t *testing.T) {
	const dialect = sq.DialectSQLite
	var functions []Function
	if dialect == sq.DialectPostgres {
		functions = append(functions, Function{
			SQL: `CREATE OR REPLACE FUNCTION last_update_trg() RETURNS trigger AS $$ BEGIN
	NEW.last_update = NOW();
	RETURN NEW;
END; $$ LANGUAGE plpgsql;`,
		})
	}
	catalog, err := NewCatalog(dialect, WithTables(
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
	), WithDDLViews(
		NEW_ACTOR_INFO(dialect, ""),
		NEW_CUSTOMER_LIST(dialect, ""),
		NEW_FILM_LIST(dialect, ""),
		NEW_NICER_BUT_SLOWER_FILM_LIST(dialect, ""),
		NEW_SALES_BY_FILM_CATEGORY(dialect, ""),
		NEW_SALES_BY_STORE(dialect, ""),
		NEW_STAFF_LIST(dialect, ""),
	), WithFunctions(functions...))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = catalog.Commands().WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	// b, err := json.MarshalIndent(catalog, "", "  ")
	// if err != nil {
	// 	t.Fatal(testcallers(), err)
	// }
	// os.Stdout.Write(b)
}

func Test_catalog_postgres(t *testing.T) {
	const dialect = sq.DialectPostgres
	var functions []Function
	if dialect == sq.DialectPostgres {
		functions = append(functions, Function{
			SQL: `CREATE OR REPLACE FUNCTION last_update_trg() RETURNS trigger AS $$ BEGIN
	NEW.last_update = NOW();
	RETURN NEW;
END; $$ LANGUAGE plpgsql;`,
		})
	}
	catalog, err := NewCatalog(dialect, WithTables(
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
	), WithDDLViews(
		NEW_ACTOR_INFO(dialect, ""),
		NEW_CUSTOMER_LIST(dialect, ""),
		NEW_FILM_LIST(dialect, ""),
		NEW_NICER_BUT_SLOWER_FILM_LIST(dialect, ""),
		NEW_SALES_BY_FILM_CATEGORY(dialect, ""),
		NEW_SALES_BY_STORE(dialect, ""),
		NEW_STAFF_LIST(dialect, ""),
	), WithFunctions(functions...))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = catalog.Commands().WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	// b, err := json.MarshalIndent(catalog, "", "  ")
	// if err != nil {
	// 	t.Fatal(testcallers(), err)
	// }
	// os.Stdout.Write(b)
}

func Test_catalog_mysql(t *testing.T) {
	const dialect = sq.DialectMySQL
	var functions []Function
	if dialect == sq.DialectPostgres {
		functions = append(functions, Function{
			SQL: `CREATE OR REPLACE FUNCTION last_update_trg() RETURNS trigger AS $$ BEGIN
	NEW.last_update = NOW();
	RETURN NEW;
END; $$ LANGUAGE plpgsql;`,
		})
	}
	catalog, err := NewCatalog(dialect, WithTables(
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
	), WithDDLViews(
		NEW_ACTOR_INFO(dialect, ""),
		NEW_CUSTOMER_LIST(dialect, ""),
		NEW_FILM_LIST(dialect, ""),
		NEW_NICER_BUT_SLOWER_FILM_LIST(dialect, ""),
		NEW_SALES_BY_FILM_CATEGORY(dialect, ""),
		NEW_SALES_BY_STORE(dialect, ""),
		NEW_STAFF_LIST(dialect, ""),
	), WithFunctions(functions...))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = catalog.Commands().WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	// b, err := json.MarshalIndent(catalog, "", "  ")
	// if err != nil {
	// 	t.Fatal(testcallers(), err)
	// }
	// os.Stdout.Write(b)
}
