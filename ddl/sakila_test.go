package ddl

import (
	"github.com/bokwoon95/sq"
)

type ACTOR struct {
	sq.GenericTable
	ACTOR_ID           sq.NumberField `ddl:"type=INTEGER primarykey"`
	FIRST_NAME         sq.StringField `ddl:"notnull"`
	LAST_NAME          sq.StringField `ddl:"notnull index"`
	FULL_NAME          sq.StringField `ddl:"generated={first_name || ' ' || last_name} virtual"`
	FULL_NAME_REVERSED sq.StringField `ddl:"generated={last_name || ' ' || first_name} stored"`
	LAST_UPDATE        sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl ACTOR) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.ACTOR_ID).Identity()
		t.Column(tbl.FULL_NAME_REVERSED).Generated("{} || ' ' || {}", tbl.FIRST_NAME, tbl.LAST_NAME).Stored()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
	case sq.DialectMySQL:
		t.Column(tbl.ACTOR_ID).Autoincrement()
		t.Column(tbl.FIRST_NAME).Type("VARCHAR(45)")
		t.Column(tbl.LAST_NAME).Type("VARCHAR(45)")
		t.Column(tbl.FULL_NAME).Type("VARCHAR(45)").Generated("CONCAT({}, ' ', {})", tbl.FIRST_NAME, tbl.LAST_NAME)
		t.Column(tbl.FULL_NAME_REVERSED).Config(func(c *Column) {
			c.ColumnType = "VARCHAR(45)"
			c.GeneratedExpr = t.Sprintf("CONCAT({}, ' ', {})", tbl.FIRST_NAME, tbl.LAST_NAME)
			c.GeneratedExprStored = true
		})
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_ACTOR(dialect, alias string) ACTOR {
	var tbl ACTOR
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type CATEGORY struct {
	sq.GenericTable
	CATEGORY_ID sq.NumberField `ddl:"type=INTEGER primarykey"`
	NAME        sq.StringField `ddl:"notnull"`
	LAST_UPDATE sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl CATEGORY) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.CATEGORY_ID).Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
	case sq.DialectMySQL:
		t.Column(tbl.CATEGORY_ID).Autoincrement()
		t.Column(tbl.NAME).Type("VARCHAR(25)")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_CATEGORY(dialect, alias string) CATEGORY {
	var tbl CATEGORY
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type COUNTRY struct {
	sq.GenericTable `sq:"name=country"`
	COUNTRY_ID      sq.NumberField `ddl:"type=INTEGER primarykey"`
	COUNTRY         sq.StringField `ddl:"notnull"`
	LAST_UPDATE     sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl COUNTRY) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.COUNTRY_ID).Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
	case sq.DialectMySQL:
		t.Column(tbl.COUNTRY_ID).Autoincrement()
		t.Column(tbl.COUNTRY).Type("VARCHAR(50)")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_COUNTRY(dialect, alias string) COUNTRY {
	var tbl COUNTRY
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type CITY struct {
	sq.GenericTable `sq:"name=city"`
	CITY_ID         sq.NumberField `ddl:"type=INTEGER primarykey"`
	CITY            sq.StringField `ddl:"notnull"`
	COUNTRY_ID      sq.NumberField `ddl:"notnull references={country onupdate=cascade ondelete=restrict} index"`
	LAST_UPDATE     sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl CITY) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.CITY_ID).Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
	case sq.DialectMySQL:
		t.Column(tbl.CITY_ID).Autoincrement()
		t.Column(tbl.CITY).Type("VARCHAR(50)")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_CITY(dialect, alias string) CITY {
	var tbl CITY
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type ADDRESS struct {
	sq.GenericTable `sq:"name=address"`
	ADDRESS_ID      sq.NumberField `ddl:"type=INTEGER primarykey"`
	ADDRESS         sq.StringField `ddl:"notnull"`
	ADDRESS2        sq.StringField
	DISTRICT        sq.StringField `ddl:"notnull"`
	CITY_ID         sq.NumberField `ddl:"notnull references={city onupdate=cascade ondelete=restrict} index"`
	POSTAL_CODE     sq.StringField
	PHONE           sq.StringField `ddl:"notnull"`
	LAST_UPDATE     sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl ADDRESS) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.ADDRESS_ID).Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
	case sq.DialectMySQL:
		t.Column(tbl.ADDRESS_ID).Autoincrement()
		t.Column(tbl.ADDRESS).Type("VARCHAR(50)")
		t.Column(tbl.ADDRESS2).Type("VARCHAR(50)")
		t.Column(tbl.DISTRICT).Type("VARCHAR(20)")
		t.Column(tbl.POSTAL_CODE).Type("VARCHAR(10)")
		t.Column(tbl.PHONE).Type("VARCHAR(20)")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_ADDRESS(dialect, alias string) ADDRESS {
	var tbl ADDRESS
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type LANGUAGE struct {
	sq.GenericTable `sq:"name=language"`
	LANGUAGE_ID     sq.NumberField `ddl:"type=INTEGER primarykey"`
	NAME            sq.StringField `ddl:"notnull"`
	LAST_UPDATE     sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl LANGUAGE) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.LANGUAGE_ID).Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
	case sq.DialectMySQL:
		t.Column(tbl.LANGUAGE_ID).Autoincrement()
		t.Column(tbl.NAME).Type("CHAR(20)")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_LANGUAGE(dialect, alias string) LANGUAGE {
	var tbl LANGUAGE
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type FILM struct {
	sq.GenericTable      `sq:"name=film"`
	FILM_ID              sq.NumberField `ddl:"type=INTEGER primarykey"`
	TITLE                sq.StringField `ddl:"notnull index"`
	DESCRIPTION          sq.StringField
	RELEASE_YEAR         sq.NumberField
	LANGUAGE_ID          sq.NumberField `ddl:"notnull references={language onupdate=cascade ondelete=restrict} index"`
	ORIGINAL_LANGUAGE_ID sq.NumberField `ddl:"references={language onupdate=cascade ondelete=restrict} index"`
	RENTAL_DURATION      sq.NumberField `ddl:"default=3 notnull"`
	RENTAL_RATE          sq.NumberField `ddl:"type=DECIMAL(4,2) default=4.99 notnull"`
	LENGTH               sq.NumberField
	REPLACEMENT_COST     sq.NumberField `ddl:"type=DECIMAL(5,2) default=19.99 notnull"`
	RATING               sq.StringField `ddl:"default='G'"`
	SPECIAL_FEATURES     sq.JSONField
	LAST_UPDATE          sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
	FULLTEXT             sq.StringField `ddl:"notnull"`
}

func (tbl FILM) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.FILM_ID).Identity()
		t.Column(tbl.RELEASE_YEAR).Type("year")
		t.Column(tbl.RATING).Type("mpaa_rating").Default("'G'::mpaa_rating")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Column(tbl.SPECIAL_FEATURES).Type("TEXT[]") // TODO: ArrayField
		t.Column(tbl.FULLTEXT).Type("TSVECTOR")
	case sq.DialectMySQL:
		t.Column(tbl.FILM_ID).Autoincrement()
		t.Column(tbl.TITLE).Type("VARCHAR(255)")
		t.Column(tbl.DESCRIPTION).Type("TEXT")
		t.Column(tbl.RATING).Type("ENUM('G','PG','PG-13','R','NC-17')")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
		t.Check("film_release_year_check", "{1} >= 1901 AND {1} <= 2155", tbl.RELEASE_YEAR)
	case "sqlite3":
		t.Check("film_release_year_check", "{1} >= 1901 AND {1} <= 2155", tbl.RELEASE_YEAR)
		t.Check("film_rating_check", "{} IN ('G','PG','PG-13','R','NC-17')", tbl.RATING)
	}
}

