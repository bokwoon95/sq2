package ddl

import (
	"github.com/bokwoon95/sq"
)

const sqliteLastUpdateTriggerFmt = `
CREATE TRIGGER {1} AFTER UPDATE ON {2} BEGIN
    UPDATE {2} SET last_update = DATETIME('now') WHERE ROWID = NEW.ROWID;
END;`

const postgresLastUpdateTriggerFmt = `
CREATE TRIGGER {1} BEFORE UPDATE ON {2}
FOR EACH ROW EXECUTE PROCEDURE last_update_trg();`

type ACTOR struct {
	sq.TableInfo
	ACTOR_ID           sq.NumberField `ddl:"sqlite:type=INTEGER primarykey auto_increment autoincrement identity"`
	FIRST_NAME         sq.StringField `ddl:"mysql:type=VARCHAR(45) notnull"`
	LAST_NAME          sq.StringField `ddl:"mysql:type=VARCHAR(45) notnull index"`
	FULL_NAME          sq.StringField `ddl:"generated={first_name || ' ' || last_name} mysql:generated={CONCAT(first_name, ' ', last_name)} virtual"`
	FULL_NAME_REVERSED sq.StringField `ddl:"generated={last_name || ' ' || first_name} mysql:generated={CONCAT(last_name, ' ', first_name)} stored"`
	LAST_UPDATE        sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
}

func NEW_ACTOR(alias string) ACTOR {
	var tbl ACTOR
	tbl.TableInfo = sq.TableInfo{TableName: "actor", TableAlias: alias}
	tbl.ACTOR_ID = sq.NewNumberField("actor_id", tbl.TableInfo)
	tbl.FIRST_NAME = sq.NewStringField("first_name", tbl.TableInfo)
	tbl.LAST_NAME = sq.NewStringField("last_name", tbl.TableInfo)
	tbl.FULL_NAME = sq.NewStringField("full_name", tbl.TableInfo)
	tbl.FULL_NAME_REVERSED = sq.NewStringField("full_name_reversed", tbl.TableInfo)
	tbl.LAST_UPDATE = sq.NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

func (tbl ACTOR) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("actor_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("actor_last_update_before_update_trg"), tbl)
	}
}

type CATEGORY struct {
	sq.TableInfo
	CATEGORY_ID sq.NumberField `ddl:"sqlite:type=INTEGER primarykey auto_increment identity"`
	NAME        sq.StringField `ddl:"mysql:type=VARCHAR(25) notnull"`
	LAST_UPDATE sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
}

func NEW_CATEGORY(alias string) CATEGORY {
	var tbl CATEGORY
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl CATEGORY) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("category_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("category_last_update_before_update_trg"), tbl)
	}
}

type COUNTRY struct {
	sq.TableInfo
	COUNTRY_ID  sq.NumberField `ddl:"sqlite:type=INTEGER primarykey auto_increment identity"`
	COUNTRY     sq.StringField `ddl:"notnull mysql:type=VARCHAR(50)"`
	LAST_UPDATE sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
}

func NEW_COUNTRY(alias string) COUNTRY {
	var tbl COUNTRY
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl COUNTRY) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("country_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("country_last_update_before_update_trg"), tbl)
	}
}

type CITY struct {
	sq.TableInfo
	CITY_ID     sq.NumberField `ddl:"sqlite:type=INTEGER primarykey auto_increment identity"`
	CITY        sq.StringField `ddl:"notnull mysql:type=VARCHAR(50)"`
	COUNTRY_ID  sq.NumberField `ddl:"notnull references={country.country_id onupdate=cascade ondelete=restrict} index"`
	LAST_UPDATE sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
}

func NEW_CITY(alias string) CITY {
	var tbl CITY
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl CITY) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("city_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("city_last_update_before_update_trg"), tbl)
	}
}

type ADDRESS struct {
	sq.TableInfo
	ADDRESS_ID  sq.NumberField `ddl:"sqlite:type=INTEGER primarykey auto_increment identity"`
	ADDRESS     sq.StringField `ddl:"notnull mysql:type=VARCHAR(50)"`
	ADDRESS2    sq.StringField `ddl:"mysql:type=VARCHAR(50)"`
	DISTRICT    sq.StringField `ddl:"notnull mysql:type=VARCHAR(20)"`
	CITY_ID     sq.NumberField `ddl:"notnull references={city.city_id onupdate=cascade ondelete=restrict} index"`
	POSTAL_CODE sq.StringField `ddl:"mysql:type=VARCHAR(10)"`
	PHONE       sq.StringField `ddl:"notnull mysql:type=VARCHAR(20)"`
	LAST_UPDATE sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
}

