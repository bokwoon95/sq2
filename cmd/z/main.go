package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5440/db?sslmode=disable")
	db, err := sql.Open("mysql", "root:root@tcp(localhost:3310)/db")
	if err != nil {
		log.Fatalln(err)
	}
	// _, err = db.Exec(`CREATE FUNCTION test (text) RETURNS text AS $$ BEGIN RETURN '5'; END; $$ LANGUAGE plpgsql;`)
	_, err = db.Exec(`CREATE FUNCTION test (t text) RETURNS text DETERMINISTIC BEGIN RETURN '5'; END;`)
	if err != nil {
		log.Fatalln(err)
	}
}
