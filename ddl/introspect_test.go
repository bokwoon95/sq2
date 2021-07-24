package ddl

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"

	"github.com/bokwoon95/sq"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func Test_introspect_postgres(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5442/db?sslmode=disable")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	catalog := Catalog{Dialect: sq.DialectPostgres}
	err = introspectPostgres(context.Background(), db, &catalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(catalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = catalog.Commands().WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
}

func Test_introspect_sqlite(t *testing.T) {
	db, err := sql.Open("sqlite3", "/Users/bokwoon/Documents/sq2/db.sqlite3")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	catalog := Catalog{Dialect: sq.DialectSQLite}
	err = introspectSQLite(context.Background(), db, &catalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(catalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = catalog.Commands().WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
}

func Test_introspect_mysql(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3312)/db")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	catalog := Catalog{Dialect: sq.DialectMySQL}
	err = introspectMySQL(context.Background(), db, &catalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(catalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = catalog.Commands().WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
}