func NEW_ADDRESS(alias string) ADDRESS {
	var tbl ADDRESS
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl ADDRESS) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("address_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("city_last_update_before_update_trg"), tbl)
	}
}

type LANGUAGE struct {
	sq.TableInfo
	LANGUAGE_ID sq.NumberField `ddl:"sqlite:type=INTEGER primarykey auto_increment identity"`
	NAME        sq.StringField `ddl:"notnull mysql:type=CHAR(20)"`
	LAST_UPDATE sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
}

func NEW_LANGUAGE(alias string) LANGUAGE {
	var tbl LANGUAGE
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl LANGUAGE) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("language_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("language_last_update_before_update_trg"), tbl)
	}
}

type FILM struct {
	sq.TableInfo
	FILM_ID              sq.NumberField `ddl:"sqlite:type=INTEGER primarykey auto_increment identity"`
	TITLE                sq.StringField `ddl:"notnull index"`
	DESCRIPTION          sq.StringField `ddl:"type=TEXT"`
	RELEASE_YEAR         sq.NumberField
	LANGUAGE_ID          sq.NumberField `ddl:"notnull references={language.language_id onupdate=cascade ondelete=restrict} index"`
	ORIGINAL_LANGUAGE_ID sq.NumberField `ddl:"references={language.language_id onupdate=cascade ondelete=restrict} index"`
	RENTAL_DURATION      sq.NumberField `ddl:"default=3 notnull"`
	RENTAL_RATE          sq.NumberField `ddl:"type=DECIMAL(4,2) default=4.99 notnull"`
	LENGTH               sq.NumberField
	REPLACEMENT_COST     sq.NumberField `ddl:"type=DECIMAL(5,2) default=19.99 notnull"`
	RATING               sq.StringField `ddl:"mysql:type=ENUM('G','PG','PG-13','R','NC-17') default='G'"`
	SPECIAL_FEATURES     sq.CustomField `ddl:"type=JSON postgres:type=TEXT[]"`
	LAST_UPDATE          sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
	FULLTEXT             sq.StringField `ddl:"ignore=mysql,sqlite postgres:type=TSVECTOR postgres:index={. using=GIST}"`
	// TODO: CREATE TYPE mpaa_rating, CREATE DOMAIN year
}

func NEW_FILM(alias string) FILM {
	var tbl FILM
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl FILM) DDL(dialect string, t *T) {
	t.Check("film_release_year_check", "{1} >= 1901 AND {1} <= 2155", tbl.RELEASE_YEAR)
	if dialect == sq.DialectSQLite || dialect == sq.DialectPostgres {
		t.Check("film_rating_check", "{} IN ('G','PG','PG-13','R','NC-17')", tbl.RATING)
	}
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("film_last_update_after_update_trg"), tbl)
		t.Trigger(`
CREATE TRIGGER film_fts5_after_insert_trg AFTER INSERT ON {1} BEGIN
    INSERT INTO film_text (ROWID, title, description) VALUES (NEW.film_id, NEW.title, NEW.description);
END; `, tbl)
		t.Trigger(`
CREATE TRIGGER film_fts5_after_delete_trg AFTER DELETE ON {1} BEGIN
    INSERT INTO film_text (film_text, ROWID, title, description) VALUES ('delete', OLD.film_id, OLD.title, OLD.description);
END;`, tbl)
		t.Trigger(`
CREATE TRIGGER film_fts5_after_update_trg AFTER UPDATE ON {1} BEGIN
    INSERT INTO film_text (film_text, ROWID, title, description) VALUES ('delete', OLD.film_id, OLD.title, OLD.description);
    INSERT INTO film_text (ROWID, title, description) VALUES (NEW.film_id, NEW.title, NEW.description);
END;`, tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("film_last_update_before_update_trg"), tbl)
		t.Trigger(`
CREATE TRIGGER film_fulltext_before_insert_update_trg BEFORE INSERT OR UPDATE ON {1}
FOR EACH ROW EXECUTE PROCEDURE tsvector_update_trigger(fulltext, 'pg_catalog.english', title, description);`, tbl)
	}
	if dialect == sq.DialectMySQL {
		t.Trigger(`
CREATE TRIGGER film_after_insert_trg AFTER INSERT ON film FOR EACH ROW BEGIN
    INSERT INTO film_text (film_id, title, description) VALUES (NEW.film_id, NEW.title, NEW.description);
END`)
		t.Trigger(`
CREATE TRIGGER film_after_update_trg AFTER UPDATE ON film FOR EACH ROW BEGIN
    IF OLD.title <> NEW.title OR OLD.description <> NEW.description THEN
        UPDATE film_text
        SET title = NEW.title, description = NEW.description, film_id = NEW.film_id
        WHERE film_id = OLD.film_id;
    END IF;
END`)
		t.Trigger(`
CREATE TRIGGER film_after_delete_trg AFTER DELETE ON film FOR EACH ROW BEGIN
    DELETE FROM film_text WHERE film_id = OLD.film_id;
END`)
	}
}

