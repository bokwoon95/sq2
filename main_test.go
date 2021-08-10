package sq_test

import (
	"database/sql"
	"embed"
	"flag"
	"io/fs"
	"log"
	"os"
	"testing"

	"github.com/bokwoon95/sq"
	"github.com/bokwoon95/sq/ddl"
	"github.com/bokwoon95/sq/internal/testutil"
)

//go:embed testdata
var embeddedFiles embed.FS

func TestMain(m *testing.M) {
	flag.Parse()
	sqliteSetup()
	postgresSetup()
	mysqlSetup()
	os.Exit(m.Run())
}

func sqliteSetup() {
	const dialect = sq.DialectSQLite
	if testing.Short() {
		return
	}
	db, err := sql.Open("sqlite3", "/Users/bokwoon/Documents/sq2/db.sqlite3")
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = ddl.AutoMigrate(dialect, tx, ddl.DropExtraneous|ddl.DropCascade)
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	err = ddl.AutoMigrate(dialect, tx, ddl.CreateMissing, ddl.WithTables(
		NEW_ACTOR(""),
		NEW_ADDRESS(""),
		NEW_CATEGORY(""),
		NEW_CITY(""),
		NEW_COUNTRY(""),
		NEW_CUSTOMER(""),
		NEW_FILM(""),
		NEW_FILM_ACTOR(""),
		NEW_FILM_ACTOR_REVIEW(""),
		NEW_FILM_CATEGORY(""),
		NEW_FILM_TEXT(""),
		NEW_INVENTORY(""),
		NEW_LANGUAGE(""),
		NEW_PAYMENT(""),
		NEW_RENTAL(""),
		NEW_STAFF(""),
		NEW_STORE(""),
	), ddl.WithDDLViews(
		NEW_ACTOR_INFO(""),
		NEW_CUSTOMER_LIST(""),
		NEW_FILM_LIST(""),
		NEW_FULL_ADDRESS(""),
		NEW_NICER_BUT_SLOWER_FILM_LIST(""),
		NEW_SALES_BY_FILM_CATEGORY(""),
		NEW_SALES_BY_STORE(""),
		NEW_STAFF_LIST(""),
	))
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	b, err := fs.ReadFile(embeddedFiles, "testdata/sqlite_sakila_data.sql")
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	_, err = tx.Exec(string(b))
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
}

func postgresSetup() {
	const dialect = sq.DialectPostgres
	if testing.Short() {
		return
	}
	extensions := []string{"btree_gist", "uuid-ossp"}
	functions, err := ddl.FilesToFunctions(sq.DialectPostgres, embeddedFiles,
		"testdata/postgres_last_update_trg.sql",
		"testdata/postgres_refresh_full_address.sql",
	)
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5442/db?sslmode=disable")
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = ddl.AutoMigrate(dialect, tx, ddl.DropExtraneous|ddl.DropCascade)
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	err = ddl.AutoMigrate(dialect, tx, ddl.CreateMissing, ddl.WithTables(
		NEW_ACTOR(""),
		NEW_ADDRESS(""),
		NEW_CATEGORY(""),
		NEW_CITY(""),
		NEW_COUNTRY(""),
		NEW_CUSTOMER(""),
		NEW_FILM(""),
		NEW_FILM_ACTOR(""),
		NEW_FILM_ACTOR_REVIEW(""),
		NEW_FILM_CATEGORY(""),
		NEW_FILM_TEXT(""),
		NEW_INVENTORY(""),
		NEW_LANGUAGE(""),
		NEW_PAYMENT(""),
		NEW_RENTAL(""),
		NEW_STAFF(""),
		NEW_STORE(""),
	), ddl.WithDDLViews(
		NEW_ACTOR_INFO(""),
		NEW_CUSTOMER_LIST(""),
		NEW_FILM_LIST(""),
		NEW_FULL_ADDRESS(""),
		NEW_NICER_BUT_SLOWER_FILM_LIST(""),
		NEW_SALES_BY_FILM_CATEGORY(""),
		NEW_SALES_BY_STORE(""),
		NEW_STAFF_LIST(""),
	), ddl.WithFunctions(
		functions...,
	), ddl.WithExtensions(
		extensions...,
	))
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	b, err := fs.ReadFile(embeddedFiles, "testdata/postgres_sakila_data.sql")
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	_, err = tx.Exec(string(b))
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
}

func mysqlSetup() {
	const dialect = sq.DialectMySQL
	if testing.Short() {
		return
	}
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3312)/db?parseTime=true&multiStatements=true")
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	err = ddl.AutoMigrate(dialect, db, ddl.DropExtraneous|ddl.DropCascade)
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	err = ddl.AutoMigrate(dialect, db, ddl.CreateMissing, ddl.WithTables(
		NEW_ACTOR(""),
		NEW_ADDRESS(""),
		NEW_CATEGORY(""),
		NEW_CITY(""),
		NEW_COUNTRY(""),
		NEW_CUSTOMER(""),
		NEW_FILM(""),
		NEW_FILM_ACTOR(""),
		NEW_FILM_ACTOR_REVIEW(""),
		NEW_FILM_CATEGORY(""),
		NEW_FILM_TEXT(""),
		NEW_INVENTORY(""),
		NEW_LANGUAGE(""),
		NEW_PAYMENT(""),
		NEW_RENTAL(""),
		NEW_STAFF(""),
		NEW_STORE(""),
	), ddl.WithDDLViews(
		NEW_ACTOR_INFO(""),
		NEW_CUSTOMER_LIST(""),
		NEW_FILM_LIST(""),
		NEW_FULL_ADDRESS(""),
		NEW_NICER_BUT_SLOWER_FILM_LIST(""),
		NEW_SALES_BY_FILM_CATEGORY(""),
		NEW_SALES_BY_STORE(""),
		NEW_STAFF_LIST(""),
	))
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	b, err := fs.ReadFile(embeddedFiles, "testdata/mysql_sakila_data.sql")
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
	_, err = db.Exec(string(b))
	if err != nil {
		log.Fatal(testutil.Callers(), err)
	}
}
