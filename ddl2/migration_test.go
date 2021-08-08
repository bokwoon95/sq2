package ddl2

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/bokwoon95/sq"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func Test_SakilaSQLite(t *testing.T) {
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
	err = AutoMigrate(dialect, tx, DropExtraneous|DropCascade)
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
	upMigration, err := Migrate(CreateMissing|UpdateExisting, Catalog{}, wantCatalog)
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
	b, err := fs.ReadFile(srcDir, "sql/sqlite_up.sql")
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
	introspectMigration, err := Migrate(CreateMissing|UpdateExisting, Catalog{}, gotCatalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf.Reset()
	err = introspectMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotIntrospectSQL := buf.String()
	b, err = fs.ReadFile(srcDir, "sql/sqlite_introspect.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantIntrospectSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotIntrospectSQL, wantIntrospectSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
	downMigration, err := Migrate(DropExtraneous|DropCascade, gotCatalog, Catalog{})
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf.Reset()
	err = downMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotDownSQL := buf.String()
	b, err = fs.ReadFile(srcDir, "sql/sqlite_down.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantDownSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotDownSQL, wantDownSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
}

func Test_SakilaPostgres(t *testing.T) {
	const dialect = sq.DialectPostgres
	extensions := []string{"btree_gist", "uuid-ossp"}
	functions, err := FilesToFunctions(sq.DialectPostgres, srcDir, "sql/last_update_trg.sql", "sql/refresh_full_address.sql")
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
	err = AutoMigrate(dialect, tx, DropExtraneous|DropCascade)
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
	if len(extensions) > 0 {
		query, args, _, err := sq.ToSQL(dialect, &DropExtensionCommand{
			DropIfExists: true,
			Extensions:   extensions,
			DropCascade:  true,
		})
		if err != nil {
			t.Fatal(testcallers(), err)
		}
		_, err = tx.Exec(query, args...)
		if err != nil {
			t.Fatal(testcallers(), err)
		}
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
	upMigration, err := Migrate(CreateMissing|UpdateExisting, Catalog{}, wantCatalog)
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
	b, err := fs.ReadFile(srcDir, "sql/postgres_up.sql")
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
	for i, extension := range gotCatalog.Extensions {
		if n := strings.IndexByte(extension, '@'); n >= 0 {
			gotCatalog.Extensions[i] = extension[:n]
		}
	}
	introspectMigration, err := Migrate(CreateMissing|UpdateExisting, Catalog{}, gotCatalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf.Reset()
	err = introspectMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotIntrospectSQL := buf.String()
	b, err = fs.ReadFile(srcDir, "sql/postgres_introspect.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantIntrospectSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotIntrospectSQL, wantIntrospectSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
	downMigration, err := Migrate(DropExtraneous|DropCascade, gotCatalog, Catalog{})
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	buf.Reset()
	err = downMigration.WriteSQL(buf)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotDownSQL := buf.String()
	b, err = fs.ReadFile(srcDir, "sql/postgres_down.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantDownSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotDownSQL, wantDownSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
}

func Test_SakilaMySQL(t *testing.T) {
	const dialect = sq.DialectMySQL
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3312)/db")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = AutoMigrate(dialect, db, DropExtraneous|DropCascade)
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
	upMigration, err := Migrate(CreateMissing|UpdateExisting, Catalog{}, wantCatalog)
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
	b, err := fs.ReadFile(srcDir, "sql/mysql_up.sql")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	wantUpSQL := strings.TrimSpace(string(b))
	if diff := testdiff(gotUpSQL, wantUpSQL); diff != "" {
		t.Fatal(testcallers(), diff)
	}
	// err = upMigration.Exec(db)
	// if err != nil {
	// 	t.Fatal(testcallers(), err)
	// }
	// gotCatalog, err := NewCatalog(dialect, WithDB(db, &Filter{SortOutput: true}))
	// if err != nil {
	// 	t.Fatal(testcallers(), err)
	// }
	// introspectMigration, err := Migrate(CreateMissing|UpdateExisting, Catalog{}, gotCatalog)
	// if err != nil {
	// 	t.Fatal(testcallers(), err)
	// }
	// buf.Reset()
	// err = introspectMigration.WriteSQL(buf)
	// if err != nil {
	// 	t.Fatal(testcallers(), err)
	// }
	// gotIntrospectSQL := buf.String()
	// b, err = fs.ReadFile(dataDir, "sql/mysql_introspect.sql")
	// if err != nil {
	// 	t.Fatal(testcallers(), err)
	// }
	// wantIntrospectSQL := strings.TrimSpace(string(b))
	// if diff := testdiff(gotIntrospectSQL, wantIntrospectSQL); diff != "" {
	// 	t.Fatal(testcallers(), diff)
	// }
	// downMigration, err := Migrate(DropExtraneous|DropCascade, gotCatalog, Catalog{})
	// if err != nil {
	// 	t.Fatal(testcallers(), err)
	// }
	// buf.Reset()
	// err = downMigration.WriteSQL(buf)
	// if err != nil {
	// 	t.Fatal(testcallers(), err)
	// }
	// gotDownSQL := buf.String()
	// b, err = fs.ReadFile(dataDir, "sql/mysql_down.sql")
	// if err != nil {
	// 	t.Fatal(testcallers(), err)
	// }
	// wantDownSQL := strings.TrimSpace(string(b))
	// if diff := testdiff(gotDownSQL, wantDownSQL); diff != "" {
	// 	t.Fatal(testcallers(), diff)
	// }
}

func Test_DropMySQL(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3312)/db")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	gotCatalog, err := NewCatalog(sq.DialectMySQL, WithDB(db, &Filter{SortOutput: true}))
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
		NEW_FULL_ADDRESS(""),
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

// func Test_ResetMySQL(t *testing.T) {
// 	db, err := sql.Open("mysql", "root:root@tcp(localhost:3312)/db")
// 	if err != nil {
// 		t.Fatal(testcallers(), err)
// 	}
// 	err = AutoMigrate(sq.DialectMySQL, db, DropExtraneous|DropCascade)
// 	if err != nil {
// 		t.Fatal(testcallers(), err)
// 	}
// }
//
// func Test_SetupMySQL(t *testing.T) {
// 	db, err := sql.Open("mysql", "root:root@tcp(localhost:3312)/db")
// 	if err != nil {
// 		t.Fatal(testcallers(), err)
// 	}
// 	err = AutoMigrate(sq.DialectMySQL, db, CreateMissing|UpdateExisting, WithTables(
// 		NEW_ACTOR(""),
// 		NEW_CATEGORY(""),
// 		NEW_COUNTRY(""),
// 		NEW_CITY(""),
// 		NEW_ADDRESS(""),
// 		NEW_LANGUAGE(""),
// 		NEW_FILM(""),
// 		NEW_FILM_TEXT(""),
// 		NEW_FILM_ACTOR(""),
// 		NEW_FILM_ACTOR_REVIEW(""),
// 		NEW_FILM_CATEGORY(""),
// 		NEW_STAFF(""),
// 		NEW_STORE(""),
// 		NEW_CUSTOMER(""),
// 		NEW_INVENTORY(""),
// 		NEW_RENTAL(""),
// 		NEW_PAYMENT(""),
// 	), WithDDLViews(
// 		NEW_ACTOR_INFO(""),
// 		NEW_CUSTOMER_LIST(""),
// 		NEW_FILM_LIST(""),
// 		NEW_NICER_BUT_SLOWER_FILM_LIST(""),
// 		NEW_SALES_BY_FILM_CATEGORY(""),
// 		NEW_SALES_BY_STORE(""),
// 		NEW_FULL_ADDRESS(""),
// 	))
// 	if err != nil {
// 		t.Fatal(testcallers(), err)
// 	}
// }

func Test_IntrospectMySQL(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3312)/db")
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	catalog, err := NewCatalog(sq.DialectMySQL, WithDB(db, &Filter{SortOutput: true}))
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	migration, err := Migrate(CreateMissing|UpdateExisting, Catalog{}, catalog)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
	err = migration.WriteSQL(os.Stdout)
	if err != nil {
		t.Fatal(testcallers(), err)
	}
}