type FILM_TEXT struct {
	sq.TableInfo `ddl:"ignore=postgres virtual={fts5 content='film' content_rowid='film_id'} mysql:index={. cols=title,description using=FULLTEXT}"`
	FILM_ID      sq.NumberField `ddl:"ignore=sqlite notnull primarykey"`
	TITLE        sq.StringField `ddl:"mysql:type=VARCHAR(255)"`
	DESCRIPTION  sq.StringField `ddl:"mysql:type=TEXT"`
}

func NEW_FILM_TEXT(alias string) FILM_TEXT {
	var tbl FILM_TEXT
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

type FILM_ACTOR struct {
	sq.TableInfo `ddl:"index={. cols=actor_id,film_id unique}"`
	FILM_ID      sq.NumberField `ddl:"notnull references={film.film_id onupdate=cascade ondelete=restrict} index"`
	ACTOR_ID     sq.NumberField `ddl:"notnull references={actor.actor_id onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE  sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
}

func NEW_FILM_ACTOR(alias string) FILM_ACTOR {
	var tbl FILM_ACTOR
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl FILM_ACTOR) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("film_actor_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("film_actor_last_update_before_update_trg"), tbl)
	}
}

type FILM_ACTOR_REVIEW struct {
	sq.TableInfo
	FILM_ID      sq.NumberField `ddl:"type=INT"`
	ACTOR_ID     sq.NumberField `ddl:"type=INT"`
	REVIEW_TITLE sq.StringField `ddl:"mysql:type=VARCHAR(50) notnull default='' sqlite:collate=nocase postgres:collate=C mysql:collate=latin1_swedish_ci"`
	REVIEW_BODY  sq.StringField `ddl:"notnull default=''"`
	METADATA     sq.JSONField
	LAST_UPDATE  sq.TimeField `ddl:"notnull default=CURRENT_TIMESTAMP postgres:default=NOW() sqlite:default=DATETIME('now') onupdatecurrenttimestamp"`
	DELETE_DATE  sq.TimeField
}

func NEW_FILM_ACTOR_REVIEW(alias string) FILM_ACTOR_REVIEW {
	var tbl FILM_ACTOR_REVIEW
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl FILM_ACTOR_REVIEW) DDL(dialect string, t *T) {
	FILM_ACTOR := NEW_FILM_ACTOR("")
	t.PrimaryKey(tbl.FILM_ID, tbl.ACTOR_ID)
	t.ForeignKey(tbl.FILM_ID, tbl.ACTOR_ID).References(FILM_ACTOR, FILM_ACTOR.FILM_ID, FILM_ACTOR.ACTOR_ID).OnUpdate(CASCADE).Deferrable().InitiallyDeferred()
	t.Check("film_actor_review_check", "LENGTH({}) > LENGTH({})", tbl.REVIEW_BODY, tbl.REVIEW_TITLE)
	if dialect == sq.DialectSQLite {
		t.NameIndex("film_actor_review_misc",
			tbl.FILM_ID,
			sq.Fieldf("SUBSTR({}, 2, 10)", tbl.REVIEW_BODY),
			sq.Fieldf("{} || {}", tbl.REVIEW_TITLE, " abcd"),
			sq.Fieldf("CAST(JSON_EXTRACT({}, {}) AS INT)", tbl.METADATA, "$.score"),
		).Where("{} IS NULL", tbl.DELETE_DATE)
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("film_actor_review_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.NameIndex("film_actor_review_review_title_idx", sq.Literal("review_title text_pattern_ops"))
		t.NameIndex("film_actor_review_review_body_idx", sq.Literal(`review_body COLLATE "C"`))
		t.NameIndex("film_actor_review_misc",
			tbl.FILM_ID,
			sq.Fieldf("SUBSTR({}, 2, 10)", tbl.REVIEW_BODY),
			sq.Fieldf("{} || {}", tbl.REVIEW_TITLE, " abcd"),
			sq.Fieldf("({}->>{})::INT", tbl.METADATA, "score"),
		).Include(tbl.ACTOR_ID, tbl.LAST_UPDATE).Where("{} IS NULL", tbl.DELETE_DATE)
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("film_actor_review_last_update_before_update_trg"), tbl)
	}
	if dialect == sq.DialectMySQL {
		t.NameIndex("film_actor_review_misc",
			tbl.FILM_ID,
			sq.Fieldf("SUBSTR({}, 2, 10)", tbl.REVIEW_BODY),
			sq.Fieldf("CONCAT({}, {})", tbl.REVIEW_TITLE, " abcd"),
			sq.Fieldf("CAST({}->>{} AS SIGNED)", tbl.METADATA, "$.score"),
		)
	}
}

type FILM_CATEGORY struct {
	sq.TableInfo
	FILM_ID     sq.NumberField `ddl:"notnull references={film.film_id onupdate=cascade ondelete=restrict}"`
	CATEGORY_ID sq.NumberField `ddl:"notnull references={category.category_id onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
}

func NEW_FILM_CATEGORY(alias string) FILM_CATEGORY {
	var tbl FILM_CATEGORY
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl FILM_CATEGORY) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("film_category_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("film_category_last_update_before_update_trg"), tbl)
	}
}

