package sq

import (
	"context"
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/sync/errgroup"
)

//go:embed ddl/testdata
var embeddedFiles embed.FS

var (
	resetdbFlag     = flag.Bool("resetdb", false, "")
	sqliteDSNFlag   = flag.String("sqlite-dsn", "./db.sqlite3", "")
	postgresDSNFlag = flag.String("postgres-dsn", "postgres://postgres:postgres@localhost:5452/db?sslmode=disable&timezone=UTC", "")
	mysqlDSNFlag    = flag.String("mysql-dsn", "root:root@tcp(localhost:3326)/db?parseTime=true&time_zone=UTC&multiStatements=true", "")

	sqliteDB   *sql.DB
	postgresDB *sql.DB
	mysqlDB    *sql.DB
)

func TestMain(m *testing.M) {
	flag.Parse()
	initializeDBs()
	os.Exit(m.Run())
}

func runScript(ctx context.Context, db *sql.DB, fsys fs.FS, name string) error {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, string(b))
	return err
}

func initializeDBs() {
	if testing.Short() {
		return
	}
	dbinfos := []struct {
		driverName     string
		dataSourceName string
		downScript     string
		upScript       string
		dataScript     string
		tableQuery     string
		dataQuery      string
	}{
		{
			driverName:     "sqlite3",
			dataSourceName: *sqliteDSNFlag,
			downScript:     "ddl/testdata/sqlite_sakila_down.sql",
			upScript:       "ddl/testdata/sqlite_sakila_up.sql",
			dataScript:     "ddl/testdata/sqlite_sakila_data.sql",
			tableQuery:     "SELECT EXISTS(SELECT 1 FROM sqlite_schema WHERE tbl_name = 'actor')",
			dataQuery:      "SELECT EXISTS(SELECT 1 from actor)",
		},
		{
			driverName:     "postgres",
			dataSourceName: *postgresDSNFlag,
			downScript:     "ddl/testdata/postgres_sakila_down.sql",
			upScript:       "ddl/testdata/postgres_sakila_up.sql",
			dataScript:     "ddl/testdata/postgres_sakila_data.sql",
			tableQuery:     "SELECT EXISTS(SELECT 1 FROM pg_class WHERE relkind = 'r' AND relname = 'actor')",
			dataQuery:      "SELECT EXISTS(SELECT 1 from actor)",
		},
		{
			driverName:     "mysql",
			dataSourceName: *mysqlDSNFlag,
			downScript:     "ddl/testdata/mysql_sakila_down.sql",
			upScript:       "ddl/testdata/mysql_sakila_up.sql",
			dataScript:     "ddl/testdata/mysql_sakila_data.sql",
			tableQuery:     "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_name = 'actor')",
			dataQuery:      "SELECT EXISTS(SELECT 1 from actor)",
		},
	}
	g, gctx := errgroup.WithContext(context.Background())
	for _, dbinfo := range dbinfos {
		dbinfo := dbinfo
		g.Go(func() error {
			db, err := sql.Open(dbinfo.driverName, dbinfo.dataSourceName)
			if err != nil {
				return err
			}
			err = db.Ping()
			if err != nil {
				return fmt.Errorf("could not ping %s, is the database reachable? %w", dbinfo.dataSourceName, err)
			}
			switch dbinfo.driverName {
			case "sqlite3":
				sqliteDB = db
			case "postgres":
				postgresDB = db
			case "mysql":
				mysqlDB = db
			}
			if *resetdbFlag {
				fmt.Printf("[%8s] dropping tables\n", dbinfo.driverName)
				err = runScript(gctx, db, embeddedFiles, dbinfo.downScript)
				if err != nil {
					return err
				}
			}
			var tablesExist bool
			err = db.QueryRowContext(gctx, dbinfo.tableQuery).Scan(&tablesExist)
			if err != nil {
				return err
			}
			if !tablesExist {
				fmt.Printf("[%8s] creating tables\n", dbinfo.driverName)
				err = runScript(gctx, db, embeddedFiles, dbinfo.upScript)
				if err != nil {
					return err
				}
			}
			var dataExists bool
			err = db.QueryRowContext(gctx, dbinfo.dataQuery).Scan(&dataExists)
			if err != nil {
				return err
			}
			if !dataExists {
				fmt.Printf("[%8s] inserting data\n", dbinfo.driverName)
				err = runScript(gctx, db, embeddedFiles, dbinfo.dataScript)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Fatalln(err)
	}
}
