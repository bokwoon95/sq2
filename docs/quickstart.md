# {#quickstart} <a href="quickstart"></a>

sq is a type-safe query builder for Go that works by mapping your SQL tables to Go structs. These structs can then be maintained in one of three ways:

- **As a source of truth**
    - You write the structs, and generate SQL DDL from them
- **As a secondary source of truth**
    - You write the structs, and use it to check that the tables actually exist in the database
- **As code-generated output by introspecting the database**
    - your database is the source of truth, generate Go structs from it

Once you have your table-representative structs, you can use them in the various query builders.

```sql
-- sql
CREATE TABLE actor (
    actor_id BIGINT PRIMARY KEY
    ,first_name TEXT NOT NULL
    ,last_name TEXT NOT NULL
    ,last_update DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

```go
// go
type ACTOR struct {
    sq.TableInfo
    ACTOR_ID    sq.NumberField `ddl:"type=BIGINT primarykey"`
    FIRST_NAME  sq.StringField `ddl:"notnull"`
    LAST_NAME   sq.StringField `ddl:"notnull"`
    LAST_UPDATE sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP"`
}
```

You can then use these Go structs in the query builder (**SQLite**, **Postgres** and **MySQL** dialects are supported):

```go
// select
selectQuery := sq.SQLite.
    Select(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
    From(ACTOR).
    Where(ACTOR.ACTOR_ID.EqInt(1))
// insert
insertQuery := sq.SQLite.
    InsertInto(ACTOR).
    Columns(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
    Values("PENELOPE", "GUINESS")
// update
updateQuery := sq.SQLite.
    Update(ACTOR).
    Set(
        ACTOR.FIRST_NAME.SetString("Penelope"),
        ACTOR.LAST_NAME.SetString("Guiness"),
    ).
    Where(ACTOR.ACTOR_ID.EqInt(1))
// delete
deleteQuery := sq.SQLite.
    DeleteFrom(ACTOR).
    Where(ACTOR.ACTOR_ID.EqInt(1))
```

## Using
