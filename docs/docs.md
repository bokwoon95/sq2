# What is `sq`?

`sq` is a type-safe query builder and data mapper for Go. It supports the **SQLite**, **Postgres** and **MySQL** dialects. Among its features are:

- **Type-safety**
    - Every table gets a struct type.
    - Every column gets a struct field.
    - You no longer have to hardcode tables and columns as raw strings (even [ORMs](https://gorm.io/docs/query.html#Conditions) are guilty of this).
- **Bidirectional Schema definition**
    - Code-generate table structs from your database (database-first).
    - Generate DDL from your table structs (code-first).
        - DDL generation is idempotent, missing tables and columns are added as needed.
    - What is supported: schemas, tables, columns, constraints, indexes, (materialized) views, triggers, functions, extensions, enums
- **Uses Go generics for data fetching**
    - Data mapping is built on [callback mapper functions](#)
    - [FetchOne/FetchSlice](#) are generic fetch functions that return whatever the callback mapper function returns
- **Faithful emulation of each SQL dialect**
    - Each dialect has its own query builder that can leverage dialect-specific syntax.
    - This does not mean that queries are not portable: queries are as portable as the SQL that you write.
    - `sq` comes with its own [tricks](#) to help with writing queries that can target multiple dialects.
- **Application-side Row Level Security (i.e. multitenancy support)**
    - Query-level variables can be added by passing in a `map[string]interface{}`.
    - Based on these variables, tables participating in a query that implement the [PredicateInjector](#) interface can inject additional predicates to exclude rows from a SELECT, UPDATE or DELETE.
    - This emulates Postgres' [Row Level Security feature](#), but more importantly it plays well with `database/sql`'s connection pooling since variables are set per-query and not per-session.
    - SQLite and MySQL get Row Level Security for free.
- **And many more**
    - Multiple schemas, Generated Columns, Full Text Search, JSON, Collations.

# Quickstart Example

```go
package main

import (
    "database/sql"
    "log"
    "time"

    "github.com/bokwoon95/sq"
    "github.com/bokwoon95/sq/ddl"
    _ "github.com/mattn/go-sqlite3"
)

type ACTOR struct {
    sq.TableInfo
    ACTOR_ID    sq.NumberField `ddl:"type=INTEGER primarykey"`
    FIRST_NAME  sq.StringField `ddl:"notnull"`
    LAST_NAME   sq.StringField `ddl:"notnull"`
    LAST_UPDATE sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP"`
}

type Actor struct {
    ActorID    int
    FirstName  string
    LastName   string
    LastUpdate time.Time
}

var (
    now    = time.Now()
    actors = []Actor{
        {ActorID: 1, FirstName: "PENELOPE", LastName: "GUINESS", LastUpdate: now},
        {ActorID: 2, FirstName: "NICK", LastName: "WAHLBERG", LastUpdate: now},
        {ActorID: 3, FirstName: "ED", LastName: "CHASE", LastUpdate: now},
        {ActorID: 4, FirstName: "JENNIFER", LastName: "DAVIS", LastUpdate: now},
    }
)

func main() {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        log.Fatalln(err)
    }
    ACTOR := ACTOR{}
    _ = sq.ReflectTable(&ACTOR, "")
    err = ddl.AutoMigrate(sq.DialectSQLite, db, ddl.CreateMissing|ddl.UpdateExisting,
        ddl.WithTables(ACTOR),
    )
    if err != nil {
        log.Fatalln(err)
    }

    // INSERT
    rowsAffected, _, err := sq.Exec(sq.Log(db), sq.SQLite.
        InsertInto(ACTOR).
        Valuesx(func(col *sq.Column) error {
            for _, actor := range actors {
                col.SetInt(ACTOR.ACTOR_ID, actor.ActorID)
                col.SetString(ACTOR.FIRST_NAME, actor.FirstName)
                col.SetString(ACTOR.LAST_NAME, actor.LastName)
                col.SetTime(ACTOR.LAST_UPDATE, actor.LastUpdate)
            }
            return nil
        }),
    )
    if err != nil {
        log.Fatalln(err)
    }
    log.Printf("INSERT: %d rows inserted\n", rowsAffected)

    // SELECT (uses FetchOne)
    var penelope Actor
    _, err = sq.Fetch(sq.Log(db), sq.SQLite.
        From(ACTOR).
        Where(
            ACTOR.FIRST_NAME.EqString("PENELOPE"),
            ACTOR.LAST_NAME.EqString("GUINESS"),
        ),
        func(row *sq.Row) {
            penelope.ActorID = row.Int(ACTOR.ACTOR_ID)
            penelope.FirstName = row.String(ACTOR.FIRST_NAME)
            penelope.LastName = row.String(ACTOR.LAST_NAME)
            penelope.LastUpdate = row.Time(ACTOR.LAST_UPDATE)
        },
    )
    if err != nil {
        log.Fatalln(err)
    }
    log.Printf("%+v\n", penelope)

    // UPDATE
    _, _, err = sq.Exec(sq.Log(db), sq.SQLite.
        Update(ACTOR).
        Setx(func(col *sq.Column) error {
            col.SetString(ACTOR.FIRST_NAME, "Penelope")
            col.SetString(ACTOR.LAST_NAME, "Guiness")
            return nil
        }).
        Where(ACTOR.ACTOR_ID.EqInt(penelope.ActorID)),
    )
    if err != nil {
        log.Fatalln(err)
    }

    // DELETE
    _, _, err = sq.Exec(sq.Log(db), sq.SQLite.
        DeleteFrom(ACTOR).
        Where(
            ACTOR.FIRST_NAME.EqString("ED"),
            ACTOR.LAST_NAME.EqString("CHASE"),
        ),
        0,
    )
    if err != nil {
        log.Fatalln(err)
    }

    // print table contents (uses FetchSlice)
    var dbActors []Actor
    _, err = sq.Fetch(sq.Log(db), sq.SQLite.From(ACTOR).OrderBy(ACTOR.ACTOR_ID), func(row *sq.Row) {
        actor := Actor{
            ActorID:    row.Int(ACTOR.ACTOR_ID),
            FirstName:  row.String(ACTOR.FIRST_NAME),
            LastName:   row.String(ACTOR.LAST_NAME),
            LastUpdate: row.Time(ACTOR.LAST_UPDATE),
        }
        row.Process(func() { dbActors = append(dbActors, actor) })
    })
    log.Printf("db actors: %#v\n", dbActors)
}
```

# DDL

## How do I generate table structs?

```bash
go install github.com/bokwoon95/sq/cmd/sq-ddl
DATABASE_URL='postgres://username:password@localhost:5432/db?sslmode=disable'
sq-ddl $DATABASE_URL -format=ddl
sq-ddl $DATABASE_URL -format=json
sq-ddl $DATABASE_URL -format=structs
sq-ddl $DATABASE_URL \
    -format=structs \
    -outfile=tables.go \
    -pkg=tables \
    -with-schemas=public,main \
    -without-tables=schema_migrations \
    -overwrite
```
