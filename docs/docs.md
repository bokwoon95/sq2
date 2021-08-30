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
sq-ddl -dsn=$DATABASE_URL -output=structs,constructors -outfile=tables.go -pkg=tables -overwrite
sq-ddl -dsn=$DATABASE_URL -with-schemas=main,public,db -without-tables=schema_migrations
# -db-driver -db-user -db-pass -db-port -db-host -db-name
```

## How do I define tables in code?

A table struct is a struct that embeds an `sq.TableInfo` followed by one or more `Field`s. The embedded `sq.TableInfo` must be the very first field in the struct.

The utility of defining table structs by hand instead of code generating them is that they can then be used to automatically generate (and apply) DDL via the [ddl.Migrate/ddl.AutoMigrate](#) functions. When used that way, the structs become the source of truth for your table schema.

An example:

```go
type ACTOR struct {
    sq.TableInfo
    ACTOR_ID    sq.NumberField
    FIRST_NAME  sq.StringField
    LAST_NAME   sq.StringField
    LAST_UPDATE sq.TimeField
}
ACTOR := sq.New[ACTOR]()
err := ddl.AutoMigrate("sqlite", db, sq.CreateMissing, sq.WithTables(ACTOR))
```

which corresponds to

```sql
CREATE TABLE IF NOT EXISTS actor (
    actor_id INT
    ,first_name TEXT
    ,last_name TEXT
    ,last_update DATETIME
);
```

### Instantiating table structs with `sq.New`
Prior to usage, tables structs must be instantiated using `sq.New`. That fills in the table and column names in the struct via reflection. If you wish to avoid the reflection overhead, you can use a custom constructor function instead (either handwritten or generated from the database with [`sq-ddl`](#)).

`sq.NewAliased` performs the same thing as `sq.New`, except it also sets the table alias.

```go
type ACTOR struct {
    sq.TableInfo
    ACTOR_ID    sq.NumberField
    FIRST_NAME  sq.StringField
    LAST_NAME   sq.StringField
    LAST_UPDATE sq.TimeField
}

ACTOR := sq.New[ACTOR]()       // FROM actor

a := sq.NewAliased[ACTOR]("a") // FROM actor AS a

// Custom constructor function
func NEW_ACTOR() ACTOR {
    tableInfo := sq.TableInfo{TableName: "actor"}
    return ACTOR{
        TableInfo:   tableInfo,
        ACTOR_ID:    sq.NewNumberField("actor_id", tableInfo),
        FIRST_NAME:  sq.NewStringField("first_name", tableInfo)
        LAST_NAME:   sq.NewStringField("last_name", tableInfo),
        LAST_UPDATE: sq.NewTimeField("last_update", tableInfo),
    }
}
ACTOR := NEW_ACTOR()     // FROM actor
ACTOR.ACTOR_ID           // actor.actor_id
ACTOR.FIRST_NAME         // actor.first_name
ACTOR.LAST_NAME          // actor.last_name
ACTOR.LAST_UPDATE        // actor.last_update
```

By default, the default name assigned by `sq.New` is simply the lowercased struct name/field name. This naming default can be overridden with an `sq.name` struct annotation for the corresponding table or field. Note that names (i.e. SQL identifiers) are quoted accordingly based on the database dialect.

```go
type ACTOR struct {
    sq.TableInfo `sq:"name=Actor"`
    ACTOR_ID     sq.NumberField `sq:"name=ActorID"`
    FIRST_NAME   sq.StringField `sq:"name=firstname"`
    LastName     sq.StringField
    Last_Update  sq.TimeField
}

ACTOR := sq.New[ACTOR]() // FROM "Actor"
ACTOR.ACTOR_ID           // "Actor"."ActorID"
ACTOR.FIRST_NAME         // "Actor".firstname
ACTOR.LAST_NAME          // "Actor".lastname
ACTOR.LAST_UPDATE        // "Actor".last_update
```

### Available Field types

`sq` provides several built-in Field types for mapping column types to. Any other SQL type (e.g. INT[], TEXT[], TSVECTOR) would fall under `CustomField`.

| Field | Go Type | Default SQL type |
|-------|-----|-------|
| `sq.BlobField` | `[]byte` | `BLOB` (SQLite/MySQL), `BYTEA` (Postgres) |
| `sq.BooleanField` | `bool` | `BOOLEAN` |
| `sq.NumberField` | `int, int64, float64, etc` | `INT` |
| `sq.StringField` | `string` | `TEXT` (SQLite/Postgres), `VARCHAR(255)` (MySQL) |
| `sq.TimeField` | `time.Time` | `DATETIME` (SQLite/MySQL), `TIMESTAMPTZ` (Postgres) |
| `sq.JSONField` | any JSON compatible type | `JSON` (SQLite/MySQL), `JSONB` (Postgres) |
| `sq.UUIDField` | google/uuid | `BINARY(16)` (SQLite/MySQL), `UUID` (Postgres) |
| `sq.CustomField` | - | - |

TODO: consider moving the comprehensive reference of ddl annotations into ddl/docs instead. A separate documentation file can then be generated for ddl, and it will be hosted an a different URL.

For more information about type mapping, see the [documentation for ddl](#)

## How do I handle migrations?

See [migrations](#).
