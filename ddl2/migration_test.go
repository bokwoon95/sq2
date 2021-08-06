package ddl2

import (
	"database/sql"
	"os"
	"testing"

	"github.com/bokwoon95/sq"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func Test_DropViewSQLite(t *testing.T) {
	db, err := sql.Open("sqlite3", "/Users/bokwoon/Documents/sq2/db.sqlite3")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotCatalog, err := NewCatalog(sq.DialectSQLite, WithDB(db))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	migration, err := Migrate(DropExtraneous|DropCascade, gotCatalog, Catalog{})
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = migration.WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
}

func Test_DropViewPostgres(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5442/db?sslmode=disable")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotCatalog, err := NewCatalog(sq.DialectPostgres, WithDB(db))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	migration, err := Migrate(DropExtraneous|DropCascade, gotCatalog, Catalog{})
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = migration.WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
}

func Test_DropViewMySQL(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3312)/db")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotCatalog, err := NewCatalog(sq.DialectMySQL, WithDB(db))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	migration, err := Migrate(DropExtraneous|DropCascade, gotCatalog, Catalog{})
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = migration.WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
}

func Test_CatalogSQLite(t *testing.T) {
	wantCatalog, err := NewCatalog(sq.DialectSQLite, WithTables(
		NEW_ACTOR(""),
		NEW_CATEGORY(""),
		NEW_COUNTRY(""),
		NEW_CITY(""),
		NEW_ADDRESS(""),
		NEW_LANGUAGE(""),
		NEW_FILM(""),
		NEW_FILM_TEXT(""),
		NEW_FILM_ACTOR(""),
		NEW_FILM_ACTOR_REVIEW(""),
		NEW_FILM_CATEGORY(""),
		NEW_STAFF(""),
		NEW_STORE(""),
		NEW_CUSTOMER(""),
		NEW_INVENTORY(""),
		NEW_RENTAL(""),
		NEW_PAYMENT(""),
	), WithDDLViews(
		NEW_ACTOR_INFO(""),
		NEW_CUSTOMER_LIST(""),
		NEW_FILM_LIST(""),
		NEW_NICER_BUT_SLOWER_FILM_LIST(""),
		NEW_SALES_BY_FILM_CATEGORY(""),
		NEW_SALES_BY_STORE(""),
		NEW_STAFF_LIST(""),
	))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	migration, err := Migrate(CreateMissing, Catalog{}, wantCatalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = migration.WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
}

func Test_CatalogPostgres(t *testing.T) {
	functions, err := FilesToFunctions(sq.DialectPostgres, sqlDir, "sql/last_update_trg.sql", "sql/refresh_full_address.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantCatalog, err := NewCatalog(sq.DialectPostgres, WithTables(
		NEW_ACTOR(""),
		NEW_CATEGORY(""),
		NEW_COUNTRY(""),
		NEW_CITY(""),
		NEW_ADDRESS(""),
		NEW_LANGUAGE(""),
		NEW_FILM(""),
		NEW_FILM_TEXT(""),
		NEW_FILM_ACTOR(""),
		NEW_FILM_ACTOR_REVIEW(""),
		NEW_FILM_CATEGORY(""),
		NEW_STAFF(""),
		NEW_STORE(""),
		NEW_CUSTOMER(""),
		NEW_INVENTORY(""),
		NEW_RENTAL(""),
		NEW_PAYMENT(""),
	), WithDDLViews(
		NEW_ACTOR_INFO(""),
		NEW_CUSTOMER_LIST(""),
		NEW_FILM_LIST(""),
		NEW_NICER_BUT_SLOWER_FILM_LIST(""),
		NEW_SALES_BY_FILM_CATEGORY(""),
		NEW_SALES_BY_STORE(""),
		NEW_STAFF_LIST(""),
		NEW_FULL_ADDRESS(""),
	), WithFunctions(functions...))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	migration, err := Migrate(CreateMissing, Catalog{}, wantCatalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = migration.WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
}

func Test_CatalogMySQL(t *testing.T) {
	wantCatalog, err := NewCatalog(sq.DialectMySQL, WithTables(
		NEW_ACTOR(""),
		NEW_CATEGORY(""),
		NEW_COUNTRY(""),
		NEW_CITY(""),
		NEW_ADDRESS(""),
		NEW_LANGUAGE(""),
		NEW_FILM(""),
		NEW_FILM_TEXT(""),
		NEW_FILM_ACTOR(""),
		NEW_FILM_ACTOR_REVIEW(""),
		NEW_FILM_CATEGORY(""),
		NEW_STAFF(""),
		NEW_STORE(""),
		NEW_CUSTOMER(""),
		NEW_INVENTORY(""),
		NEW_RENTAL(""),
		NEW_PAYMENT(""),
	), WithDDLViews(
		NEW_ACTOR_INFO(""),
		NEW_CUSTOMER_LIST(""),
		NEW_FILM_LIST(""),
		NEW_NICER_BUT_SLOWER_FILM_LIST(""),
		NEW_SALES_BY_FILM_CATEGORY(""),
		NEW_SALES_BY_STORE(""),
		NEW_STAFF_LIST(""),
	))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	migration, err := Migrate(CreateMissing, Catalog{}, wantCatalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = migration.WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
}
