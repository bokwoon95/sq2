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
