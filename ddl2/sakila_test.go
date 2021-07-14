package ddl2

import (
	"github.com/bokwoon95/sq"
)

func NEW_ACTOR(dialect, alias string) ACTOR {
	var tbl ACTOR
	tbl.TableInfo = sq.TableInfo{TableName: "actor", TableAlias: alias}
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	tbl.ACTOR_ID = sq.NewNumberField("actor_id", tbl.TableInfo)
	tbl.FIRST_NAME = sq.NewStringField("first_name", tbl.TableInfo)
	tbl.LAST_NAME = sq.NewStringField("last_name", tbl.TableInfo)
	tbl.FULL_NAME = sq.NewStringField("full_name", tbl.TableInfo)
	tbl.FULL_NAME_REVERSED = sq.NewStringField("full_name_reversed", tbl.TableInfo)
	tbl.LAST_UPDATE = sq.NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type ACTOR struct {
	sq.TableInfo
	ACTOR_ID           sq.NumberField `ddl:"primarykey"`
	FIRST_NAME         sq.StringField `ddl:"notnull"`
	LAST_NAME          sq.StringField `ddl:"notnull index"`
	FULL_NAME          sq.StringField `ddl:"generated={first_name || ' ' || last_name} virtual"`
	FULL_NAME_REVERSED sq.StringField `ddl:"generated={last_name || ' ' || first_name} stored"`
	LAST_UPDATE        sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl ACTOR) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectSQLite:
		t.Column(tbl.ACTOR_ID).Type("INTEGER").Autoincrement()
		t.Trigger(t.Sprintf(`CREATE TRIGGER actor_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.ACTOR_ID).Type("INT").Identity()
		t.Column(tbl.FULL_NAME).Generated("{} || ' ' || {}", tbl.FIRST_NAME, tbl.LAST_NAME).Stored()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Trigger(t.Sprintf(`CREATE TRIGGER actor_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.ACTOR_ID).Type("INT").Autoincrement()
		t.Column(tbl.FIRST_NAME).Type("VARCHAR(45)")
		t.Column(tbl.LAST_NAME).Type("VARCHAR(45)")
		t.Column(tbl.FULL_NAME).Type("VARCHAR(45)").Generated("CONCAT({}, ' ', {})", tbl.FIRST_NAME, tbl.LAST_NAME)
		t.Column(tbl.FULL_NAME_REVERSED).Config(func(c *Column) {
			c.ColumnType = "VARCHAR(45)"
			c.GeneratedExpr = t.Sprintf("CONCAT({}, ' ', {})", tbl.LAST_NAME, tbl.FIRST_NAME)
			c.GeneratedExprStored = true
		})
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_CATEGORY(dialect, alias string) CATEGORY {
	var tbl CATEGORY
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type CATEGORY struct {
	sq.TableInfo
	CATEGORY_ID sq.NumberField `ddl:"type=INTEGER primarykey"`
	NAME        sq.StringField `ddl:"notnull"`
	LAST_UPDATE sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl CATEGORY) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectSQLite:
		t.Trigger(t.Sprintf(`CREATE TRIGGER category_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.CATEGORY_ID).Type("INT").Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Trigger(t.Sprintf(`CREATE TRIGGER category_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.CATEGORY_ID).Type("INT").Autoincrement()
		t.Column(tbl.NAME).Type("VARCHAR(25)")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_COUNTRY(dialect, alias string) COUNTRY {
	var tbl COUNTRY
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type COUNTRY struct {
	sq.TableInfo
	COUNTRY_ID  sq.NumberField `ddl:"type=INTEGER primarykey"`
	COUNTRY     sq.StringField `ddl:"notnull"`
	LAST_UPDATE sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl COUNTRY) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectSQLite:
		t.Trigger(t.Sprintf(`CREATE TRIGGER country_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.COUNTRY_ID).Type("INT").Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Trigger(t.Sprintf(`CREATE TRIGGER country_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.COUNTRY_ID).Type("INT").Autoincrement()
		t.Column(tbl.COUNTRY).Type("VARCHAR(50)")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_CITY(dialect, alias string) CITY {
	var tbl CITY
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type CITY struct {
	sq.TableInfo
	CITY_ID     sq.NumberField `ddl:"type=INTEGER primarykey"`
	CITY        sq.StringField `ddl:"notnull"`
	COUNTRY_ID  sq.NumberField `ddl:"notnull references={country onupdate=cascade ondelete=restrict} index"`
	LAST_UPDATE sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl CITY) DDL(dialect string, t *T) {
	COUNTRY := NEW_COUNTRY(dialect, "")
	switch dialect {
	case sq.DialectSQLite:
		t.ForeignKey(tbl.COUNTRY_ID).References(COUNTRY, COUNTRY.COUNTRY_ID).OnUpdate(CASCADE).OnDelete(RESTRICT)
		t.Trigger(t.Sprintf(`CREATE TRIGGER city_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.CITY_ID).Type("INT").Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Trigger(t.Sprintf(`CREATE TRIGGER city_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.CITY_ID).Type("INT").Autoincrement()
		t.Column(tbl.CITY).Type("VARCHAR(50)")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
		t.ForeignKey(tbl.COUNTRY_ID).References(COUNTRY, COUNTRY.COUNTRY_ID).OnUpdate(CASCADE).OnDelete(CASCADE)
	}
}

func NEW_ADDRESS(dialect, alias string) ADDRESS {
	var tbl ADDRESS
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type ADDRESS struct {
	sq.TableInfo
	ADDRESS_ID  sq.NumberField `ddl:"type=INTEGER primarykey"`
	ADDRESS     sq.StringField `ddl:"notnull"`
	ADDRESS2    sq.StringField
	DISTRICT    sq.StringField `ddl:"notnull"`
	CITY_ID     sq.NumberField `ddl:"notnull references={city.city_id onupdate=cascade ondelete=restrict} index"`
	POSTAL_CODE sq.StringField
	PHONE       sq.StringField `ddl:"notnull"`
	LAST_UPDATE sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl ADDRESS) DDL(dialect string, t *T) {
	CITY := NEW_CITY(dialect, "")
	switch dialect {
	case sq.DialectSQLite:
		t.Trigger(t.Sprintf(`CREATE TRIGGER address_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.ADDRESS_ID).Type("INT").Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Trigger(t.Sprintf(`CREATE TRIGGER city_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.ADDRESS_ID).Type("INT").Autoincrement()
		t.Column(tbl.ADDRESS).Type("VARCHAR(50)")
		t.Column(tbl.ADDRESS2).Type("VARCHAR(50)")
		t.Column(tbl.DISTRICT).Type("VARCHAR(20)")
		t.ForeignKey(tbl.CITY_ID).References(CITY, CITY.CITY_ID).OnUpdate(CASCADE).OnDelete(CASCADE)
		t.Column(tbl.POSTAL_CODE).Type("VARCHAR(10)")
		t.Column(tbl.PHONE).Type("VARCHAR(20)")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_LANGUAGE(dialect, alias string) LANGUAGE {
	var tbl LANGUAGE
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type LANGUAGE struct {
	sq.TableInfo
	LANGUAGE_ID sq.NumberField `ddl:"type=INTEGER primarykey"`
	NAME        sq.StringField `ddl:"notnull"`
	LAST_UPDATE sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl LANGUAGE) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectSQLite:
		t.Trigger(t.Sprintf(`CREATE TRIGGER language_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.LANGUAGE_ID).Type("INT").Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Trigger(t.Sprintf(`CREATE TRIGGER language_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.LANGUAGE_ID).Type("INT").Autoincrement()
		t.Column(tbl.NAME).Type("CHAR(20)")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_FILM(dialect, alias string) FILM {
	var tbl FILM
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type FILM struct {
	sq.TableInfo
	FILM_ID              sq.NumberField `ddl:"type=INTEGER primarykey"`
	TITLE                sq.StringField `ddl:"notnull index"`
	DESCRIPTION          sq.StringField
	RELEASE_YEAR         sq.NumberField
	LANGUAGE_ID          sq.NumberField `ddl:"notnull references={language.language_id onupdate=cascade ondelete=restrict} index"`
	ORIGINAL_LANGUAGE_ID sq.NumberField `ddl:"references={language.language_id onupdate=cascade ondelete=restrict} index"`
	RENTAL_DURATION      sq.NumberField `ddl:"default=3 notnull"`
	RENTAL_RATE          sq.NumberField `ddl:"type=DECIMAL(4,2) default=4.99 notnull"`
	LENGTH               sq.NumberField
	REPLACEMENT_COST     sq.NumberField `ddl:"type=DECIMAL(5,2) default=19.99 notnull"`
	RATING               sq.StringField `ddl:"default='G'"`
	SPECIAL_FEATURES     sq.CustomField `ddl:"type=JSON"`
	LAST_UPDATE          sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
	FULLTEXT             sq.StringField
}

func (tbl FILM) DDL(dialect string, t *T) {
	FILM_TEXT := NEW_FILM_TEXT(dialect, "")
	switch dialect {
	case sq.DialectSQLite:
		New := func(field sq.Field) sq.Field { return sq.Literal("NEW." + field.GetName()) }
		Old := func(field sq.Field) sq.Field { return sq.Literal("OLD." + field.GetName()) }
		table := sq.Param("table", tbl)
		ftsTable := sq.Param("ftsTable", FILM_TEXT)
		ftsFields := sq.Param("ftsFields", sq.Fields{sq.Literal("ROWID"), FILM_TEXT.TITLE, FILM_TEXT.DESCRIPTION})
		insertValues := sq.Param("insertValues", sq.Fields{New(tbl.FILM_ID), New(tbl.TITLE), New(tbl.DESCRIPTION)})
		deleteValues := sq.Param("deleteValues", sq.Fields{Old(tbl.FILM_ID), Old(tbl.TITLE), Old(tbl.DESCRIPTION)})
		t.Column(tbl.FULLTEXT).Ignore()
		t.Check("film_release_year_check", "{1} >= 1901 AND {1} <= 2155", tbl.RELEASE_YEAR)
		t.Check("film_rating_check", "{} IN ('G','PG','PG-13','R','NC-17')", tbl.RATING)
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_fts5_after_insert_trg AFTER INSERT ON {table} BEGIN
	INSERT INTO {ftsTable} ({ftsFields}) VALUES ({insertValues});
END;`, table, ftsTable, ftsFields, insertValues))
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_fts5_after_delete_trg AFTER DELETE ON {table} BEGIN
	INSERT INTO {ftsTable} ({ftsTable}, {ftsFields}) VALUES ('delete', {deleteValues});
END;`, table, ftsTable, ftsFields, deleteValues))
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_fts5_after_update_trg AFTER UPDATE ON {table} BEGIN
	INSERT INTO {ftsTable} ({ftsTable}, {ftsFields}) VALUES ('delete', {deleteValues});
	INSERT INTO {ftsTable} ({ftsTable}, {ftsFields}) VALUES ({insertValues});
END;`, table, ftsTable, ftsFields, deleteValues, insertValues))
	case sq.DialectPostgres:
		t.Column(tbl.FILM_ID).Type("INT").Identity()
		t.Column(tbl.RELEASE_YEAR).Type("year")
		t.Column(tbl.RATING).Type("mpaa_rating").Default("'G'::mpaa_rating")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Column(tbl.SPECIAL_FEATURES).Type("TEXT[]")
		t.Column(tbl.FULLTEXT).Type("TSVECTOR")
		t.Index(tbl.FULLTEXT).Using("GIST")
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_fulltext_before_insert_update_trg BEFORE INSERT OR UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE tsvector_update_trigger(fulltext, 'pg_catalog.english', title, description);`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.FILM_ID).Type("INT").Autoincrement()
		t.Column(tbl.TITLE).Type("VARCHAR(255)")
		t.Column(tbl.DESCRIPTION).Type("TEXT")
		t.Column(tbl.RATING).Type("ENUM('G','PG','PG-13','R','NC-17')")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
		t.Column(tbl.FULLTEXT).Ignore()
		t.Check("film_release_year_check", "{1} >= 1901 AND {1} <= 2155", tbl.RELEASE_YEAR)
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_after_insert_trg AFTER INSERT ON film FOR EACH ROW BEGIN
	INSERT INTO film_text (film_id, title, description) VALUES (NEW.film_id, NEW.title, NEW.description);
END`))
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_after_update_trg AFTER UPDATE ON film FOR EACH ROW BEGIN
	IF OLD.title <> NEW.title OR OLD.description <> NEW.description THEN
		UPDATE film_text
		SET title = NEW.title, description = NEW.description, film_id = NEW.film_id
		WHERE film_id = OLD.film_id;
	END IF;
END`))
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_after_delete_trg AFTER DELETE ON film FOR EACH ROW BEGIN
	DELETE FROM film_text WHERE film_id = OLD.film_id;
END`))
	}
}

func NEW_FILM_TEXT(dialect, alias string) FILM_TEXT {
	var tbl FILM_TEXT
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type FILM_TEXT struct {
	sq.TableInfo
	FILM_ID     sq.NumberField
	TITLE       sq.StringField
	DESCRIPTION sq.StringField
}

func (tbl FILM_TEXT) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectSQLite:
		t.VirtualTable("fts5", `content='film'`, `content_rowid='film_id'`)
		t.Column(tbl.FILM_ID).Ignore()
	case sq.DialectPostgres:
		break // no-op, postgres does not need a separate film_text table for full text search
	case sq.DialectMySQL:
		t.Column(tbl.FILM_ID).Type("INT").NotNull().PrimaryKey()
		t.Column(tbl.TITLE).Type("VARCHAR(255)").NotNull()
		t.Column(tbl.DESCRIPTION).Type("TEXT")
		t.Index(tbl.TITLE, tbl.DESCRIPTION).Using("FULLTEXT")
	}
}

func NEW_FILM_ACTOR(dialect, alias string) FILM_ACTOR {
	var tbl FILM_ACTOR
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type FILM_ACTOR struct {
	sq.TableInfo `ddl:"index={. cols=actor_id,film_id unique}"`
	ACTOR_ID     sq.NumberField `ddl:"notnull references={actor.actor_id onupdate=cascade ondelete=restrict}"`
	FILM_ID      sq.NumberField `ddl:"notnull references={film.film_id onupdate=cascade ondelete=restrict} index"`
	LAST_UPDATE  sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl FILM_ACTOR) DDL(dialect string, t *T) {
	t.Index(tbl.ACTOR_ID, tbl.FILM_ID).Unique()
	switch dialect {
	case sq.DialectSQLite:
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_actor_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_actor_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_FILM_CATEGORY(dialect, alias string) FILM_CATEGORY {
	var tbl FILM_CATEGORY
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type FILM_CATEGORY struct {
	sq.TableInfo
	FILM_ID     sq.NumberField `ddl:"notnull references={film.film_id onupdate=cascade ondelete=restrict}"`
	CATEGORY_ID sq.NumberField `ddl:"notnull references={category.category_id onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl FILM_CATEGORY) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectSQLite:
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_category_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Trigger(t.Sprintf(`CREATE TRIGGER film_category_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_STAFF(dialect, alias string) STAFF {
	var tbl STAFF
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type STAFF struct {
	sq.TableInfo
	STAFF_ID    sq.NumberField `ddl:"type=INTEGER primarykey"`
	FIRST_NAME  sq.StringField `ddl:"notnull"`
	LAST_NAME   sq.StringField `ddl:"notnull"`
	ADDRESS_ID  sq.NumberField `ddl:"notnull references={address.address_id onupdate=cascade ondelete=restrict}"`
	EMAIL       sq.StringField
	STORE_ID    sq.NumberField  `ddl:"references=store.store_id"`
	ACTIVE      sq.BooleanField `ddl:"default=TRUE notnull"`
	USERNAME    sq.StringField  `ddl:"notnull"`
	PASSWORD    sq.StringField
	LAST_UPDATE sq.TimeField `ddl:"default=DATETIME('now') notnull"`
	PICTURE     sq.BlobField
}

func (tbl STAFF) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectSQLite:
		t.Trigger(t.Sprintf(`CREATE TRIGGER staff_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.STAFF_ID).Type("INT").Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Column(tbl.PICTURE).Type("BYTEA")
		t.Trigger(t.Sprintf(`CREATE TRIGGER staff_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.STAFF_ID).Type("INT").Autoincrement()
		t.Column(tbl.FIRST_NAME).Type("VARCHAR(45)")
		t.Column(tbl.LAST_NAME).Type("VARCHAR(45)")
		t.Column(tbl.EMAIL).Type("VARCHAR(50)")
		t.Column(tbl.USERNAME).Type("VARCHAR(16)")
		t.Column(tbl.PASSWORD).Type("VARCHAR(40)")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_STORE(dialect, alias string) STORE {
	var tbl STORE
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type STORE struct {
	sq.TableInfo
	STORE_ID         sq.NumberField `ddl:"type=INTEGER primarykey"`
	MANAGER_STAFF_ID sq.NumberField `ddl:"notnull references={staff.staff_id onupdate=cascade ondelete=restrict} index={. unique}"`
	ADDRESS_ID       sq.NumberField `ddl:"notnull references={address.address_id onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE      sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl STORE) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectSQLite:
		t.Trigger(t.Sprintf(`CREATE TRIGGER store_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.STORE_ID).Type("INT").Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Trigger(t.Sprintf(`CREATE TRIGGER store_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.STORE_ID).Type("INT").Autoincrement()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_CUSTOMER(dialect, alias string) CUSTOMER {
	var tbl CUSTOMER
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type CUSTOMER struct {
	sq.TableInfo `sq:"name=customer" ddl:"unique={. cols=email,first_name,last_name}"`
	CUSTOMER_ID  sq.NumberField  `ddl:"type=INTEGER primarykey"`
	STORE_ID     sq.NumberField  `ddl:"notnull index"`
	FIRST_NAME   sq.StringField  `ddl:"notnull"`
	LAST_NAME    sq.StringField  `ddl:"notnull index"`
	EMAIL        sq.StringField  `ddl:"unique"`
	ADDRESS_ID   sq.NumberField  `ddl:"notnull references={address.address_id onupdate=cascade ondelete=restrict} index"`
	ACTIVE       sq.BooleanField `ddl:"default=TRUE notnull"`
	DATA         sq.JSONField
	CREATE_DATE  sq.TimeField `ddl:"default=DATETIME('now') notnull"`
	LAST_UPDATE  sq.TimeField `ddl:"default=DATETIME('now')"`
}

func (tbl CUSTOMER) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectSQLite:
		t.Trigger(t.Sprintf(`CREATE TRIGGER customer_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.CUSTOMER_ID).Type("INT").Identity()
		t.Column(tbl.CREATE_DATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Trigger(t.Sprintf(`CREATE TRIGGER customer_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.CUSTOMER_ID).Type("INT").Autoincrement()
		t.Column(tbl.FIRST_NAME).Type("VARCHAR(45)")
		t.Column(tbl.LAST_NAME).Type("VARCHAR(45)")
		t.Column(tbl.EMAIL).Type("VARCHAR(50)")
		t.Column(tbl.CREATE_DATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_INVENTORY(dialect, alias string) INVENTORY {
	var tbl INVENTORY
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type INVENTORY struct {
	sq.TableInfo `sq:"name=inventory" ddl:"index={. cols=store_id,film_id}"`
	INVENTORY_ID sq.NumberField `ddl:"type=INTEGER primarykey"`
	FILM_ID      sq.NumberField `ddl:"notnull references={film.film_id onupdate=cascade ondelete=restrict}"`
	STORE_ID     sq.NumberField `ddl:"notnull references={store.store_id onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE  sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl INVENTORY) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectSQLite:
		t.Trigger(t.Sprintf(`CREATE TRIGGER inventory_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.INVENTORY_ID).Type("INT").Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Trigger(t.Sprintf(`CREATE TRIGGER inventory_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.INVENTORY_ID).Type("INT").Autoincrement()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_RENTAL(dialect, alias string) RENTAL {
	var tbl RENTAL
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type RENTAL struct {
	sq.TableInfo `ddl:"index={. cols=rental_date,inventory_id,customer_id unique}"`
	RENTAL_ID    sq.NumberField `ddl:"type=INTEGER primarykey"`
	RENTAL_DATE  sq.TimeField   `ddl:"notnull"`
	INVENTORY_ID sq.NumberField `ddl:"notnull index references={inventory.inventory_id onupdate=cascade ondelete=restrict}"`
	CUSTOMER_ID  sq.NumberField `ddl:"notnull index references={customer.customer_id onupdate=cascade ondelete=restrict}"`
	RETURN_DATE  sq.TimeField
	STAFF_ID     sq.NumberField `ddl:"notnull index references={staff.staff_id onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE  sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl RENTAL) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectSQLite:
		t.Trigger(t.Sprintf(`CREATE TRIGGER rental_last_update_after_update_trg AFTER UPDATE ON {1} BEGIN
	UPDATE {1} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`, tbl))
	case sq.DialectPostgres:
		t.Column(tbl.RENTAL_ID).Type("INT").Identity()
		t.Column(tbl.RETURN_DATE).Type("TIMESTAMPTZ")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Trigger(t.Sprintf(`CREATE TRIGGER rental_last_update_before_update_trg BEFORE UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`, tbl))
	case sq.DialectMySQL:
		t.Column(tbl.RENTAL_ID).Type("INT").Autoincrement()
		t.Column(tbl.RETURN_DATE).Type("TIMESTAMP")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_PAYMENT(dialect, alias string) PAYMENT {
	var tbl PAYMENT
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type PAYMENT struct {
	sq.TableInfo
	PAYMENT_ID   sq.NumberField `ddl:"type=INTEGER primarykey"`
	CUSTOMER_ID  sq.NumberField `ddl:"notnull index references={customer.customer_id onupdate=cascade ondelete=restrict}"`
	STAFF_ID     sq.NumberField `ddl:"notnull index references={staff.staff_id onupdate=cascade ondelete=restrict}"`
	RENTAL_ID    sq.NumberField `ddl:"references={rental.rental_id onupdate=cascade ondelete=restrict}"`
	AMOUNT       sq.NumberField `ddl:"type=DECIMAL(5,2) notnull"`
	PAYMENT_DATE sq.TimeField   `ddl:"notnull"`
}

func (tbl PAYMENT) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.PAYMENT_ID).Type("INT").Identity()
		t.Column(tbl.PAYMENT_DATE).Type("TIMESTAMPTZ")
	case sq.DialectMySQL:
		t.Column(tbl.PAYMENT_ID).Type("INT").Autoincrement()
		t.Column(tbl.PAYMENT_DATE).Type("TIMESTAMP")
	}
}

func NEW_DUMMY_TABLE(dialect, alias string) DUMMY_TABLE {
	var tbl DUMMY_TABLE
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type DUMMY_TABLE struct {
	sq.TableInfo `ddl:"primarykey={. cols=id1,id2} unique={. cols=score,color}"`
	ID1          sq.NumberField
	ID2          sq.StringField
	SCORE        sq.NumberField
	COLOR        sq.StringField `ddl:"collate=nocase default='red'"`
	DATA         sq.JSONField
}

func (tbl DUMMY_TABLE) DDL(dialect string, t *T) {
	const indexName = "dummy_table_complex_expr_idx"
	switch dialect {
	case sq.DialectSQLite:
		t.Column(tbl.COLOR).Collate("nocase")
		t.NameIndex(indexName,
			tbl.SCORE,
			sq.Fieldf("SUBSTR({}, 1, 2)", tbl.COLOR),
			sq.Fieldf("{} || {}", tbl.COLOR, " abcd"),
			sq.Fieldf("CAST(JSON_EXTRACT({}, {}) AS INT)", tbl.DATA, "$.age"),
		).Where("{} = {}", tbl.COLOR, "red")
	case sq.DialectPostgres:
		t.Column(tbl.ID1).AlwaysIdentity()
		t.Column(tbl.COLOR).Collate("C")
		t.NameIndex(indexName,
			tbl.SCORE,
			sq.Fieldf("SUBSTR({}, 1, 2)", tbl.COLOR),
			sq.Fieldf("{} || {}", tbl.COLOR, " abcd"),
			sq.Fieldf("({}->>{})::INT", tbl.DATA, "age"),
		).Where("{} = {}", tbl.COLOR, "red")
		t.NameIndex("dummy_table_id2_idx", sq.Literal(`id2 COLLATE "C"`))
		t.NameIndex("dummy_table_color_idx", sq.Literal("color text_pattern_ops"))
	case sq.DialectMySQL:
		t.Column(tbl.COLOR).Type("VARCHAR(50)").Collate("latin1_swedish_ci")
		t.NameIndex(indexName,
			tbl.SCORE,
			sq.Fieldf("SUBSTR({}, 1, 2)", tbl.COLOR),
			sq.Fieldf("CONCAT({}, {})", tbl.COLOR, " abcd"),
			sq.Fieldf("CAST({}->>{} AS SIGNED)", tbl.DATA, "$.age"),
		)
	}
	t.Check("dummy_table_score_positive_check", "{} > 0", tbl.SCORE)
	t.Check("dummy_table_score_id1_greater_than_check", "{} > {}", tbl.SCORE, tbl.ID1)
	t.PrimaryKey(tbl.ID1, tbl.ID2)
	t.Unique(tbl.SCORE, tbl.COLOR)
}

func NEW_DUMMY_TABLE_2(dialect, alias string) DUMMY_TABLE_2 {
	var tbl DUMMY_TABLE_2
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type DUMMY_TABLE_2 struct {
	sq.TableInfo `ddl:"references={dummy_table.id1,id2 cols=id1,id2 onupdate=cascade ondelete=restrict}"`
	ID1          sq.NumberField
	ID2          sq.StringField
}

func (tbl DUMMY_TABLE_2) DDL(dialect string, t *T) {
	ref := NEW_DUMMY_TABLE(dialect, "")
	switch dialect {
	case sq.DialectPostgres, sq.DialectMySQL:
		t.ForeignKey(tbl.ID1, tbl.ID2).References(ref, ref.ID1, ref.ID2).OnUpdate("CASCADE").OnDelete("RESTRICT")
	}
}

func json_object_agg(name, value interface{}) sq.CustomField {
	if query, ok := value.(sq.Query); ok {
		value = sq.Fieldf("({})", query)
	}
	return sq.FieldfDialect(map[string]string{
		"default":        "jsonb_object_agg({}, {})",
		sq.DialectSQLite: "json_group_object({}, {})",
		sq.DialectMySQL:  "json_objectagg({}, {})",
	}, name, value)
}

func json_array_agg(value interface{}) sq.CustomField {
	return sq.FieldfDialect(map[string]string{
		"default":        "jsonb_agg({})",
		sq.DialectSQLite: "json_group_array({})",
		sq.DialectMySQL:  "jsonb_arrayagg({})",
	}, value)
}

func NEW_ACTOR_INFO(dialect, alias string) ACTOR_INFO {
	var tbl ACTOR_INFO
	tbl.TableInfo = sq.TableInfo{TableName: "actor_info", TableAlias: alias}
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	tbl.ACTOR_ID = sq.NewNumberField("actor_id", tbl.TableInfo)
	tbl.FIRST_NAME = sq.NewStringField("first_name", tbl.TableInfo)
	tbl.LAST_NAME = sq.NewStringField("last_name", tbl.TableInfo)
	tbl.FILM_INFO = sq.NewJSONField("film_info", tbl.TableInfo)
	return tbl
}

type ACTOR_INFO struct {
	sq.TableInfo
	ACTOR_ID   sq.NumberField
	FIRST_NAME sq.StringField
	LAST_NAME  sq.StringField
	FILM_INFO  sq.JSONField
}

func (view ACTOR_INFO) DDL(dialect string, v *V) {
	ACTOR := NEW_ACTOR(dialect, "a")
	FILM := NEW_FILM(dialect, "f")
	FILM_ACTOR := NEW_FILM_ACTOR(dialect, "fa")
	FILM_CATEGORY := NEW_FILM_CATEGORY(dialect, "fc")
	CATEGORY := NEW_CATEGORY(dialect, "c")
	v.AsQuery(sq.SQLite.
		From(ACTOR).
		LeftJoin(FILM_ACTOR, FILM_ACTOR.ACTOR_ID.Eq(ACTOR.ACTOR_ID)).
		LeftJoin(FILM_CATEGORY, FILM_CATEGORY.FILM_ID.Eq(FILM_ACTOR.FILM_ID)).
		LeftJoin(CATEGORY, CATEGORY.CATEGORY_ID.Eq(FILM_CATEGORY.CATEGORY_ID)).
		GroupBy(ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
		Select(
			ACTOR.ACTOR_ID,
			ACTOR.FIRST_NAME,
			ACTOR.LAST_NAME,
			json_object_agg(CATEGORY.NAME, sq.SQLite.
				Select(json_array_agg(FILM.TITLE)).
				From(FILM).
				Join(FILM_CATEGORY, FILM_CATEGORY.FILM_ID.Eq(FILM.FILM_ID)).
				Join(FILM_ACTOR, FILM_ACTOR.FILM_ID.Eq(FILM.FILM_ID)).
				Where(
					FILM_CATEGORY.CATEGORY_ID.Eq(CATEGORY.CATEGORY_ID),
					FILM_ACTOR.ACTOR_ID.Eq(ACTOR.ACTOR_ID),
				).
				GroupBy(FILM_ACTOR.ACTOR_ID),
			).As("film_info"),
		),
	)
}

func NEW_CUSTOMER_LIST(dialect, alias string) CUSTOMER_LIST {
	var tbl CUSTOMER_LIST
	tbl.TableInfo = sq.TableInfo{TableName: "customer_list", TableAlias: alias}
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	tbl.ID = sq.NewNumberField("id", tbl.TableInfo)
	tbl.NAME = sq.NewStringField("name", tbl.TableInfo)
	tbl.ADDRESS = sq.NewStringField("address", tbl.TableInfo)
	tbl.ZIP_CODE = sq.NewStringField("zip code", tbl.TableInfo)
	tbl.PHONE = sq.NewStringField("phone", tbl.TableInfo)
	tbl.CITY = sq.NewStringField("city", tbl.TableInfo)
	tbl.COUNTRY = sq.NewStringField("country", tbl.TableInfo)
	tbl.NOTES = sq.NewStringField("notes", tbl.TableInfo)
	tbl.SID = sq.NewNumberField("sid", tbl.TableInfo)
	return tbl
}

type CUSTOMER_LIST struct {
	sq.TableInfo
	ID       sq.NumberField
	NAME     sq.StringField
	ADDRESS  sq.StringField
	ZIP_CODE sq.StringField
	PHONE    sq.StringField
	CITY     sq.StringField
	COUNTRY  sq.StringField
	NOTES    sq.StringField
	SID      sq.NumberField
}

func (view CUSTOMER_LIST) DDL(dialect string, v *V) {
	CUSTOMER := NEW_CUSTOMER(dialect, "cu")
	ADDRESS := NEW_ADDRESS(dialect, "a")
	CITY := NEW_CITY(dialect, "")
	COUNTRY := NEW_COUNTRY(dialect, "")
	v.AsQuery(sq.SQLite.
		From(CUSTOMER).
		Join(ADDRESS, ADDRESS.ADDRESS_ID.Eq(CUSTOMER.ADDRESS_ID)).
		Join(CITY, CITY.CITY_ID.Eq(ADDRESS.CITY_ID)).
		Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
		Select(
			CUSTOMER.CUSTOMER_ID.As("id"),
			sq.FieldfDialect(map[string]string{
				"default":       "{} || ' ' || {}",
				sq.DialectMySQL: "CONCAT({}, ' ', {})",
			}, CUSTOMER.FIRST_NAME, CUSTOMER.LAST_NAME).As("name"),
			ADDRESS.ADDRESS,
			ADDRESS.POSTAL_CODE.As("zip code"),
			ADDRESS.PHONE,
			CITY.CITY,
			COUNTRY.COUNTRY,
			sq.CaseWhen(CUSTOMER.ACTIVE, "active").Else("").As("notes"),
			CUSTOMER.STORE_ID.As("sid"),
		),
	)
}

func NEW_FILM_LIST(dialect, alias string) FILM_LIST {
	var tbl FILM_LIST
	tbl.TableInfo = sq.TableInfo{TableName: "film_list", TableAlias: alias}
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	tbl.FID = sq.NewNumberField("fid", tbl.TableInfo)
	tbl.TITLE = sq.NewStringField("title", tbl.TableInfo)
	tbl.DESCRIPTION = sq.NewStringField("description", tbl.TableInfo)
	tbl.CATEGORY = sq.NewStringField("category", tbl.TableInfo)
	tbl.PRICE = sq.NewNumberField("price", tbl.TableInfo)
	tbl.LENGTH = sq.NewNumberField("length", tbl.TableInfo)
	tbl.RATING = sq.NewStringField("rating", tbl.TableInfo)
	tbl.ACTORS = sq.NewJSONField("actors", tbl.TableInfo)
	return tbl
}

type FILM_LIST struct {
	sq.TableInfo
	FID         sq.NumberField
	TITLE       sq.StringField
	DESCRIPTION sq.StringField
	CATEGORY    sq.StringField
	PRICE       sq.NumberField
	LENGTH      sq.NumberField
	RATING      sq.StringField
	ACTORS      sq.JSONField
}

func (view FILM_LIST) DDL(dialect string, v *V) {
	CATEGORY := NEW_CATEGORY(dialect, "")
	FILM_CATEGORY := NEW_FILM_CATEGORY(dialect, "")
	FILM := NEW_FILM(dialect, "")
	FILM_ACTOR := NEW_FILM_ACTOR(dialect, "")
	ACTOR := NEW_ACTOR(dialect, "")
	v.AsQuery(sq.SQLite.
		From(CATEGORY).
		LeftJoin(FILM_CATEGORY, FILM_CATEGORY.CATEGORY_ID.Eq(CATEGORY.CATEGORY_ID)).
		LeftJoin(FILM, FILM.FILM_ID.Eq(FILM_CATEGORY.FILM_ID)).
		Join(FILM_ACTOR, FILM_ACTOR.FILM_ID.Eq(FILM.FILM_ID)).
		Join(ACTOR, ACTOR.ACTOR_ID.Eq(FILM_ACTOR.ACTOR_ID)).
		GroupBy(
			FILM.FILM_ID,
			FILM.TITLE,
			FILM.DESCRIPTION,
			CATEGORY.NAME,
			FILM.RENTAL_RATE,
			FILM.LENGTH,
			FILM.RATING,
		).
		Select(
			FILM.FILM_ID.As("fid"),
			FILM.TITLE,
			FILM.DESCRIPTION,
			CATEGORY.NAME.As("category"),
			FILM.RENTAL_RATE.As("price"),
			FILM.LENGTH,
			FILM.RATING,
			json_array_agg(
				sq.FieldfDialect(map[string]string{
					"default":       "{} || ' ' || {}",
					sq.DialectMySQL: "CONCAT({}, ' ', {})",
				}, ACTOR.FIRST_NAME, ACTOR.LAST_NAME),
			).As("actors"),
		),
	)
}

func NEW_NICER_BUT_SLOWER_FILM_LIST(dialect, alias string) NICER_BUT_SLOWER_FILM_LIST {
	var tbl NICER_BUT_SLOWER_FILM_LIST
	tbl.TableInfo = sq.TableInfo{TableName: "nicer_but_slower_film_list", TableAlias: alias}
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	tbl.FID = sq.NewNumberField("fid", tbl.TableInfo)
	tbl.TITLE = sq.NewStringField("title", tbl.TableInfo)
	tbl.DESCRIPTION = sq.NewStringField("description", tbl.TableInfo)
	tbl.CATEGORY = sq.NewStringField("category", tbl.TableInfo)
	tbl.PRICE = sq.NewNumberField("price", tbl.TableInfo)
	tbl.LENGTH = sq.NewNumberField("length", tbl.TableInfo)
	tbl.RATING = sq.NewStringField("rating", tbl.TableInfo)
	tbl.ACTORS = sq.NewJSONField("actors", tbl.TableInfo)
	return tbl
}

type NICER_BUT_SLOWER_FILM_LIST struct {
	sq.TableInfo
	FID         sq.NumberField
	TITLE       sq.StringField
	DESCRIPTION sq.StringField
	CATEGORY    sq.StringField
	PRICE       sq.NumberField
	LENGTH      sq.NumberField
	RATING      sq.StringField
	ACTORS      sq.JSONField
}

func (view NICER_BUT_SLOWER_FILM_LIST) DDL(dialect string, v *V) {
	CATEGORY := NEW_CATEGORY(dialect, "")
	FILM_CATEGORY := NEW_FILM_CATEGORY(dialect, "")
	FILM := NEW_FILM(dialect, "")
	FILM_ACTOR := NEW_FILM_ACTOR(dialect, "")
	ACTOR := NEW_ACTOR(dialect, "")
	v.AsQuery(sq.SQLite.
		From(CATEGORY).
		LeftJoin(FILM_CATEGORY, FILM_CATEGORY.CATEGORY_ID.Eq(CATEGORY.CATEGORY_ID)).
		LeftJoin(FILM, FILM.FILM_ID.Eq(FILM_CATEGORY.FILM_ID)).
		Join(FILM_ACTOR, FILM_ACTOR.FILM_ID.Eq(FILM.FILM_ID)).
		Join(ACTOR, ACTOR.ACTOR_ID.Eq(FILM_ACTOR.ACTOR_ID)).
		GroupBy(
			FILM.FILM_ID,
			FILM.TITLE,
			FILM.DESCRIPTION,
			CATEGORY.NAME,
			FILM.RENTAL_RATE,
			FILM.LENGTH,
			FILM.RATING,
		).
		Select(
			FILM.FILM_ID.As("fid"),
			FILM.TITLE,
			FILM.DESCRIPTION,
			CATEGORY.NAME.As("category"),
			FILM.RENTAL_RATE.As("price"),
			FILM.LENGTH,
			FILM.RATING,
			json_array_agg(sq.FieldfDialect(map[string]string{
				"default": "UPPER(SUBSTRING({1}, 1, 1))" +
					" || LOWER(SUBSTRING({1}, 2))" +
					" || ' '" +
					" || UPPER(SUBSTRING({2}, 1, 1))" +
					" || LOWER(SUBSTRING({2}, 2))",
				sq.DialectMySQL: "CONCAT(UPPER(SUBSTRING({1}, 1, 1))" +
					", LOWER(SUBSTRING({1}, 2))" +
					", ' '" +
					", UPPER(SUBSTRING({2}, 1, 1))" +
					", LOWER(SUBSTRING({2}, 2)))",
			}, ACTOR.FIRST_NAME, ACTOR.LAST_NAME)).As("actors"),
		),
	)
}

func NEW_SALES_BY_FILM_CATEGORY(dialect, alias string) SALES_BY_FILM_CATEGORY {
	var tbl SALES_BY_FILM_CATEGORY
	tbl.TableInfo = sq.TableInfo{TableName: "sales_by_film_category", TableAlias: alias}
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	tbl.CATEGORY = sq.NewStringField("category", tbl.TableInfo)
	tbl.TOTAL_SALES = sq.NewNumberField("total_sales", tbl.TableInfo)
	return tbl
}

type SALES_BY_FILM_CATEGORY struct {
	sq.TableInfo
	CATEGORY    sq.StringField
	TOTAL_SALES sq.NumberField
}

func (view SALES_BY_FILM_CATEGORY) DDL(dialect string, v *V) {
	PAYMENT := NEW_PAYMENT(dialect, "p")
	RENTAL := NEW_RENTAL(dialect, "r")
	INVENTORY := NEW_INVENTORY(dialect, "i")
	FILM := NEW_FILM(dialect, "f")
	FILM_CATEGORY := NEW_FILM_CATEGORY(dialect, "fc")
	CATEGORY := NEW_CATEGORY(dialect, "c")
	v.AsQuery(sq.SQLite.
		From(PAYMENT).
		Join(RENTAL, RENTAL.RENTAL_ID.Eq(PAYMENT.RENTAL_ID)).
		Join(INVENTORY, INVENTORY.INVENTORY_ID.Eq(RENTAL.INVENTORY_ID)).
		Join(FILM, FILM.FILM_ID.Eq(INVENTORY.FILM_ID)).
		Join(FILM_CATEGORY, FILM_CATEGORY.FILM_ID.Eq(FILM.FILM_ID)).
		Join(CATEGORY, CATEGORY.CATEGORY_ID.Eq(FILM_CATEGORY.CATEGORY_ID)).
		GroupBy(CATEGORY.NAME).
		OrderBy(sq.Fieldf("SUM({})", PAYMENT.AMOUNT).Desc()).
		Select(
			CATEGORY.NAME.As("category"),
			sq.Fieldf("SUM({})", PAYMENT.AMOUNT).As("total_sales"),
		),
	)
}

func NEW_SALES_BY_STORE(dialect, alias string) SALES_BY_STORE {
	var tbl SALES_BY_STORE
	tbl.TableInfo = sq.TableInfo{TableName: "sales_by_store", TableAlias: alias}
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	tbl.STORE = sq.NewStringField("store", tbl.TableInfo)
	tbl.MANAGER = sq.NewStringField("manager", tbl.TableInfo)
	tbl.TOTAL_SALES = sq.NewNumberField("total_sales", tbl.TableInfo)
	return tbl
}

type SALES_BY_STORE struct {
	sq.TableInfo
	STORE       sq.StringField
	MANAGER     sq.StringField
	TOTAL_SALES sq.NumberField
}

func (view SALES_BY_STORE) DDL(dialect string, v *V) {
	PAYMENT := NEW_PAYMENT(dialect, "p")
	RENTAL := NEW_RENTAL(dialect, "r")
	INVENTORY := NEW_INVENTORY(dialect, "i")
	STORE := NEW_STORE(dialect, "s")
	ADDRESS := NEW_ADDRESS(dialect, "a")
	CITY := NEW_CITY(dialect, "ci")
	COUNTRY := NEW_COUNTRY(dialect, "co")
	STAFF := NEW_STAFF(dialect, "m")
	v.AsQuery(sq.SQLite.
		From(PAYMENT).
		Join(RENTAL, RENTAL.RENTAL_ID.Eq(PAYMENT.RENTAL_ID)).
		Join(INVENTORY, INVENTORY.INVENTORY_ID.Eq(RENTAL.INVENTORY_ID)).
		Join(STORE, STORE.STORE_ID.Eq(INVENTORY.STORE_ID)).
		Join(ADDRESS, ADDRESS.ADDRESS_ID.Eq(STORE.ADDRESS_ID)).
		Join(CITY, CITY.CITY_ID.Eq(ADDRESS.CITY_ID)).
		Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
		Join(STAFF, STAFF.STAFF_ID.Eq(STORE.MANAGER_STAFF_ID)).
		GroupBy(
			COUNTRY.COUNTRY,
			CITY.CITY,
			STORE.STORE_ID,
			STAFF.FIRST_NAME,
			STAFF.LAST_NAME,
		).
		OrderBy(
			COUNTRY.COUNTRY,
			CITY.CITY,
		).
		Select(
			sq.FieldfDialect(map[string]string{
				"default":       "{} || ',' || {}",
				sq.DialectMySQL: "CONCAT({}, ',', {})",
			}, CITY.CITY, COUNTRY.COUNTRY).As("store"),
			sq.FieldfDialect(map[string]string{
				"default":       "{} || ' ' || {}",
				sq.DialectMySQL: "CONCAT({}, ' ', {})",
			}, STAFF.FIRST_NAME, STAFF.LAST_NAME).As("manager"),
			sq.Fieldf("SUM({})", PAYMENT.AMOUNT).As("total_sales"),
		),
	)
}

func NEW_STAFF_LIST(dialect, alias string) STAFF_LIST {
	var tbl STAFF_LIST
	tbl.TableInfo = sq.TableInfo{TableName: "staff_list", TableAlias: alias}
	switch dialect {
	case sq.DialectPostgres:
		tbl.TableInfo.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.TableInfo.TableSchema = "db"
	}
	tbl.ID = sq.NewNumberField("id", tbl.TableInfo)
	tbl.NAME = sq.NewStringField("name", tbl.TableInfo)
	tbl.ADDRESS = sq.NewStringField("address", tbl.TableInfo)
	tbl.ZIP_CODE = sq.NewStringField("zip code", tbl.TableInfo)
	tbl.PHONE = sq.NewStringField("phone", tbl.TableInfo)
	tbl.CITY = sq.NewStringField("city", tbl.TableInfo)
	tbl.COUNTRY = sq.NewStringField("country", tbl.TableInfo)
	tbl.SID = sq.NewNumberField("sid", tbl.TableInfo)
	return tbl
}

type STAFF_LIST struct {
	sq.TableInfo
	ID       sq.NumberField
	NAME     sq.StringField
	ADDRESS  sq.StringField
	ZIP_CODE sq.StringField
	PHONE    sq.StringField
	CITY     sq.StringField
	COUNTRY  sq.StringField
	SID      sq.NumberField
}

func (view STAFF_LIST) DDL(dialect string, v *V) {
	STAFF := NEW_STAFF(dialect, "s")
	ADDRESS := NEW_ADDRESS(dialect, "a")
	CITY := NEW_CITY(dialect, "ci")
	COUNTRY := NEW_COUNTRY(dialect, "co")
	v.AsQuery(sq.SQLite.
		From(STAFF).
		Join(ADDRESS, ADDRESS.ADDRESS_ID.Eq(STAFF.ADDRESS_ID)).
		Join(CITY, CITY.CITY_ID.Eq(ADDRESS.CITY_ID)).
		Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
		Select(
			STAFF.STAFF_ID.As("id"),
			sq.FieldfDialect(map[string]string{
				"default":       "{} || ' ' || {}",
				sq.DialectMySQL: "CONCAT({}, ' ', {})",
			}, STAFF.FIRST_NAME, STAFF.LAST_NAME).As("name"),
			ADDRESS.ADDRESS,
			ADDRESS.POSTAL_CODE.As("zip code"),
			ADDRESS.PHONE,
			CITY.CITY,
			COUNTRY.COUNTRY,
			STAFF.STORE_ID.As("sid"),
		),
	)
}