func NEW_FILM(dialect, alias string) FILM {
	var tbl FILM
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type FILM_TEXT struct {
	sq.GenericTable `sq:"name=film_text" ddl:"fts5={content='film' content_rowid='film_id'}"`
	FILM_ID         sq.NumberField
	TITLE           sq.StringField
	DESCRIPTION     sq.StringField
}

func (tbl FILM_TEXT) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres: // no-op, we will ignore this table if postgres
	case sq.DialectMySQL:
		t.Column(tbl.TITLE).Type("VARCHAR(255)").NotNull()
		t.Index(tbl.TITLE, tbl.DESCRIPTION).Using("FULLTEXT")
	case "sqlite3":
		t.Column(tbl.FILM_ID).Ignore() // Ignore will literally delete the column from t.Table.Columns
		// t.Field(tbl.FILM_ID).Config("notnull references") // don't allow for Config() otherwise it will give possible cause for errors
		// i, _ := t.Table.GetColumn(tbl.FILM_ID.GetName())
		// t.Table.DeleteColumns(i)
	}
}

func NEW_FILM_TEXT(dialect, alias string) FILM_TEXT {
	var tbl FILM_TEXT
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type FILM_ACTOR struct {
	sq.GenericTable `sq:"name=film_actor" ddl:"index={. cols=actor_id,film_id unique}"`
	ACTOR_ID        sq.NumberField `ddl:"notnull references={actor onupdate=cascade ondelete=restrict}"`
	FILM_ID         sq.NumberField `ddl:"notnull references={film onupdate=cascade ondelete=restrict} index"`
	LAST_UPDATE     sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl FILM_ACTOR) DDL(dialect string, t *T) {
	t.Index(tbl.ACTOR_ID, tbl.FILM_ID).Unique()
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
	case sq.DialectMySQL:
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_FILM_ACTOR(dialect, alias string) FILM_ACTOR {
	var tbl FILM_ACTOR
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type FILM_CATEGORY struct {
	sq.GenericTable `sq:"name=film_category"`
	FILM_ID         sq.NumberField `ddl:"notnull references={film onupdate=cascade ondelete=restrict}"`
	CATEGORY_ID     sq.NumberField `ddl:"notnull references={category onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE     sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl FILM_CATEGORY) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
	case sq.DialectMySQL:
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_FILM_CATEGORY(dialect, alias string) FILM_CATEGORY {
	var tbl FILM_CATEGORY
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type STAFF struct {
	sq.GenericTable `sq:"name=staff"`
	STAFF_ID        sq.NumberField `ddl:"type=INTEGER primarykey"`
	FIRST_NAME      sq.StringField `ddl:"notnull"`
	LAST_NAME       sq.StringField `ddl:"notnull"`
	ADDRESS_ID      sq.NumberField `ddl:"notnull references={address onupdate=cascade ondelete=restrict}"`
	EMAIL           sq.StringField
	STORE_ID        sq.NumberField  `ddl:"references=store"`
	ACTIVE          sq.BooleanField `ddl:"default=TRUE notnull"`
	USERNAME        sq.StringField  `ddl:"notnull"`
	PASSWORD        sq.StringField
	LAST_UPDATE     sq.TimeField `ddl:"default=DATETIME('now') notnull"`
	PICTURE         sq.BlobField
}

func (tbl STAFF) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.STAFF_ID).Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Column(tbl.PICTURE).Type("BYTEA")
	case sq.DialectMySQL:
		t.Column(tbl.STAFF_ID).Autoincrement()
		t.Column(tbl.FIRST_NAME).Type("VARCHAR(45)")
		t.Column(tbl.LAST_NAME).Type("VARCHAR(45)")
		t.Column(tbl.EMAIL).Type("VARCHAR(50)")
		t.Column(tbl.USERNAME).Type("VARCHAR(16)")
		t.Column(tbl.PASSWORD).Type("VARCHAR(40)")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_STAFF(dialect, alias string) STAFF {
	var tbl STAFF
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type STORE struct {
	sq.GenericTable  `sq:"name=store"`
	STORE_ID         sq.NumberField `ddl:"type=INTEGER primarykey"`
	MANAGER_STAFF_ID sq.NumberField `ddl:"notnull references={staff onupdate=cascade ondelete=restrict} index={. unique}"`
	ADDRESS_ID       sq.NumberField `ddl:"notnull references={address onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE      sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl STORE) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.STORE_ID).Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
	case sq.DialectMySQL:
		t.Column(tbl.STORE_ID).Autoincrement()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_STORE(dialect, alias string) STORE {
	var tbl STORE
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type CUSTOMER struct {
	sq.GenericTable `sq:"name=customer" ddl:"unique={. cols=email,first_name,last_name}"`
	CUSTOMER_ID     sq.NumberField  `ddl:"type=INTEGER primarykey"`
	STORE_ID        sq.NumberField  `ddl:"notnull index"`
	FIRST_NAME      sq.StringField  `ddl:"notnull"`
	LAST_NAME       sq.StringField  `ddl:"notnull index"`
	EMAIL           sq.StringField  `ddl:"unique"`
	ADDRESS_ID      sq.NumberField  `ddl:"notnull references={address onupdate=cascade ondelete=restrict} index"`
	ACTIVE          sq.BooleanField `ddl:"default=TRUE notnull"`
	DATA            sq.JSONField
	CREATE_DATE     sq.TimeField `ddl:"default=DATETIME('now') notnull"`
	LAST_UPDATE     sq.TimeField `ddl:"default=DATETIME('now')"`
}

func (tbl CUSTOMER) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.CUSTOMER_ID).Identity()
		t.Column(tbl.CREATE_DATE).Type("TIMESTAMPTZ").Default("NOW()")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
	case sq.DialectMySQL:
		t.Column(tbl.CUSTOMER_ID).Autoincrement()
		t.Column(tbl.FIRST_NAME).Type("VARCHAR(45)")
		t.Column(tbl.LAST_NAME).Type("VARCHAR(45)")
		t.Column(tbl.EMAIL).Type("VARCHAR(50)")
		t.Column(tbl.CREATE_DATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_CUSTOMER(dialect, alias string) CUSTOMER {
	var tbl CUSTOMER
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type INVENTORY struct {
	sq.GenericTable `sq:"name=inventory" ddl:"index={. cols=store_id,film_id}"`
	INVENTORY_ID    sq.NumberField `ddl:"type=INTEGER primarykey"`
	FILM_ID         sq.NumberField `ddl:"notnull references={film onupdate=cascade ondelete=restrict}"`
	STORE_ID        sq.NumberField `ddl:"notnull references={store onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE     sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl INVENTORY) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.INVENTORY_ID).Identity()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
	case sq.DialectMySQL:
		t.Column(tbl.INVENTORY_ID).Autoincrement()
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_INVENTORY(dialect, alias string) INVENTORY {
	var tbl INVENTORY
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type RENTAL struct {
	sq.GenericTable `sq:"name=rental" ddl:"index={. cols=rental_date,inventory_id,customer_id unique}"`
	RENTAL_ID       sq.NumberField `ddl:"type=INTEGER primarykey"`
	RENTAL_DATE     sq.TimeField   `ddl:"notnull"`
	INVENTORY_ID    sq.NumberField `ddl:"notnull index references={inventory onupdate=cascade ondelete=restrict}"`
	CUSTOMER_ID     sq.NumberField `ddl:"notnull index references={customer onupdate=cascade ondelete=restrict}"`
	RETURN_DATE     sq.TimeField
	STAFF_ID        sq.NumberField `ddl:"notnull index references={staff onupdate=cascade ondelete=restrict}"`
	LAST_UPDATE     sq.TimeField   `ddl:"default=DATETIME('now') notnull"`
}

func (tbl RENTAL) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.RENTAL_ID).Identity()
		t.Column(tbl.RETURN_DATE).Type("TIMESTAMPTZ")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMPTZ").Default("NOW()")
	case sq.DialectMySQL:
		t.Column(tbl.RENTAL_ID).Autoincrement()
		t.Column(tbl.RETURN_DATE).Type("TIMESTAMP")
		t.Column(tbl.LAST_UPDATE).Type("TIMESTAMP").Default("CURRENT_TIMESTAMP").OnUpdateCurrentTimestamp()
	}
}

func NEW_RENTAL(dialect, alias string) RENTAL {
	var tbl RENTAL
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type PAYMENT struct {
	sq.GenericTable `sq:"name=payment"`
	PAYMENT_ID      sq.NumberField `ddl:"type=INTEGER primarykey"`
	CUSTOMER_ID     sq.NumberField `ddl:"notnull index references={customer onupdate=cascade ondelete=restrict}"`
	STAFF_ID        sq.NumberField `ddl:"notnull index references={staff onupdate=cascade ondelete=restrict}"`
	RENTAL_ID       sq.NumberField `ddl:"references={rental onupdate=cascade ondelete=restrict}"`
	AMOUNT          sq.NumberField `ddl:"type=DECIMAL(5,2) notnull"`
	PAYMENT_DATE    sq.TimeField   `ddl:"notnull"`
}

func (tbl PAYMENT) DDL(dialect string, t *T) {
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.PAYMENT_ID).Identity()
		t.Column(tbl.PAYMENT_DATE).Type("TIMESTAMPTZ")
	case sq.DialectMySQL:
		t.Column(tbl.PAYMENT_ID).Autoincrement()
		t.Column(tbl.PAYMENT_DATE).Type("TIMESTAMP")
	}
}

func NEW_PAYMENT(dialect, alias string) PAYMENT {
	var tbl PAYMENT
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}

type DUMMY_TABLE struct {
	sq.GenericTable `sq:"name=dummy_table" ddl:"primarykey={. cols=id1,id2} unique={. cols=score,color}"`
	ID1             sq.NumberField
	ID2             sq.StringField
	SCORE           sq.NumberField
	COLOR           sq.StringField `ddl:"collate=nocase default='red'"`
	DATA            sq.JSONField
}

func (tbl DUMMY_TABLE) DDL(dialect string, t *T) {
	t.Check("dummy_table_score_positive_check", "{} > 0", tbl.SCORE)
	t.Check("dummy_table_score_id1_greater_than_check", "{} > {}", tbl.SCORE, tbl.ID1)
	t.PrimaryKey(tbl.ID1, tbl.ID2)
	t.Unique(tbl.SCORE, tbl.COLOR)
	switch dialect {
	case sq.DialectPostgres:
		t.Column(tbl.COLOR).Collate("C")
		t.NameIndex("dummy_table_score_color_data_idx",
			tbl.SCORE,
			sq.Fieldf("({}->>{})::INT", tbl.DATA, "age"),
			tbl.COLOR,
		).Where("{} = {}", tbl.COLOR, "red")
	case sq.DialectMySQL:
		t.Column(tbl.COLOR).Type("VARCHAR(50)").Collate("latin_swedish_ci")
		t.NameIndex("dummy_table_score_color_data_idx",
			tbl.SCORE,
			sq.Fieldf("CAST({}->>{} AS SIGNED)", tbl.DATA, "$.age"),
			tbl.COLOR,
		)
	case "sqlite3":
		t.Column(tbl.COLOR).Collate("nocase")
		t.NameIndex("dummy_table_score_color_data_idx",
			tbl.SCORE,
			sq.Fieldf("CAST(JSON_EXTRACT({}, {}) AS INT)", tbl.DATA, "$.age"),
			tbl.COLOR,
		).Where("{} = {}", tbl.COLOR, "red")
	}
}

func NEW_DUMMY_TABLE(dialect, alias string) DUMMY_TABLE {
	var tbl DUMMY_TABLE
	switch dialect {
	case sq.DialectPostgres:
		tbl.GenericTable.TableSchema = "public"
	case sq.DialectMySQL:
		tbl.GenericTable.TableSchema = "db"
	}
	_ = sq.ReflectTable(&tbl)
	tbl.GenericTable.TableAlias = alias
	return tbl
}
