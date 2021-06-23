package sq

type ACTOR struct {
	TableInfo
	ACTOR_ID           NumberField
	FIRST_NAME         StringField
	LAST_NAME          StringField
	FULL_NAME          StringField
	FULL_NAME_REVERSED StringField
	LAST_UPDATE        TimeField
}

func NEW_ACTOR(alias string) ACTOR {
	var tbl ACTOR
	tbl.TableInfo = TableInfo{TableName: "actor", TableAlias: alias}
	tbl.ACTOR_ID = NewNumberField("actor_id", tbl.TableInfo)
	tbl.FIRST_NAME = NewStringField("first_name", tbl.TableInfo)
	tbl.LAST_NAME = NewStringField("last_name", tbl.TableInfo)
	tbl.FULL_NAME = NewStringField("full_name", tbl.TableInfo)
	tbl.FULL_NAME_REVERSED = NewStringField("full_name_reversed", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type CATEGORY struct {
	TableInfo
	CATEGORY_ID NumberField
	NAME        StringField
	LAST_UPDATE TimeField
}

func NEW_CATEGORY(alias string) CATEGORY {
	var tbl CATEGORY
	tbl.TableInfo = TableInfo{TableName: "category", TableAlias: alias}
	tbl.CATEGORY_ID = NewNumberField("category_id", tbl.TableInfo)
	tbl.NAME = NewStringField("name", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type COUNTRY struct {
	TableInfo
	COUNTRY_ID  NumberField
	COUNTRY     StringField
	LAST_UPDATE TimeField
}

func NEW_COUNTRY(alias string) COUNTRY {
	var tbl COUNTRY
	tbl.TableInfo = TableInfo{TableName: "country", TableAlias: alias}
	tbl.COUNTRY_ID = NewNumberField("country_id", tbl.TableInfo)
	tbl.COUNTRY = NewStringField("country", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type CITY struct {
	TableInfo
	CITY_ID     NumberField
	CITY        StringField
	COUNTRY_ID  NumberField
	LAST_UPDATE TimeField
}

func NEW_CITY(alias string) CITY {
	var tbl CITY
	tbl.TableInfo = TableInfo{TableName: "city", TableAlias: alias}
	tbl.CITY_ID = NewNumberField("city_id", tbl.TableInfo)
	tbl.CITY = NewStringField("city", tbl.TableInfo)
	tbl.COUNTRY_ID = NewNumberField("country_id", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type ADDRESS struct {
	TableInfo
	ADDRESS_ID  NumberField
	ADDRESS     StringField
	ADDRESS2    StringField
	DISTRICT    StringField
	CITY_ID     NumberField
	POSTAL_CODE StringField
	PHONE       StringField
	LAST_UPDATE TimeField
}

func NEW_ADDRESS(alias string) ADDRESS {
	var tbl ADDRESS
	tbl.TableInfo = TableInfo{TableName: "address", TableAlias: alias}
	tbl.ADDRESS_ID = NewNumberField("address_id", tbl.TableInfo)
	tbl.ADDRESS = NewStringField("address", tbl.TableInfo)
	tbl.ADDRESS2 = NewStringField("address2", tbl.TableInfo)
	tbl.DISTRICT = NewStringField("district", tbl.TableInfo)
	tbl.CITY_ID = NewNumberField("city_id", tbl.TableInfo)
	tbl.POSTAL_CODE = NewStringField("postal_code", tbl.TableInfo)
	tbl.PHONE = NewStringField("phone", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type LANGUAGE struct {
	TableInfo
	LANGUAGE_ID NumberField
	NAME        StringField
	LAST_UPDATE TimeField
}

func NEW_LANGUAGE(alias string) LANGUAGE {
	var tbl LANGUAGE
	tbl.TableInfo = TableInfo{TableName: "lanugage", TableAlias: alias}
	tbl.LANGUAGE_ID = NewNumberField("language_id", tbl.TableInfo)
	tbl.NAME = NewStringField("name", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type FILM struct {
	TableInfo
	FILM_ID              NumberField
	TITLE                StringField
	DESCRIPTION          StringField
	RELEASE_YEAR         NumberField
	LANGUAGE_ID          NumberField
	ORIGINAL_LANGUAGE_ID NumberField
	RENTAL_DURATION      NumberField
	RENTAL_RATE          NumberField
	LENGTH               NumberField
	REPLACEMENT_COST     NumberField
	RATING               StringField
	SPECIAL_FEATURES     CustomField
	LAST_UPDATE          TimeField
	FULLTEXT             StringField
}

func NEW_FILM(alias string) FILM {
	var tbl FILM
	tbl.TableInfo = TableInfo{TableName: "film", TableAlias: alias}
	tbl.FILM_ID = NewNumberField("film_id", tbl.TableInfo)
	tbl.TITLE = NewStringField("title", tbl.TableInfo)
	tbl.DESCRIPTION = NewStringField("description", tbl.TableInfo)
	tbl.RELEASE_YEAR = NewNumberField("release_year", tbl.TableInfo)
	tbl.LANGUAGE_ID = NewNumberField("language_id", tbl.TableInfo)
	tbl.ORIGINAL_LANGUAGE_ID = NewNumberField("original_language_id", tbl.TableInfo)
	tbl.RENTAL_DURATION = NewNumberField("rental_duration", tbl.TableInfo)
	tbl.RENTAL_RATE = NewNumberField("rental_rate", tbl.TableInfo)
	tbl.LENGTH = NewNumberField("length", tbl.TableInfo)
	tbl.REPLACEMENT_COST = NewNumberField("replacement_cost", tbl.TableInfo)
	tbl.RATING = NewStringField("rating", tbl.TableInfo)
	tbl.SPECIAL_FEATURES = NewCustomField("special_features", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	tbl.FULLTEXT = NewStringField("fulltext", tbl.TableInfo)
	return tbl
}

type FILM_TEXT struct {
	TableInfo
	FILM_ID     NumberField
	TITLE       StringField
	DESCRIPTION StringField
}

func NEW_FILM_TEXT(alias string) FILM_TEXT {
	var tbl FILM_TEXT
	tbl.TableInfo = TableInfo{TableName: "film_text", TableAlias: alias}
	tbl.FILM_ID = NewNumberField("film_id", tbl.TableInfo)
	tbl.TITLE = NewStringField("title", tbl.TableInfo)
	tbl.DESCRIPTION = NewStringField("description", tbl.TableInfo)
	return tbl
}

type FILM_ACTOR struct {
	TableInfo
	ACTOR_ID    NumberField
	FILM_ID     NumberField
	LAST_UPDATE TimeField
}

func NEW_FILM_ACTOR(alias string) FILM_ACTOR {
	var tbl FILM_ACTOR
	tbl.TableInfo = TableInfo{TableName: "film_actor", TableAlias: alias}
	tbl.ACTOR_ID = NewNumberField("actor_id", tbl.TableInfo)
	tbl.FILM_ID = NewNumberField("film_id", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type FILM_CATEGORY struct {
	TableInfo
	FILM_ID     NumberField
	CATEGORY_ID NumberField
	LAST_UPDATE TimeField
}

func NEW_FILM_CATEGORY(alias string) FILM_CATEGORY {
	var tbl FILM_CATEGORY
	tbl.TableInfo = TableInfo{TableName: "film_category", TableAlias: alias}
	tbl.FILM_ID = NewNumberField("film_id", tbl.TableInfo)
	tbl.CATEGORY_ID = NewNumberField("category_id", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type STAFF struct {
	TableInfo
	STAFF_ID    NumberField
	FIRST_NAME  StringField
	LAST_NAME   StringField
	ADDRESS_ID  NumberField
	EMAIL       StringField
	STORE_ID    NumberField
	ACTIVE      BooleanField
	USERNAME    StringField
	PASSWORD    StringField
	LAST_UPDATE TimeField
	PICTURE     BlobField
}

func NEW_STAFF(alias string) STAFF {
	var tbl STAFF
	tbl.TableInfo = TableInfo{TableName: "staff", TableAlias: alias}
	tbl.STAFF_ID = NewNumberField("staff_id", tbl.TableInfo)
	tbl.FIRST_NAME = NewStringField("first_name", tbl.TableInfo)
	tbl.LAST_NAME = NewStringField("last_name", tbl.TableInfo)
	tbl.ADDRESS_ID = NewNumberField("address_id", tbl.TableInfo)
	tbl.EMAIL = NewStringField("email", tbl.TableInfo)
	tbl.STORE_ID = NewNumberField("store_id", tbl.TableInfo)
	tbl.ACTIVE = NewBooleanField("active", tbl.TableInfo)
	tbl.USERNAME = NewStringField("username", tbl.TableInfo)
	tbl.PASSWORD = NewStringField("password", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	tbl.PICTURE = NewBlobField("picture", tbl.TableInfo)
	return tbl
}

type STORE struct {
	TableInfo
	STORE_ID         NumberField
	MANAGER_STAFF_ID NumberField
	ADDRESS_ID       NumberField
	LAST_UPDATE      TimeField
}

func NEW_STORE(alias string) STORE {
	var tbl STORE
	tbl.TableInfo = TableInfo{TableName: "store", TableAlias: alias}
	tbl.STORE_ID = NewNumberField("store_id", tbl.TableInfo)
	tbl.MANAGER_STAFF_ID = NewNumberField("manager_staff_id", tbl.TableInfo)
	tbl.ADDRESS_ID = NewNumberField("address_id", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type CUSTOMER struct {
	TableInfo
	CUSTOMER_ID NumberField
	STORE_ID    NumberField
	FIRST_NAME  StringField
	LAST_NAME   StringField
	EMAIL       StringField
	ADDRESS_ID  NumberField
	ACTIVE      BooleanField
	DATA        JSONField
	CREATE_DATE TimeField
	LAST_UPDATE TimeField
}

func NEW_CUSTOMER(alias string) CUSTOMER {
	var tbl CUSTOMER
	tbl.TableInfo = TableInfo{TableName: "customer", TableAlias: alias}
	tbl.CUSTOMER_ID = NewNumberField("customer_id", tbl.TableInfo)
	tbl.STORE_ID = NewNumberField("store_id", tbl.TableInfo)
	tbl.FIRST_NAME = NewStringField("first_name", tbl.TableInfo)
	tbl.LAST_NAME = NewStringField("last_name", tbl.TableInfo)
	tbl.EMAIL = NewStringField("email", tbl.TableInfo)
	tbl.ADDRESS_ID = NewNumberField("address_id", tbl.TableInfo)
	tbl.ACTIVE = NewBooleanField("active", tbl.TableInfo)
	tbl.DATA = NewJSONField("data", tbl.TableInfo)
	tbl.CREATE_DATE = NewTimeField("create_date", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type INVENTORY struct {
	TableInfo
	INVENTORY_ID NumberField
	FILM_ID      NumberField
	STORE_ID     NumberField
	LAST_UPDATE  TimeField
}

func NEW_INVENTORY(alias string) INVENTORY {
	var tbl INVENTORY
	tbl.TableInfo = TableInfo{TableName: "inventory", TableAlias: alias}
	tbl.INVENTORY_ID = NewNumberField("inventory_id", tbl.TableInfo)
	tbl.FILM_ID = NewNumberField("film_id", tbl.TableInfo)
	tbl.STORE_ID = NewNumberField("store_id", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type RENTAL struct {
	TableInfo
	RENTAL_ID    NumberField
	RENTAL_DATE  TimeField
	INVENTORY_ID NumberField
	CUSTOMER_ID  NumberField
	RETURN_DATE  TimeField
	STAFF_ID     NumberField
	LAST_UPDATE  TimeField
}

func NEW_RENTAL(alias string) RENTAL {
	var tbl RENTAL
	tbl.TableInfo = TableInfo{TableName: "rental", TableAlias: alias}
	tbl.RENTAL_ID = NewNumberField("rental_id", tbl.TableInfo)
	tbl.RENTAL_DATE = NewTimeField("rental_date", tbl.TableInfo)
	tbl.INVENTORY_ID = NewNumberField("inventory_id", tbl.TableInfo)
	tbl.CUSTOMER_ID = NewNumberField("customer_id", tbl.TableInfo)
	tbl.RETURN_DATE = NewTimeField("return_date", tbl.TableInfo)
	tbl.STAFF_ID = NewNumberField("staff_id", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type PAYMENT struct {
	TableInfo
	PAYMENT_ID   NumberField
	CUSTOMER_ID  NumberField
	STAFF_ID     NumberField
	RENTAL_ID    NumberField
	AMOUNT       NumberField
	PAYMENT_DATE TimeField
}

func NEW_PAYMENT(alias string) PAYMENT {
	var tbl PAYMENT
	tbl.TableInfo = TableInfo{TableName: "payment", TableAlias: alias}
	tbl.PAYMENT_ID = NewNumberField("payment_id", tbl.TableInfo)
	tbl.CUSTOMER_ID = NewNumberField("customer_id", tbl.TableInfo)
	tbl.STAFF_ID = NewNumberField("staff_id", tbl.TableInfo)
	tbl.RENTAL_ID = NewNumberField("rental_id", tbl.TableInfo)
	tbl.AMOUNT = NewNumberField("amount", tbl.TableInfo)
	tbl.PAYMENT_DATE = NewTimeField("payment_date", tbl.TableInfo)
	return tbl
}
