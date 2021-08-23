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
