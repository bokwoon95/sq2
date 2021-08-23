{{ block "01" . -}}

`sq` is a type-safe query builder and data mapper for Go. It supports the **SQLite**, **Postgres** and **MySQL** dialects. Among its features are:

- [**Type-safety**](#)
- [**Declarative schema as code (provided by the `ddl` companion package)**](#)
- [**Able to utilize Go generics for data fetching**](#)
- [**Emulates each SQL dialect faithfully**](#)
- [**Application-side Row Level Security (i.e. multitenancy support)**](#)

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

    // SELECT
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

    // print table contents
    actors = actors[:0]
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
