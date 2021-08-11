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
	"github.com/bokwoon95/sq/internal/testutil"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func Test_SakilaSQLite(t *testing.T) {
	const dialect = sq.DialectSQLite
	db, err := sql.Open("sqlite3", "/Users/bokwoon/Documents/sq2/db.sqlite3")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = AutoMigrate(dialect, tx, DropExtraneous|DropCascade)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	dbi, err := NewDatabaseIntrospector(dialect, tx, nil)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	tbls, err := dbi.GetTables(context.Background(), nil)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if len(tbls) != 0 {
		t.Fatal(testutil.Callers(), " AutoMigrate did not drop all tables:", fmt.Sprint(tbls))
	}
	views, err := dbi.GetViews(context.Background(), nil)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if len(views) != 0 {
		t.Fatal(testutil.Callers(), " AutoMigrate did not drop all views:", fmt.Sprint(views))
	}
	wantDBMetadata, err := NewDatabaseMetadata(dialect, WithTables(
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
		t.Fatal(testutil.Callers(), err)
	}
	upMigration, err := Migrate(CreateMissing|UpdateExisting, DatabaseMetadata{}, wantDBMetadata)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	err = upMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	gotUpSQL := buf.String()
	b, err := fs.ReadFile(embeddedFiles, "sql/sqlite_sakila_up.sql")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	wantUpSQL := strings.TrimSpace(string(b))
	if diff := testutil.Diff(gotUpSQL, wantUpSQL); diff != "" {
		t.Fatal(testutil.Callers(), diff)
	}
	err = upMigration.Exec(tx)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	gotDBMetadata, err := NewDatabaseMetadata(dialect, WithDB(tx, &Filter{SortOutput: true}))
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	introspectMigration, err := Migrate(CreateMissing|UpdateExisting, DatabaseMetadata{}, gotDBMetadata)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	buf.Reset()
	err = introspectMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	gotIntrospectSQL := buf.String()
	b, err = fs.ReadFile(embeddedFiles, "sql/sqlite_sakila_introspect.sql")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	wantIntrospectSQL := strings.TrimSpace(string(b))
	if diff := testutil.Diff(gotIntrospectSQL, wantIntrospectSQL); diff != "" {
		t.Fatal(testutil.Callers(), diff)
	}
	downMigration, err := Migrate(DropExtraneous|DropCascade, gotDBMetadata, DatabaseMetadata{})
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	buf.Reset()
	err = downMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	gotDownSQL := buf.String()
	b, err = fs.ReadFile(embeddedFiles, "sql/sqlite_sakila_down.sql")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	wantDownSQL := strings.TrimSpace(string(b))
	if diff := testutil.Diff(gotDownSQL, wantDownSQL); diff != "" {
		t.Fatal(testutil.Callers(), diff)
	}
}

func Test_SakilaPostgres(t *testing.T) {
	const dialect = sq.DialectPostgres
	extensions := []string{"btree_gist", "uuid-ossp"}
	functions, err := FilesToFunctions(sq.DialectPostgres, embeddedFiles,
		"sql/functions/postgres_last_update_trg.sql",
		"sql/functions/postgres_refresh_full_address.sql",
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5452/db?sslmode=disable")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = AutoMigrate(dialect, tx, DropExtraneous|DropCascade)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	dbi, err := NewDatabaseIntrospector(dialect, tx, nil)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	tbls, err := dbi.GetTables(context.Background(), nil)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if len(tbls) != 0 {
		t.Fatal(testutil.Callers(), " AutoMigrate did not drop all tables:", fmt.Sprint(tbls))
	}
	views, err := dbi.GetViews(context.Background(), nil)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if len(views) != 0 {
		t.Fatal(testutil.Callers(), " AutoMigrate did not drop all views:", fmt.Sprint(views))
	}
	funs, err := dbi.GetFunctions(context.Background(), nil)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if len(funs) != 0 {
		t.Fatal(testutil.Callers(), " AutoMigrate did not drop all functions:", fmt.Sprint(funs))
	}
	exts, err := dbi.GetExtensions(context.Background(), nil)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	for _, ext := range exts {
		if strings.HasPrefix(ext, "plpgsql") {
			continue
		}
		t.Fatal(testutil.Callers(), " AutoMigrate did not drop all extensions:", fmt.Sprint(exts))
	}
	wantDBMetadata, err := NewDatabaseMetadata(dialect, WithTables(
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
		t.Fatal(testutil.Callers(), err)
	}
	upMigration, err := Migrate(CreateMissing|UpdateExisting, DatabaseMetadata{}, wantDBMetadata)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	err = upMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	gotUpSQL := buf.String()
	b, err := fs.ReadFile(embeddedFiles, "sql/postgres_sakila_up.sql")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	wantUpSQL := strings.TrimSpace(string(b))
	if diff := testutil.Diff(gotUpSQL, wantUpSQL); diff != "" {
		t.Fatal(testutil.Callers(), diff)
	}
	err = upMigration.Exec(tx)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	gotDBMetadata, err := NewDatabaseMetadata(dialect, WithDB(tx, &Filter{SortOutput: true}))
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	// remove the version numbers
	for i, extension := range gotDBMetadata.Extensions {
		if n := strings.IndexByte(extension, '@'); n >= 0 {
			gotDBMetadata.Extensions[i] = extension[:n]
		}
	}
	introspectMigration, err := Migrate(CreateMissing|UpdateExisting, DatabaseMetadata{}, gotDBMetadata)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	buf.Reset()
	err = introspectMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	gotIntrospectSQL := buf.String()
	b, err = fs.ReadFile(embeddedFiles, "sql/postgres_sakila_introspect.sql")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	wantIntrospectSQL := strings.TrimSpace(string(b))
	if diff := testutil.Diff(gotIntrospectSQL, wantIntrospectSQL); diff != "" {
		t.Fatal(testutil.Callers(), diff)
	}
	downMigration, err := Migrate(DropExtraneous|DropCascade, gotDBMetadata, DatabaseMetadata{})
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	buf.Reset()
	err = downMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	gotDownSQL := buf.String()
	b, err = fs.ReadFile(embeddedFiles, "sql/postgres_sakila_down.sql")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	wantDownSQL := strings.TrimSpace(string(b))
	if diff := testutil.Diff(gotDownSQL, wantDownSQL); diff != "" {
		t.Fatal(testutil.Callers(), diff)
	}
}

func Test_SakilaMySQL(t *testing.T) {
	const dialect = sq.DialectMySQL
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3326)/db")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	err = AutoMigrate(dialect, db, DropExtraneous|DropCascade)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	dbi, err := NewDatabaseIntrospector(dialect, db, nil)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	tbls, err := dbi.GetTables(context.Background(), nil)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if len(tbls) != 0 {
		t.Fatal(testutil.Callers(), " AutoMigrate did not drop all tables:", fmt.Sprint(tbls))
	}
	views, err := dbi.GetViews(context.Background(), nil)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if len(views) != 0 {
		t.Fatal(testutil.Callers(), " AutoMigrate did not drop all views:", fmt.Sprint(views))
	}
	wantDBMetadata, err := NewDatabaseMetadata(dialect, WithTables(
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
		t.Fatal(testutil.Callers(), err)
	}
	upMigration, err := Migrate(CreateMissing|UpdateExisting, DatabaseMetadata{}, wantDBMetadata)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	err = upMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	gotUpSQL := buf.String()
	b, err := fs.ReadFile(embeddedFiles, "sql/mysql_sakila_up.sql")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	wantUpSQL := strings.TrimSpace(string(b))
	if diff := testutil.Diff(gotUpSQL, wantUpSQL); diff != "" {
		t.Fatal(testutil.Callers(), diff)
	}
	err = upMigration.Exec(db)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	gotDBMetadata, err := NewDatabaseMetadata(dialect, WithDB(db, &Filter{SortOutput: true}))
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	introspectMigration, err := Migrate(CreateMissing|UpdateExisting, DatabaseMetadata{}, gotDBMetadata)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	buf.Reset()
	err = introspectMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	gotIntrospectSQL := buf.String()
	b, err = fs.ReadFile(embeddedFiles, "sql/mysql_sakila_introspect.sql")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	wantIntrospectSQL := strings.TrimSpace(string(b))
	if diff := testutil.Diff(gotIntrospectSQL, wantIntrospectSQL); diff != "" {
		t.Fatal(testutil.Callers(), diff)
	}
	downMigration, err := Migrate(DropExtraneous|DropCascade, gotDBMetadata, DatabaseMetadata{})
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	buf.Reset()
	err = downMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	gotDownSQL := buf.String()
	b, err = fs.ReadFile(embeddedFiles, "sql/mysql_sakila_down.sql")
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	wantDownSQL := strings.TrimSpace(string(b))
	if diff := testutil.Diff(gotDownSQL, wantDownSQL); diff != "" {
		t.Fatal(testutil.Callers(), diff)
	}
}