type STAFF struct {
	sq.TableInfo
	STAFF_ID    sq.NumberField  `ddl:"sqlite:type=INTEGER primarykey auto_increment identity"`
	FIRST_NAME  sq.StringField  `ddl:"mysql:type=VARCHAR(45) notnull"`
	LAST_NAME   sq.StringField  `ddl:"mysql:type=VARCHAR(45) notnull"`
	ADDRESS_ID  sq.NumberField  `ddl:"notnull references={address.address_id onupdate=cascade ondelete=restrict}"`
	EMAIL       sq.StringField  `ddl:"mysql:type=VARCHAR(50)"`
	STORE_ID    sq.NumberField  `ddl:"references=store.store_id"`
	ACTIVE      sq.BooleanField `ddl:"default=TRUE notnull"`
	USERNAME    sq.StringField  `ddl:"mysql:type=VARCHAR(16) notnull"`
	PASSWORD    sq.StringField  `ddl:"mysql:type=VARCHAR(40)"`
	LAST_UPDATE sq.TimeField    `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
	PICTURE     sq.BlobField
}

func NEW_STAFF(alias string) STAFF {
	var tbl STAFF
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl STAFF) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("staff_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("staff_last_update_before_update_trg"), tbl)
	}
}

type STORE struct {
	sq.TableInfo
	STORE_ID         sq.NumberField `ddl:"sqlite:type=INTEGER primarykey auto_increment identity"`
	MANAGER_STAFF_ID sq.NumberField `ddl:"notnull references={staff.staff_id onupdate=cascade ondelete=restrict} index={. unique}"`
	ADDRESS_ID       sq.NumberField `ddl:"notnull references={address.address_id onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE      sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
}

func NEW_STORE(alias string) STORE {
	var tbl STORE
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl STORE) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("store_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("store_last_update_before_update_trg"), tbl)
	}
}

