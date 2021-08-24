{{ block "01" . -}}

`sq` is a type-safe query builder and data mapper for Go. It supports the **SQLite**, **Postgres** and **MySQL** dialects. Among its features are:

- **Type-safety**
    - Every table gets a struct type.
    - Every column gets a struct field.
    - You no longer have to hardcode tables and columns as raw strings (even [ORMs](https://gorm.io/docs/query.html#Conditions) are guilty of this).
- **Two-way schema generation (with the companion `ddl` package)**
    - Generate table structs from your database.
    - Generate database DDL from your table structs.
    - Serialize your database schema into JSON.
        - This allows you to generate migrations against a production database without needing a live connection to it for introspection.
        - Instead you introspect the production database separately, serialize the result to JSON and use that as a reference during development.
    - Define your database schema in code declaratively.
        - `ddl` will figure out which columns or tables to add or remove.
    - `ddl` schema definition supports all major SQL features (even Postgres-specific ones):
        - (Materialized) Views, Extensions, Functions, Enums, Constraints, Indexes, Triggers
    - [more info](#)
- **Able to utilize Go generics for data fetching**
    - Data fetching uses [callback mapper functions](#).
        - Any column that you map in the callback function is automatically added to SELECT.
        - This means you select only the columns you need (say no to `SELECT *`).
    - These callback mapper functions can return any type you want, and that return type is automatically returned by the generic [FetchOne/FetchSlice](#) functions.
- **Use your favourite SQL dialect as-is**
    - `sq` does not abstract away SQL dialects. Use all dialect-specific features in their full glory.
    - Any missing features can be trivially added without cross-dialect compatibility headaches.
    - This does not mean that queries are not portable: queries are as portable as the SQL that you write.
    - Sticking with a common subset of SQL (e.g. ANSI SQL) means that the queries you write can be ported across different database dialects by simply changing the `Dialect` field.
- **Built-in Authorization support a.k.a Application-side Row Level Security**

NOTE: I can't remove the code sample, no matter how unsightly it is there's a chance someone will only browse the start and not the end.

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
    err = ddl.AutoMigrate(sq.DialectSQLite, db, ddl.CreateMissing, ddl.WithTables(ACTOR))
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
        sq.ErowsAffected,
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
        0,
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

    // print table contents (uses FetchAll)
    actors = actors[:0] // don't overwrite actors slice, create new slice. so that people can tinker with the code and print it side-by-side
    _, err = sq.Fetch(sq.Log(db), sq.SQLite.From(ACTOR).OrderBy(ACTOR.ACTOR_ID), func(row *sq.Row) {
        actor := Actor{
            ActorID:    row.Int(ACTOR.ACTOR_ID),
            FirstName:  row.String(ACTOR.FIRST_NAME),
            LastName:   row.String(ACTOR.LAST_NAME),
            LastUpdate: row.Time(ACTOR.LAST_UPDATE),
        }
        row.Process(func() { actors = append(actors, actor) })
    })
    log.Printf("actors: %+v\n", actors)
}
```

```sql
-- sql table
CREATE TABLE actor (
    actor_id BIGINT PRIMARY KEY
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

```go
// table struct
type ACTOR struct {
    sq.TableInfo
    ACTOR_ID    sq.NumberField `ddl:"type=BIGINT primarykey"`
    FIRST_NAME  sq.StringField `ddl:"notnull"`
    LAST_NAME   sq.StringField `ddl:"notnull"`
    LAST_UPDATE sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP"`
}

// select
sq.SQLite.Select(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).From(ACTOR).Where(ACTOR.ACTOR_ID.EqInt(1))
// insert
sq.SQLite.InsertInto(ACTOR).Columns(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).Values("PENELOPE", "GUINESS")
// update
sq.SQLite.Update(ACTOR).
    Set(
        ACTOR.FIRST_NAME.SetString("Penelope"),
        ACTOR.LAST_NAME.SetString("Guiness"),
    ).
    Where(ACTOR.ACTOR_ID.EqInt(1))
// delete
sq.SQLite.DeleteFrom(ACTOR).Where(ACTOR.ACTOR_ID.EqInt(1))
```

Unlike other query builders, sq covers the entire lifecycle of SQL management in an application.

These table structs can be maintained in one of three ways:

- **As a source of truth**
    - You write the structs, and generate SQL DDL from them
- **As a secondary source of truth**
    - You write the structs, and use it to check that the tables actually exist in the database
- **As code-generated output by introspecting the database**
    - your database is the source of truth, generate Go structs from it

Once you have your table-representative structs, you can use them in the query builders.

{{- end }}

{{ block "02" . -}}

- Fantastic material. [https://github.com/CourseOrchestra/2bass](https://github.com/CourseOrchestra/2bass)
- [https://dzone.com/articles/trouble-free-database-migration-idempotence-and-co](https://dzone.com/articles/trouble-free-database-migration-idempotence-and-co)

{{- end }}
