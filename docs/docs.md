# What is `sq`?

`sq` is a type-safe query builder and data mapper for Go. It supports the **SQLite**, **Postgres** and **MySQL** dialects. Among its features are:

- **Type-safety**
    - Every table gets a struct type.
    - Every column gets a struct field.
    - You no longer have to hardcode tables and columns as raw strings (even [ORMs](https://gorm.io/docs/query.html#Conditions) are guilty of this).
- **Bidirectional schema definition**
    - Code-generate table structs from your database (database-first).
    - Generate SQL commands (DDL) from your table structs (code-first).
        - DDL generation is idempotent, missing tables and columns are added as needed.
    - What is supported: schemas, tables, columns, constraints, indexes, (materialized) views, triggers, functions, extensions, enums
- **Uses Go generics for data fetching**
    - Data mapping is built on [callback mapper functions](#) which can return any type.
    - Whatever type a mapper functions returns, it is automatically returned by the generic functions [FetchOne and FetchSlice](#).
- **Faithful emulation of each SQL dialect**
    - Each dialect has its own query builder that can leverage dialect-specific syntax.
    - This does not mean that queries are not portable: queries are as portable as the SQL that you write.
    - `sq` comes with its own [tricks](#) to help with writing queries that can target multiple dialects.
- **Application-side Row Level Security (i.e. multitenancy support)**
    - Variables can be associated with a query via a `map[string]interface{}`.
    - Based on those variables, tables implementing the [PredicateInjector](#) interface can inject additional predicates into a query whenever they are invoked.
    - This emulates Postgres' [Row Level Security](#), but without needing to mess around with `current_user` or session-level variables.
    - Since it's implemented application-side, MySQL and SQLite can use Row Level Security too.
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

// actor table
type ACTOR struct {
    sq.TableInfo
    ACTOR_ID    sq.NumberField `ddl:"type=INTEGER primarykey"`
    FIRST_NAME  sq.StringField `ddl:"notnull"`
    LAST_NAME   sq.StringField `ddl:"notnull"`
    LAST_UPDATE sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP"`
}

// actor type
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
    // open database
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        log.Fatalln(err)
    }

    // initialize database
    ACTOR := ACTOR{}
    _ = sq.ReflectTable(&ACTOR, "")
    err = ddl.AutoMigrate(sq.DialectSQLite, db, ddl.CreateMissing|ddl.UpdateExisting,
        ddl.WithTables(ACTOR),
    )
    if err != nil {
        log.Fatalln(err)
    }

    // INSERT actors
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
    log.Printf("%d rows inserted\n", rowsAffected)

    // SELECT actor 'PENELOPE GUINESS' (uses FetchOne)
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
    log.Printf("penelope: %+v\n", penelope)

    // UPDATE actor 'PENELOPE GUINESS' to 'Penelope Guiness'
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

    // DELETE actor 'ED CHASE'
    rowsAffected, _, err = sq.Exec(sq.Log(db), sq.SQLite.
        DeleteFrom(ACTOR).
        Where(
            ACTOR.FIRST_NAME.EqString("ED"),
            ACTOR.LAST_NAME.EqString("CHASE"),
        ),
    )
    if err != nil {
        log.Fatalln(err)
    }
    log.Printf("%d row deleted\n", rowsAffected)

    // SELECT all actors, ordered by actor_id (uses FetchSlice)
    var allActors []Actor
    _, err = sq.Fetch(sq.Log(db), sq.SQLite.
        From(ACTOR).
        OrderBy(ACTOR.ACTOR_ID),
        func(row *sq.Row) {
            actor := Actor{
                ActorID:    row.Int(ACTOR.ACTOR_ID),
                FirstName:  row.String(ACTOR.FIRST_NAME),
                LastName:   row.String(ACTOR.LAST_NAME),
                LastUpdate: row.Time(ACTOR.LAST_UPDATE),
            }
            row.Process(func() { allActors = append(allActors, actor) })
        },
    )
    log.Printf("actors: %+v\n", allActors)
}
```

# DDL

## How do I generate table structs?

```bash
go install github.com/bokwoon95/sq/cmd/sq-ddl
DATABASE_URL='postgres://username:password@localhost:5432/db?sslmode=disable'
# driver dialect can be inferred if it starts with postgres:// or mysql://. else use sqlite
sq-ddl -dsn=$DATABASE_URL
sq-ddl -dsn=$DATABASE_URL -output=ddl -outfile=schema.sql
sq-ddl -dsn=$DATABASE_URL -output=json -outfile=db.json
sq-ddl -dsn=$DATABASE_URL -output=structs -outfile=tables.go -pkg=tables -overwrite
sq-ddl -dsn=$DATABASE_URL -with-schemas=main,public,db -without-tables=schema_migrations
# -db-driver -db-user -db-pass -db-port -db-host -db-name
```

## How do I define tables in code?

Mention how the flow is you define a struct that embeds sq.TableInfo, then how each field should correspond to a column in the table.

[Types]

Each field has types: bring up the type mapping table (as well as the backup `CustomField` for anything else).

[Naming]

Each table's name is determined by setting the Name field in the embedded sq.TableInfo struct.
Each field's name is determined when calling the Field constructor (list the field constructor examples).
Since setting the table and field names manually each time will be tedious, `sq` provides a helper function New() and NewAliased() which instantiates a table struct via reflection.
They will first check for an `sq.name` struct annotation, followed by lowercasing the field name (for columns) or struct name (for tables).

each field's name is determined upon setting
each field's name will be implicitly mapped by simply uppercasing or lowercasing depending on the mapping direction.
mention how the user can override this behaviour by adding the sq.name struct annotation, but it would only be picked up by sq.New/sq.NewAliased

## How do I handle migrations?