type CUSTOMER struct {
	sq.TableInfo `ddl:"unique={. cols=email,first_name,last_name}"`
	CUSTOMER_ID  sq.NumberField  `ddl:"sqlite:type=INTEGER primarykey auto_increment identity"`
	STORE_ID     sq.NumberField  `ddl:"notnull index"`
	FIRST_NAME   sq.StringField  `ddl:"mysql:type=VARCHAR(45) notnull"`
	LAST_NAME    sq.StringField  `ddl:"mysql:type=VARCHAR(45) notnull index"`
	EMAIL        sq.StringField  `ddl:"mysql:type=VARCHAR(50) unique"`
	ADDRESS_ID   sq.NumberField  `ddl:"notnull references={address.address_id onupdate=cascade ondelete=restrict} index"`
	ACTIVE       sq.BooleanField `ddl:"default=TRUE notnull"`
	DATA         sq.JSONField
	CREATE_DATE  sq.TimeField `ddl:"notnull default=CURRENT_TIMESTAMP"`
	LAST_UPDATE  sq.TimeField `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
}

func NEW_CUSTOMER(alias string) CUSTOMER {
	var tbl CUSTOMER
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl CUSTOMER) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("customer_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("customer_last_update_before_update_trg"), tbl)
	}
}

type INVENTORY struct {
	sq.TableInfo `ddl:"index={. cols=store_id,film_id}"`
	INVENTORY_ID sq.NumberField `ddl:"sqlite:type=INTEGER primarykey auto_increment identity"`
	FILM_ID      sq.NumberField `ddl:"notnull references={film.film_id onupdate=cascade ondelete=restrict}"`
	STORE_ID     sq.NumberField `ddl:"notnull references={store.store_id onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE  sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
}

func NEW_INVENTORY(alias string) INVENTORY {
	var tbl INVENTORY
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl INVENTORY) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("inventory_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("inventory_last_update_before_update_trg"), tbl)
	}
}

