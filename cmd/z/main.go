package main

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// pgx
	pgxConfig, err := pgx.ParseConfig("postgres://postgres:postgres@localhost:5452/db?sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}
	pgxDB, err := pgx.ConnectConfig(context.Background(), pgxConfig)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = pgxDB.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS tmppgx ( id INT, a TEXT, b TIMESTAMPTZ );
	ALTER TABLE IF EXISTS tmppgx DROP COLUMN IF EXISTS a;
	ALTER TABLE IF EXISTS tmppgx DROP COLUMN IF EXISTS b;
	DROP TABLE tmppgx;
	`)
	if err != nil {
		log.Fatalln(err)
	}
	// pq
	pqDB, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5452/db?sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = pqDB.Exec(`
	CREATE TABLE IF NOT EXISTS tmppq ( id INT, a TEXT, b TIMESTAMPTZ );
	ALTER TABLE IF EXISTS tmppq DROP COLUMN IF EXISTS a;
	ALTER TABLE IF EXISTS tmppq DROP COLUMN IF EXISTS b;
	DROP TABLE tmppq;
	`)
	// mysql
	mysqlDB, err := sql.Open("mysql", "root:root@tcp(localhost:3326)/db?multiStatements=true")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = mysqlDB.Exec(`
	CREATE TABLE IF NOT EXISTS tmp ( id INT, a TEXT, b TIMESTAMPTZ );
	ALTER TABLE tmp DROP COLUMN a;
	ALTER TABLE tmp DROP COLUMN b;
	DROP TABLE tmp;
	`)
	// sqlite
	sqliteDB, err := sql.Open("sqlite3", "./db.sqlite3")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = sqliteDB.Exec(`
	CREATE TABLE IF NOT EXISTS tmp ( id INT, a TEXT, b TIMESTAMPTZ );
	ALTER TABLE tmp DROP COLUMN a;
	ALTER TABLE tmp DROP COLUMN b;
	DROP TABLE tmp;
	`)
	if err != nil {
		log.Fatalln(err)
	}
	// _, err = db.Exec(`CREATE FUNCTION test (text) RETURNS text AS $$ BEGIN RETURN '5'; END; $$ LANGUAGE plpgsql;`)
	// _, err = db.Exec(`CREATE FUNCTION test (t text) RETURNS text DETERMINISTIC BEGIN RETURN '5'; END;`)
}
