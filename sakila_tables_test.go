package sq

type xACTOR struct {
	TableInfo
	ACTOR_ID           NumberField
	FIRST_NAME         StringField
	LAST_NAME          StringField
	FULL_NAME          StringField
	FULL_NAME_REVERSED StringField
	LAST_UPDATE        TimeField
}

func xNEW_ACTOR(alias string) xACTOR {
	var tbl xACTOR
	tbl.TableInfo = TableInfo{TableName: "actor", TableAlias: alias}
	tbl.ACTOR_ID = NewNumberField("actor_id", tbl.TableInfo)
	tbl.FIRST_NAME = NewStringField("first_name", tbl.TableInfo)
	tbl.LAST_NAME = NewStringField("last_name", tbl.TableInfo)
	tbl.FULL_NAME = NewStringField("full_name", tbl.TableInfo)
	tbl.FULL_NAME_REVERSED = NewStringField("full_name_reversed", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type xCATEGORY struct {
	TableInfo
	CATEGORY_ID NumberField
	NAME        StringField
	LAST_UPDATE TimeField
}

func xNEW_CATEGORY(alias string) xCATEGORY {
	var tbl xCATEGORY
	tbl.TableInfo = TableInfo{TableName: "category", TableAlias: alias}
	tbl.CATEGORY_ID = NewNumberField("category_id", tbl.TableInfo)
	tbl.NAME = NewStringField("name", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type xCOUNTRY struct {
	TableInfo
	COUNTRY_ID  NumberField
	COUNTRY     StringField
	LAST_UPDATE TimeField
}

func xNEW_COUNTRY(alias string) xCOUNTRY {
	var tbl xCOUNTRY
	tbl.TableInfo = TableInfo{TableName: "country", TableAlias: alias}
	tbl.COUNTRY_ID = NewNumberField("country_id", tbl.TableInfo)
	tbl.COUNTRY = NewStringField("country", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type xCITY struct {
	TableInfo
	CITY_ID     NumberField
	CITY        StringField
	COUNTRY_ID  NumberField
	LAST_UPDATE TimeField
}

func xNEW_CITY(alias string) xCITY {
	var tbl xCITY
	tbl.TableInfo = TableInfo{TableName: "city", TableAlias: alias}
	tbl.CITY_ID = NewNumberField("city_id", tbl.TableInfo)
	tbl.CITY = NewStringField("city", tbl.TableInfo)
	tbl.COUNTRY_ID = NewNumberField("country_id", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type xADDRESS struct {
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

func xNEW_ADDRESS(alias string) xADDRESS {
	var tbl xADDRESS
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

type xLANGUAGE struct {
	TableInfo
	LANGUAGE_ID NumberField
	NAME        StringField
	LAST_UPDATE TimeField
}

func xNEW_LANGUAGE(alias string) xLANGUAGE {
	var tbl xLANGUAGE
	tbl.TableInfo = TableInfo{TableName: "language", TableAlias: alias}
	tbl.LANGUAGE_ID = NewNumberField("language_id", tbl.TableInfo)
	tbl.NAME = NewStringField("name", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type xFILM struct {
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

func xNEW_FILM(alias string) xFILM {
	var tbl xFILM
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

type xFILM_TEXT struct {
	TableInfo
	FILM_ID     NumberField
	TITLE       StringField
	DESCRIPTION StringField
}

func xNEW_FILM_TEXT(alias string) xFILM_TEXT {
	var tbl xFILM_TEXT
	tbl.TableInfo = TableInfo{TableName: "film_text", TableAlias: alias}
	tbl.FILM_ID = NewNumberField("film_id", tbl.TableInfo)
	tbl.TITLE = NewStringField("title", tbl.TableInfo)
	tbl.DESCRIPTION = NewStringField("description", tbl.TableInfo)
	return tbl
}

type xFILM_ACTOR struct {
	TableInfo
	ACTOR_ID    NumberField
	FILM_ID     NumberField
	LAST_UPDATE TimeField
}

func xNEW_FILM_ACTOR(alias string) xFILM_ACTOR {
	var tbl xFILM_ACTOR
	tbl.TableInfo = TableInfo{TableName: "film_actor", TableAlias: alias}
	tbl.ACTOR_ID = NewNumberField("actor_id", tbl.TableInfo)
	tbl.FILM_ID = NewNumberField("film_id", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type xFILM_CATEGORY struct {
	TableInfo
	FILM_ID     NumberField
	CATEGORY_ID NumberField
	LAST_UPDATE TimeField
}

func xNEW_FILM_CATEGORY(alias string) xFILM_CATEGORY {
	var tbl xFILM_CATEGORY
	tbl.TableInfo = TableInfo{TableName: "film_category", TableAlias: alias}
	tbl.FILM_ID = NewNumberField("film_id", tbl.TableInfo)
	tbl.CATEGORY_ID = NewNumberField("category_id", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type xSTAFF struct {
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

func xNEW_STAFF(alias string) xSTAFF {
	var tbl xSTAFF
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

type xSTORE struct {
	TableInfo
	STORE_ID         NumberField
	MANAGER_STAFF_ID NumberField
	ADDRESS_ID       NumberField
	LAST_UPDATE      TimeField
}

func xNEW_STORE(alias string) xSTORE {
	var tbl xSTORE
	tbl.TableInfo = TableInfo{TableName: "store", TableAlias: alias}
	tbl.STORE_ID = NewNumberField("store_id", tbl.TableInfo)
	tbl.MANAGER_STAFF_ID = NewNumberField("manager_staff_id", tbl.TableInfo)
	tbl.ADDRESS_ID = NewNumberField("address_id", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type xCUSTOMER struct {
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

func xNEW_CUSTOMER(alias string) xCUSTOMER {
	var tbl xCUSTOMER
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

type xINVENTORY struct {
	TableInfo
	INVENTORY_ID NumberField
	FILM_ID      NumberField
	STORE_ID     NumberField
	LAST_UPDATE  TimeField
}

func xNEW_INVENTORY(alias string) xINVENTORY {
	var tbl xINVENTORY
	tbl.TableInfo = TableInfo{TableName: "inventory", TableAlias: alias}
	tbl.INVENTORY_ID = NewNumberField("inventory_id", tbl.TableInfo)
	tbl.FILM_ID = NewNumberField("film_id", tbl.TableInfo)
	tbl.STORE_ID = NewNumberField("store_id", tbl.TableInfo)
	tbl.LAST_UPDATE = NewTimeField("last_update", tbl.TableInfo)
	return tbl
}

type xRENTAL struct {
	TableInfo
	RENTAL_ID    NumberField
	RENTAL_DATE  TimeField
	INVENTORY_ID NumberField
	CUSTOMER_ID  NumberField
	RETURN_DATE  TimeField
	STAFF_ID     NumberField
	LAST_UPDATE  TimeField
}

func xNEW_RENTAL(alias string) xRENTAL {
	var tbl xRENTAL
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

type xPAYMENT struct {
	TableInfo
	PAYMENT_ID   NumberField
	CUSTOMER_ID  NumberField
	STAFF_ID     NumberField
	RENTAL_ID    NumberField
	AMOUNT       NumberField
	PAYMENT_DATE TimeField
}

func xNEW_PAYMENT(alias string) xPAYMENT {
	var tbl xPAYMENT
	tbl.TableInfo = TableInfo{TableName: "payment", TableAlias: alias}
	tbl.PAYMENT_ID = NewNumberField("payment_id", tbl.TableInfo)
	tbl.CUSTOMER_ID = NewNumberField("customer_id", tbl.TableInfo)
	tbl.STAFF_ID = NewNumberField("staff_id", tbl.TableInfo)
	tbl.RENTAL_ID = NewNumberField("rental_id", tbl.TableInfo)
	tbl.AMOUNT = NewNumberField("amount", tbl.TableInfo)
	tbl.PAYMENT_DATE = NewTimeField("payment_date", tbl.TableInfo)
	return tbl
}

// TODO: I *need* semantically named types, not just generic RecordN types.
// This would make it more palatable in the documentation when I inevitably
// show test snippets in it.

type Record4 struct {
	ActorID   int
	FirstName string
	LastName  string
}

type Record6 struct {
	CountryName string
	CityName    string
}

type Record7 struct {
	FilmTitle  string
	Length     int
	Rating     string
	Audience   string
	LengthType string
}

type Record8 struct {
	FilmTitle  string
	ActorCount int
}

type Record9 struct {
	CategoryName string
	Revenue      int
	Rank         int
	Quartile     int
}

type Record10 struct {
	Month       string
	HorrorCount int
	ActionCount int
	ComedyCount int
	ScifiCount  int
}

type SQLTest struct {
	Dialect string

	// Q1) Find all distinct actor last names ordered by last name. Show only
	// the top 5 results.
	Query1  Query
	Answer1 []string

	// Q2) Find if there is any actor with first name 'SCARLETT' or last name
	// 'JOHANSSON'.
	Query2  Query
	Answer2 bool

	// Q3) Find the number of distinct actor last names.
	Query3  Query
	Answer3 int

	// Q4) Find all actors whose last name contain the letters 'GEN', ordered
	// by actor_id. Return the actor_id, first_name and last_name.
	Query4  Query
	Answer4 []Record4

	// Q5) Find actor last names that only once in the database, ordered by
	// last_name.  Show only the top 5 results.
	Query5  Query
	Answer5 []string

	// Q6) Find all the cities of the countries Egypt, Greece and Puerto Rico,
	// ordered by country_name and city_name. Return the country_name and
	// city_name.
	Query6  Query
	Answer6 []Record6

	// Q7) List films together with their audience and length type, ordered by title.
	// The audiences are:
	// - rating = 'G' then 'family'
	// - rating = 'PG' or rating = 'PG-13' then 'teens'
	// - rating = 'R' or rating = 'NC-17' then 'adults'
	// The length types are:
	// - length <= 60 then 'short'
	// - length > 60 and length <= 120 then 'medium'
	// - length > 120 then 'long'
	// Return the film_title, rating, length, audience and length_type. Show only the
	// top 10 results.
	Query7  Query
	Answer7 []Record7

	// Q8) Find the films whose total number of actors is above the average,
	// ordered by title. Return the film title and the actor count. Show only
	// the top 10 results.
	Query8  Query
	Answer8 []Record8

	// Q9) List the film categories and their total revenue (rounded to nearest
	// integer), ordered by descending revenue. Return the category name,
	// revenue, the rank of that category and the quartile it belongs to
	// (relative to the other categories).
	Query9  Query
	Answer9 []Record9

	// Q10) Find the total number of 'Horror' films, 'Action' films, 'Comedy'
	// films and 'Sci-Fi' films rented out every month between the dates
	// '2005-03-01' and '2006-02-01', ordered by month. Months with 0 rentals
	// should also be included. Return the month, horror_count, action_count,
	// comedy_count and scifi_count.
	Query10  Query
	Answer10 []Record10

	// TODO: test application-side aggregation queries

	// TODO: test database-side JSON aggregation queries

	// TODO: test full text search queries

	// TODO: move the queries out of the test suite. All it should contain are
	// the answers.
}

func NewSQLTest() SQLTest {
	return SQLTest{
		Answer1: []string{"AKROYD", "ALLEN", "ASTAIRE", "BACALL", "BAILEY"},
		Answer2: true,
		Answer3: 121,
		Answer4: []Record4{
			{ActorID: 14, FirstName: "VIVIEN", LastName: "BERGEN"},
			{ActorID: 41, FirstName: "JODIE", LastName: "DEGENERES"},
			{ActorID: 107, FirstName: "GINA", LastName: "DEGENERES"},
			{ActorID: 166, FirstName: "NICK", LastName: "DEGENERES"},
		},
		Answer5: []string{"ASTAIRE", "BACALL", "BALE", "BALL", "BARRYMORE"},
		Answer6: []Record6{
			{CountryName: "Egypt", CityName: "Bilbays"},
			{CountryName: "Egypt", CityName: "Idfu"},
			{CountryName: "Egypt", CityName: "Mit Ghamr"},
			{CountryName: "Egypt", CityName: "Qalyub"},
			{CountryName: "Egypt", CityName: "Sawhaj"},
			{CountryName: "Egypt", CityName: "Shubra al-Khayma"},
			{CountryName: "Greece", CityName: "Athenai"},
			{CountryName: "Greece", CityName: "Patras"},
			{CountryName: "Puerto Rico", CityName: "Arecibo"},
			{CountryName: "Puerto Rico", CityName: "Ponce"},
		},
		Answer7: []Record7{
			{FilmTitle: "ACADEMY DINOSAUR", Rating: "PG", Length: 86, Audience: "teens", LengthType: "medium"},
			{FilmTitle: "ACE GOLDFINGER", Rating: "G", Length: 48, Audience: "family", LengthType: "short"},
			{FilmTitle: "ADAPTATION HOLES", Rating: "NC-17", Length: 50, Audience: "adults", LengthType: "short"},
			{FilmTitle: "AFFAIR PREJUDICE", Rating: "G", Length: 117, Audience: "family", LengthType: "medium"},
			{FilmTitle: "AFRICAN EGG", Rating: "G", Length: 130, Audience: "family", LengthType: "long"},
			{FilmTitle: "AGENT TRUMAN", Rating: "PG", Length: 169, Audience: "teens", LengthType: "long"},
			{FilmTitle: "AIRPLANE SIERRA", Rating: "PG-13", Length: 62, Audience: "teens", LengthType: "medium"},
			{FilmTitle: "AIRPORT POLLOCK", Rating: "R", Length: 54, Audience: "adults", LengthType: "short"},
			{FilmTitle: "ALABAMA DEVIL", Rating: "PG-13", Length: 114, Audience: "teens", LengthType: "medium"},
			{FilmTitle: "ALADDIN CALENDAR", Rating: "NC-17", Length: 63, Audience: "adults", LengthType: "medium"},
		},
		Answer8: []Record8{
			{FilmTitle: "ALABAMA DEVIL", ActorCount: 24},
			{FilmTitle: "ACADEMY DINOSAUR", ActorCount: 23},
			{FilmTitle: "ALICE FANTASIA", ActorCount: 21},
			{FilmTitle: "AGENT TRUMAN", ActorCount: 20},
			{FilmTitle: "AIRPORT POLLOCK", ActorCount: 20},
			{FilmTitle: "ALASKA PHANTOM", ActorCount: 19},
			{FilmTitle: "ALI FOREVER", ActorCount: 19},
			{FilmTitle: "AIRPLANE SIERRA", ActorCount: 17},
			{FilmTitle: "ALAMO VIDEOTAPE", ActorCount: 17},
			{FilmTitle: "ADAPTATION HOLES", ActorCount: 16},
		},
		Answer9: []Record9{
			{CategoryName: "Sports", Revenue: 5314, Rank: 1, Quartile: 4},
			{CategoryName: "Sci-Fi", Revenue: 4757, Rank: 2, Quartile: 4},
			{CategoryName: "Animation", Revenue: 4656, Rank: 3, Quartile: 4},
			{CategoryName: "Drama", Revenue: 4587, Rank: 4, Quartile: 4},
			{CategoryName: "Comedy", Revenue: 4384, Rank: 5, Quartile: 3},
			{CategoryName: "Action", Revenue: 4376, Rank: 6, Quartile: 3},
			{CategoryName: "New", Revenue: 4353, Rank: 7, Quartile: 3},
			{CategoryName: "Games", Revenue: 4281, Rank: 8, Quartile: 3},
			{CategoryName: "Foreign", Revenue: 4271, Rank: 9, Quartile: 2},
			{CategoryName: "Family", Revenue: 4235, Rank: 10, Quartile: 2},
			{CategoryName: "Documentary", Revenue: 4218, Rank: 11, Quartile: 2},
			{CategoryName: "Horror", Revenue: 3723, Rank: 12, Quartile: 2},
			{CategoryName: "Children", Revenue: 3656, Rank: 13, Quartile: 1},
			{CategoryName: "Classics", Revenue: 3640, Rank: 14, Quartile: 1},
			{CategoryName: "Travel", Revenue: 3550, Rank: 15, Quartile: 1},
			{CategoryName: "Music", Revenue: 3418, Rank: 16, Quartile: 1},
		},
		Answer10: []Record10{
			{Month: "2005 March"},
			{Month: "2005 April"},
			{Month: "2005 May", HorrorCount: 14, ActionCount: 15, ComedyCount: 9, ScifiCount: 12},
			{Month: "2005 June", HorrorCount: 27, ActionCount: 32, ComedyCount: 27, ScifiCount: 26},
			{Month: "2005 July", HorrorCount: 79, ActionCount: 91, ComedyCount: 82, ScifiCount: 83},
			{Month: "2005 August", HorrorCount: 71, ActionCount: 88, ComedyCount: 68, ScifiCount: 76},
			{Month: "2005 September"},
			{Month: "2005 October"},
			{Month: "2005 November"},
			{Month: "2005 December"},
			{Month: "2006 January"},
			{Month: "2006 February", HorrorCount: 3, ActionCount: 2, ComedyCount: 6, ScifiCount: 4},
		},
	}
}