type RENTAL struct {
	sq.TableInfo `ddl:"index={. cols=rental_date,inventory_id,customer_id unique}"`
	RENTAL_ID    sq.NumberField `ddl:"sqlite:type=INTEGER primarykey auto_increment identity"`
	RENTAL_DATE  sq.TimeField   `ddl:"notnull"`
	INVENTORY_ID sq.NumberField `ddl:"notnull index references={inventory.inventory_id onupdate=cascade ondelete=restrict}"`
	CUSTOMER_ID  sq.NumberField `ddl:"notnull index references={customer.customer_id onupdate=cascade ondelete=restrict}"`
	RETURN_DATE  sq.TimeField
	STAFF_ID     sq.NumberField `ddl:"notnull index references={staff.staff_id onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE  sq.TimeField   `ddl:"notnull default=CURRENT_TIMESTAMP onupdatecurrenttimestamp"`
}

func NEW_RENTAL(alias string) RENTAL {
	var tbl RENTAL
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (tbl RENTAL) DDL(dialect string, t *T) {
	if dialect == sq.DialectSQLite {
		t.Trigger(sqliteLastUpdateTriggerFmt, sq.Literal("rental_last_update_after_update_trg"), tbl)
	}
	if dialect == sq.DialectPostgres {
		t.NameExclude("rental_range_excl", "GIST", Exclusions{
			{tbl.INVENTORY_ID, "="},
			{sq.Fieldf("tstzrange({}, {}, '[]')", tbl.RENTAL_DATE, tbl.RETURN_DATE), "&&"},
		})
		t.Trigger(postgresLastUpdateTriggerFmt, sq.Literal("rental_last_update_before_update_trg"), tbl)
	}
}

type PAYMENT struct {
	sq.TableInfo
	PAYMENT_ID   sq.NumberField `ddl:"sqlite:type=INTEGER primarykey auto_increment identity"`
	CUSTOMER_ID  sq.NumberField `ddl:"notnull index references={customer.customer_id onupdate=cascade ondelete=restrict}"`
	STAFF_ID     sq.NumberField `ddl:"notnull index references={staff.staff_id onupdate=cascade ondelete=restrict}"`
	RENTAL_ID    sq.NumberField `ddl:"references={rental.rental_id onupdate=cascade ondelete=restrict}"`
	AMOUNT       sq.NumberField `ddl:"type=DECIMAL(5,2) notnull"`
	PAYMENT_DATE sq.TimeField   `ddl:"notnull"`
}

func NEW_PAYMENT(alias string) PAYMENT {
	var tbl PAYMENT
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
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
		sq.DialectMySQL:  "json_arrayagg({})",
	}, value)
}

type ACTOR_INFO struct {
	sq.TableInfo
	ACTOR_ID   sq.NumberField
	FIRST_NAME sq.StringField
	LAST_NAME  sq.StringField
	FILM_INFO  sq.JSONField
}

func NEW_ACTOR_INFO(alias string) ACTOR_INFO {
	var tbl ACTOR_INFO
	tbl.TableInfo = sq.TableInfo{TableName: "actor_info", TableAlias: alias}
	tbl.ACTOR_ID = sq.NewNumberField("actor_id", tbl.TableInfo)
	tbl.FIRST_NAME = sq.NewStringField("first_name", tbl.TableInfo)
	tbl.LAST_NAME = sq.NewStringField("last_name", tbl.TableInfo)
	tbl.FILM_INFO = sq.NewJSONField("film_info", tbl.TableInfo)
	return tbl
}

func (view ACTOR_INFO) DDL(dialect string, v *V) {
	ACTOR := NEW_ACTOR("a")
	FILM := NEW_FILM("f")
	FILM_ACTOR := NEW_FILM_ACTOR("fa")
	FILM_CATEGORY := NEW_FILM_CATEGORY("fc")
	CATEGORY := NEW_CATEGORY("c")
	v.SetQuery(sq.SQLite.
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

func NEW_CUSTOMER_LIST(alias string) CUSTOMER_LIST {
	var tbl CUSTOMER_LIST
	tbl.TableInfo = sq.TableInfo{TableName: "customer_list", TableAlias: alias}
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

func (view CUSTOMER_LIST) DDL(dialect string, v *V) {
	CUSTOMER := NEW_CUSTOMER("cu")
	ADDRESS := NEW_ADDRESS("a")
	CITY := NEW_CITY("")
	COUNTRY := NEW_COUNTRY("")
	v.SetQuery(sq.SQLite.
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

func NEW_FILM_LIST(alias string) FILM_LIST {
	var tbl FILM_LIST
	tbl.TableInfo = sq.TableInfo{TableName: "film_list", TableAlias: alias}
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

func (view FILM_LIST) DDL(dialect string, v *V) {
	CATEGORY := NEW_CATEGORY("")
	FILM_CATEGORY := NEW_FILM_CATEGORY("")
	FILM := NEW_FILM("")
	FILM_ACTOR := NEW_FILM_ACTOR("")
	ACTOR := NEW_ACTOR("")
	v.SetQuery(sq.SQLite.
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

func NEW_NICER_BUT_SLOWER_FILM_LIST(alias string) NICER_BUT_SLOWER_FILM_LIST {
	var tbl NICER_BUT_SLOWER_FILM_LIST
	tbl.TableInfo = sq.TableInfo{TableName: "nicer_but_slower_film_list", TableAlias: alias}
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

func (view NICER_BUT_SLOWER_FILM_LIST) DDL(dialect string, v *V) {
	CATEGORY := NEW_CATEGORY("")
	FILM_CATEGORY := NEW_FILM_CATEGORY("")
	FILM := NEW_FILM("")
	FILM_ACTOR := NEW_FILM_ACTOR("")
	ACTOR := NEW_ACTOR("")
	v.SetQuery(sq.SQLite.
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
				sq.DialectMySQL: "CONCAT(" +
					"UPPER(SUBSTRING({1}, 1, 1))" +
					", LOWER(SUBSTRING({1}, 2))" +
					", ' '" +
					", UPPER(SUBSTRING({2}, 1, 1))" +
					", LOWER(SUBSTRING({2}, 2))" +
					")",
			}, ACTOR.FIRST_NAME, ACTOR.LAST_NAME)).As("actors"),
		),
	)
}

type SALES_BY_FILM_CATEGORY struct {
	sq.TableInfo
	CATEGORY    sq.StringField
	TOTAL_SALES sq.NumberField
}

func NEW_SALES_BY_FILM_CATEGORY(alias string) SALES_BY_FILM_CATEGORY {
	var tbl SALES_BY_FILM_CATEGORY
	tbl.TableInfo = sq.TableInfo{TableName: "sales_by_film_category", TableAlias: alias}
	tbl.CATEGORY = sq.NewStringField("category", tbl.TableInfo)
	tbl.TOTAL_SALES = sq.NewNumberField("total_sales", tbl.TableInfo)
	return tbl
}

func (view SALES_BY_FILM_CATEGORY) DDL(dialect string, v *V) {
	PAYMENT := NEW_PAYMENT("p")
	RENTAL := NEW_RENTAL("r")
	INVENTORY := NEW_INVENTORY("i")
	FILM := NEW_FILM("f")
	FILM_CATEGORY := NEW_FILM_CATEGORY("fc")
	CATEGORY := NEW_CATEGORY("c")
	v.SetQuery(sq.SQLite.
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

type SALES_BY_STORE struct {
	sq.TableInfo
	STORE       sq.StringField
	MANAGER     sq.StringField
	TOTAL_SALES sq.NumberField
}

func NEW_SALES_BY_STORE(alias string) SALES_BY_STORE {
	var tbl SALES_BY_STORE
	tbl.TableInfo = sq.TableInfo{TableName: "sales_by_store", TableAlias: alias}
	tbl.STORE = sq.NewStringField("store", tbl.TableInfo)
	tbl.MANAGER = sq.NewStringField("manager", tbl.TableInfo)
	tbl.TOTAL_SALES = sq.NewNumberField("total_sales", tbl.TableInfo)
	return tbl
}

func (view SALES_BY_STORE) DDL(dialect string, v *V) {
	PAYMENT := NEW_PAYMENT("p")
	RENTAL := NEW_RENTAL("r")
	INVENTORY := NEW_INVENTORY("i")
	STORE := NEW_STORE("s")
	ADDRESS := NEW_ADDRESS("a")
	CITY := NEW_CITY("ci")
	COUNTRY := NEW_COUNTRY("co")
	STAFF := NEW_STAFF("m")
	v.SetQuery(sq.SQLite.
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

func NEW_STAFF_LIST(alias string) STAFF_LIST {
	var tbl STAFF_LIST
	tbl.TableInfo = sq.TableInfo{TableName: "staff_list", TableAlias: alias}
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

func (view STAFF_LIST) DDL(dialect string, v *V) {
	STAFF := NEW_STAFF("s")
	ADDRESS := NEW_ADDRESS("a")
	CITY := NEW_CITY("ci")
	COUNTRY := NEW_COUNTRY("co")
	v.SetQuery(sq.SQLite.
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

type FULL_ADDRESS struct {
	sq.TableInfo
	COUNTRY_ID  sq.NumberField
	CITY_ID     sq.NumberField
	ADDRESS_ID  sq.NumberField
	COUNTRY     sq.StringField
	CITY        sq.StringField
	ADDRESS     sq.StringField
	ADDRESS2    sq.StringField
	DISTRICT    sq.StringField
	POSTAL_CODE sq.StringField
	PHONE       sq.StringField
	LAST_UPDATE sq.TimeField
}

func NEW_FULL_ADDRESS(alias string) FULL_ADDRESS {
	var tbl FULL_ADDRESS
	_ = sq.ReflectTable(&tbl, alias)
	return tbl
}

func (view FULL_ADDRESS) DDL(dialect string, v *V) {
	const triggerFmt = `
CREATE TRIGGER {1} AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE ON {2}
FOR EACH STATEMENT EXECUTE PROCEDURE refresh_full_address();
`
	ADDRESS := NEW_ADDRESS("")
	CITY := NEW_CITY("")
	COUNTRY := NEW_COUNTRY("")
	v.IsMaterialized()
	v.SetQuery(sq.SQLite.
		From(ADDRESS).
		Join(CITY, CITY.CITY_ID.Eq(ADDRESS.CITY_ID)).
		Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
		Select(
			COUNTRY.COUNTRY_ID,
			CITY.CITY_ID,
			ADDRESS.ADDRESS_ID,
			COUNTRY.COUNTRY,
			CITY.CITY,
			ADDRESS.ADDRESS,
			ADDRESS.ADDRESS2,
			ADDRESS.DISTRICT,
			ADDRESS.POSTAL_CODE,
			ADDRESS.PHONE,
			ADDRESS.LAST_UPDATE,
		),
	)
	v.Index(view.COUNTRY_ID, view.CITY_ID, view.ADDRESS_ID).Unique().Include(view.COUNTRY, view.CITY, view.ADDRESS, view.ADDRESS2)
	v.Trigger(triggerFmt, sq.Literal("address_refresh_full_address_trg"), ADDRESS)
	v.Trigger(triggerFmt, sq.Literal("city_refresh_full_address_trg"), CITY)
	v.Trigger(triggerFmt, sq.Literal("country_refresh_full_address_trg"), COUNTRY)
}
