package ddl

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/bokwoon95/sq"
	_ "github.com/lib/pq"
)

func Test_introspect(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5442/db?sslmode=disable")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	catalog := Catalog{
		Dialect: sq.DialectPostgres,
	}
	err = introspectPostgres(context.Background(), db, &catalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = catalog.Commands().WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
}
