package ddl

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/bokwoon95/sq"
)

func Test_Sakila2SQLite(t *testing.T) {
	const dialect = sq.DialectSQLite
	db, err := sql.Open("sqlite3", "/Users/bokwoon/Documents/sq2/db.sqlite3")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = AutoMigrate2(dialect, tx, DropExtraneous|DropCascade)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	dbi, err := NewDatabaseIntrospector(dialect, tx, nil)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	tbls, err := dbi.GetTables(context.Background(), nil)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	if len(tbls) != 0 {
		t.Fatal(testcallers(), " AutoMigrate did not drop all tables:", fmt.Sprint(tbls))
	}
	views, err := dbi.GetViews(context.Background(), nil)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	if len(views) != 0 {
		t.Fatal(testcallers(), " AutoMigrate did not drop all views:", fmt.Sprint(views))
	}
	wantCatalog, err := NewCatalog(dialect, WithTables(
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
	), WithDDLViews(
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
		t.Fatal(testcallers(), err)
	}
	upMigration, err := Migrate2(CreateMissing|UpdateExisting, Catalog{}, wantCatalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	err = upMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotUpSQL := buf.String()
	b, err := fs.ReadFile(embeddedFiles, "sql/sqlite_sakila_up.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantUpSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotUpSQL, wantUpSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
	err = upMigration.Exec(tx)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotCatalog, err := NewCatalog(dialect, WithDB(tx, &Filter{SortOutput: true}))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	introspectMigration, err := Migrate2(CreateMissing|UpdateExisting, Catalog{}, gotCatalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf.Reset()
	err = introspectMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotIntrospectSQL := buf.String()
	b, err = fs.ReadFile(embeddedFiles, "sql/sqlite_sakila_introspect.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantIntrospectSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotIntrospectSQL, wantIntrospectSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
	downMigration, err := Migrate2(DropExtraneous|DropCascade, gotCatalog, Catalog{})
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf.Reset()
	err = downMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotDownSQL := buf.String()
	b, err = fs.ReadFile(embeddedFiles, "sql/sqlite_sakila_down.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantDownSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotDownSQL, wantDownSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
}

func Test_Sakila2Postgres(t *testing.T) {
	const dialect = sq.DialectPostgres
	extensions := []string{"btree_gist", "uuid-ossp"}
	functions, err := FilesToFunctions(sq.DialectPostgres, embeddedFiles,
		"sql/functions/postgres_last_update_trg.sql",
		"sql/functions/postgres_refresh_full_address.sql",
	)
	_, _ = extensions, functions
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5442/db?sslmode=disable")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = AutoMigrate2(dialect, tx, DropExtraneous|DropCascade)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	dbi, err := NewDatabaseIntrospector(dialect, tx, nil)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	tbls, err := dbi.GetTables(context.Background(), nil)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	if len(tbls) != 0 {
		t.Fatal(testcallers(), " AutoMigrate did not drop all tables:", fmt.Sprint(tbls))
	}
	views, err := dbi.GetViews(context.Background(), nil)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	if len(views) != 0 {
		t.Fatal(testcallers(), " AutoMigrate did not drop all views:", fmt.Sprint(views))
	}
	funs, err := dbi.GetFunctions(context.Background(), nil)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	if len(funs) != 0 {
		t.Fatal(testcallers(), " AutoMigrate did not drop all functions:", fmt.Sprint(funs))
	}
	exts, err := dbi.GetExtensions(context.Background(), nil)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	for _, ext := range exts {
		if strings.HasPrefix(ext, "plpgsql") {
			continue
		}
		t.Fatal(testcallers(), " AutoMigrate did not drop all extensions:", fmt.Sprint(exts))
	}
	wantCatalog, err := NewCatalog(dialect, WithTables(
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
	), WithDDLViews(
		NEW_ACTOR_INFO(""),
		NEW_CUSTOMER_LIST(""),
		NEW_FILM_LIST(""),
		NEW_FULL_ADDRESS(""),
		NEW_NICER_BUT_SLOWER_FILM_LIST(""),
		NEW_SALES_BY_FILM_CATEGORY(""),
		NEW_SALES_BY_STORE(""),
		NEW_STAFF_LIST(""),
	), WithFunctions(
		functions...,
	), WithExtensions(
		extensions...,
	))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	upMigration, err := Migrate2(CreateMissing|UpdateExisting, Catalog{}, wantCatalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	err = upMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotUpSQL := buf.String()
	b, err := fs.ReadFile(embeddedFiles, "sql/postgres_sakila_up2.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantUpSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotUpSQL, wantUpSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
	err = upMigration.Exec(tx)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotCatalog, err := NewCatalog(dialect, WithDB(tx, &Filter{SortOutput: true}))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	// remove the version numbers
	for i, extension := range gotCatalog.Extensions {
		if n := strings.IndexByte(extension, '@'); n >= 0 {
			gotCatalog.Extensions[i] = extension[:n]
		}
	}
	introspectMigration, err := Migrate2(CreateMissing|UpdateExisting, Catalog{}, gotCatalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf.Reset()
	err = introspectMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotIntrospectSQL := buf.String()
	b, err = fs.ReadFile(embeddedFiles, "sql/postgres_sakila_introspect2.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantIntrospectSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotIntrospectSQL, wantIntrospectSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
	downMigration, err := Migrate2(DropExtraneous|DropCascade, gotCatalog, Catalog{})
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf.Reset()
	err = downMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotDownSQL := buf.String()
	b, err = fs.ReadFile(embeddedFiles, "sql/postgres_sakila_down.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantDownSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotDownSQL, wantDownSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
}

func Test_Sakila2MySQL(t *testing.T) {
	const dialect = sq.DialectMySQL
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3312)/db")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = AutoMigrate2(dialect, db, DropExtraneous|DropCascade)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	dbi, err := NewDatabaseIntrospector(dialect, db, nil)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	tbls, err := dbi.GetTables(context.Background(), nil)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	if len(tbls) != 0 {
		t.Fatal(testcallers(), " AutoMigrate did not drop all tables:", fmt.Sprint(tbls))
	}
	views, err := dbi.GetViews(context.Background(), nil)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	if len(views) != 0 {
		t.Fatal(testcallers(), " AutoMigrate did not drop all views:", fmt.Sprint(views))
	}
	wantCatalog, err := NewCatalog(dialect, WithTables(
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
	), WithDDLViews(
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
		t.Fatal(testcallers(), err)
	}
	upMigration, err := Migrate2(CreateMissing|UpdateExisting, Catalog{}, wantCatalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	err = upMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotUpSQL := buf.String()
	b, err := fs.ReadFile(embeddedFiles, "sql/mysql_sakila_up.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantUpSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotUpSQL, wantUpSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
	err = upMigration.Exec(db)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotCatalog, err := NewCatalog(dialect, WithDB(db, &Filter{SortOutput: true}))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	introspectMigration, err := Migrate2(CreateMissing|UpdateExisting, Catalog{}, gotCatalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf.Reset()
	err = introspectMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotIntrospectSQL := buf.String()
	b, err = fs.ReadFile(embeddedFiles, "sql/mysql_sakila_introspect.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantIntrospectSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotIntrospectSQL, wantIntrospectSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
	downMigration, err := Migrate2(DropExtraneous|DropCascade, gotCatalog, Catalog{})
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf.Reset()
	err = downMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotDownSQL := buf.String()
	b, err = fs.ReadFile(embeddedFiles, "sql/mysql_sakila_down.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantDownSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotDownSQL, wantDownSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
}
