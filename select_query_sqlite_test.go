package sq

import (
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
)

func Test_SQLiteSelectQuery(t *testing.T) {
	type TT struct {
		dialect   string
		item      Query
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		gotQuery, gotArgs, _, err := ToSQL(tt.dialect, tt.item)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotQuery, tt.wantQuery); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
		if diff := testutil.Diff(gotArgs, tt.wantArgs); diff != "" {
			t.Error(testutil.Callers(), diff)
		}
	}

	t.Run("joins", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = SQLite.
			From(ACTOR).
			From(ACTOR).
			SelectOne().
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			Join(ACTOR, Eq(1, 1)).
			LeftJoin(ACTOR, Eq(1, 1)).
			CrossJoin(ACTOR).
			CustomJoin(",", ACTOR)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" SELECT 1" +
			" FROM actor AS a" +
			" JOIN actor AS a ON $1 = $2" +
			" LEFT JOIN actor AS a ON $3 = $4" +
			" CROSS JOIN actor AS a" +
			" , actor AS a"
		tt.wantArgs = []interface{}{1, 1, 1, 1}
		assert(t, tt)
	})

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = SQLite.
			SelectDistinct(ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
			SelectDistinct(ACTOR.ACTOR_ID, ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
			From(ACTOR).
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			GroupBy(ACTOR.FIRST_NAME).
			Having(ACTOR.FIRST_NAME.IsNotNull()).
			OrderBy(ACTOR.LAST_NAME).
			Limit(10).
			Offset(20)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" SELECT DISTINCT a.actor_id, a.first_name, a.last_name" +
			" FROM actor AS a" +
			" GROUP BY a.first_name" +
			" HAVING a.first_name IS NOT NULL" +
			" ORDER BY a.last_name" +
			" LIMIT $1" +
			" OFFSET $2"
		tt.wantArgs = []interface{}{int64(10), int64(20)}
		assert(t, tt)
	})
}

func TestSQLiteSakilaSelect(t *testing.T) {
	t.Run("Q1", func(t *testing.T) {
		t.Parallel()
		var gotAnswer []string
		wantAnswer := sakilaAnswer1()
		ACTOR := xNEW_ACTOR("")
		_, err := Fetch(Log(sqliteDB), SQLite.
			SelectDistinct().
			From(ACTOR).
			OrderBy(ACTOR.LAST_NAME).
			Limit(5),
			func(row *Row) {
				lastName := row.String(ACTOR.LAST_NAME)
				row.Process(func() { gotAnswer = append(gotAnswer, lastName) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotAnswer, wantAnswer); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q2", func(t *testing.T) {
		t.Parallel()
		wantAnswer := sakilaAnswer2()
		ACTOR := xNEW_ACTOR("")
		gotAnswer, err := FetchExists(Log(sqliteDB), SQLite.
			From(ACTOR).
			Where(Or(
				ACTOR.FIRST_NAME.EqString("SCARLETT"),
				ACTOR.LAST_NAME.EqString("JOHANSSON"),
			)),
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotAnswer, wantAnswer); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q3", func(t *testing.T) {
		t.Parallel()
		var gotAnswer int
		wantAnswer := sakilaAnswer3()
		ACTOR := xNEW_ACTOR("")
		_, err := Fetch(Log(sqliteDB), SQLite.From(ACTOR), func(row *Row) {
			gotAnswer = row.Int(NumberFieldf("COUNT(DISTINCT {})", ACTOR.LAST_NAME))
			row.Close()
		})
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotAnswer, wantAnswer); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q4", func(t *testing.T) {
		t.Parallel()
		var gotAnswer []Actor
		wantAnswer := sakilaAnswer4()
		ACTOR := xNEW_ACTOR("")
		_, err := Fetch(Log(sqliteDB), SQLite.
			From(ACTOR).
			Where(ACTOR.LAST_NAME.LikeString("%GEN%")).
			OrderBy(ACTOR.ACTOR_ID),
			func(row *Row) {
				actor := Actor{
					ActorID:    row.Int(ACTOR.ACTOR_ID),
					FirstName:  row.String(ACTOR.FIRST_NAME),
					LastName:   row.String(ACTOR.LAST_NAME),
					LastUpdate: row.Time(ACTOR.LAST_UPDATE),
				}
				row.Process(func() { gotAnswer = append(gotAnswer, actor) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotAnswer, wantAnswer); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q5", func(t *testing.T) {
		t.Parallel()
		var gotAnswer []string
		wantAnswer := sakilaAnswer5()
		ACTOR := xNEW_ACTOR("")
		_, err := Fetch(Log(sqliteDB), SQLite.
			From(ACTOR).
			GroupBy(ACTOR.LAST_NAME).
			Having(Fieldf("COUNT(*)").Eq(1)).
			OrderBy(ACTOR.LAST_NAME).
			Limit(5),
			func(row *Row) {
				lastName := row.String(ACTOR.LAST_NAME)
				row.Process(func() { gotAnswer = append(gotAnswer, lastName) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotAnswer, wantAnswer); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q6", func(t *testing.T) {
		t.Parallel()
		var gotAnswer []City
		wantAnswer := sakilaAnswer6()
		CITY, COUNTRY := xNEW_CITY(""), xNEW_COUNTRY("")
		_, err := Fetch(Log(sqliteDB), SQLite.
			From(CITY).
			Join(COUNTRY, COUNTRY.COUNTRY_ID.Eq(CITY.COUNTRY_ID)).
			Where(COUNTRY.COUNTRY.In([]string{"Egypt", "Greece", "Puerto Rico"})).
			OrderBy(COUNTRY.COUNTRY, CITY.CITY),
			func(row *Row) {
				city := City{
					Country: Country{
						CountryID:   row.Int(COUNTRY.COUNTRY_ID),
						CountryName: row.String(COUNTRY.COUNTRY),
						LastUpdate:  row.Time(COUNTRY.LAST_UPDATE),
					},
					CityID:     row.Int(CITY.CITY_ID),
					CityName:   row.String(CITY.CITY),
					LastUpdate: row.Time(CITY.LAST_UPDATE),
				}
				row.Process(func() { gotAnswer = append(gotAnswer, city) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotAnswer, wantAnswer); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q7", func(t *testing.T) {
		t.Parallel()
		var gotAnswer []Film
		wantAnswer := sakilaAnswer7()
		FILM := xNEW_FILM("")
		_, err := Fetch(Log(sqliteDB), SQLite.
			From(FILM).
			OrderBy(FILM.TITLE).
			Limit(10),
			func(row *Row) {
				film := Film{
					FilmID:          row.Int(FILM.FILM_ID),
					Title:           row.String(FILM.TITLE),
					Description:     row.String(FILM.DESCRIPTION),
					ReleaseYear:     row.Int(FILM.RELEASE_YEAR),
					RentalDuration:  row.Int(FILM.RENTAL_DURATION),
					RentalRate:      row.Float64(FILM.RENTAL_RATE),
					Length:          row.Int(FILM.LENGTH),
					ReplacementCost: row.Float64(FILM.REPLACEMENT_COST),
					Rating:          row.String(FILM.RATING),
					LastUpdate:      row.Time(FILM.LAST_UPDATE),
				}
				row.ScanJSON(&film.SpecialFeatures, FILM.SPECIAL_FEATURES)
				row.ScanInto(&film.Audience, Case(FILM.RATING).
					When("G", "family").
					When("PG", "teens").
					When("PG-13", "teens").
					When("R", "adults").
					When("NC-17", "adults"),
				)
				row.ScanInto(&film.LengthType, CaseWhen(FILM.LENGTH.LeInt(60), "short").
					When(And(FILM.LENGTH.GtInt(60), FILM.LENGTH.LeInt(120)), "medium").
					Else("long"),
				)
				row.Process(func() { gotAnswer = append(gotAnswer, film) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotAnswer, wantAnswer); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q8", func(t *testing.T) {
		t.Parallel()
		var gotAnswer []FilmActorStats
		wantAnswer := sakilaAnswer8()
		FILM, FILM_ACTOR := xNEW_FILM(""), xNEW_FILM_ACTOR("")
		film_stats := NewCTE("film_stats", nil, SQLite.
			Select(FILM_ACTOR.FILM_ID, Fieldf("COUNT(*)").As("actor_count")).
			From(FILM_ACTOR).
			GroupBy(FILM_ACTOR.FILM_ID),
		)
		_, err := Fetch(Log(sqliteDB), SQLite.
			SelectWith(film_stats).
			From(film_stats).
			Join(FILM, film_stats.Field("film_id").Eq(FILM.FILM_ID)).
			Where(film_stats.Field("actor_count").Gt(SQLite.Select(Fieldf("AVG(actor_count)")).From(film_stats))).
			OrderBy(
				film_stats.Field("actor_count").Desc(),
				FILM.TITLE.Asc(),
			).
			Limit(10),
			func(row *Row) {
				film := Film{
					FilmID:          row.Int(FILM.FILM_ID),
					Title:           row.String(FILM.TITLE),
					Description:     row.String(FILM.DESCRIPTION),
					ReleaseYear:     row.Int(FILM.RELEASE_YEAR),
					RentalDuration:  row.Int(FILM.RENTAL_DURATION),
					RentalRate:      row.Float64(FILM.RENTAL_RATE),
					Length:          row.Int(FILM.LENGTH),
					ReplacementCost: row.Float64(FILM.REPLACEMENT_COST),
					Rating:          row.String(FILM.RATING),
					LastUpdate:      row.Time(FILM.LAST_UPDATE),
				}
				row.ScanJSON(&film.SpecialFeatures, FILM.SPECIAL_FEATURES)
				filmActorStats := FilmActorStats{Film: film}
				row.ScanInto(&filmActorStats.ActorCount, film_stats.Field("actor_count"))
				row.Process(func() { gotAnswer = append(gotAnswer, filmActorStats) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotAnswer, wantAnswer); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q9", func(t *testing.T) {
		t.Parallel()
		var gotAnswer []CategoryRevenueStats
		wantAnswer := sakilaAnswer9()
		CATEGORY := xNEW_CATEGORY("")
		FILM_CATEGORY := xNEW_FILM_CATEGORY("")
		INVENTORY := xNEW_INVENTORY("")
		RENTAL := xNEW_RENTAL("")
		PAYMENT := xNEW_PAYMENT("")
		_, err := Fetch(Log(sqliteDB), SQLite.
			From(CATEGORY).
			Join(FILM_CATEGORY, FILM_CATEGORY.CATEGORY_ID.Eq(CATEGORY.CATEGORY_ID)).
			Join(INVENTORY, INVENTORY.FILM_ID.Eq(FILM_CATEGORY.FILM_ID)).
			Join(RENTAL, RENTAL.INVENTORY_ID.Eq(INVENTORY.INVENTORY_ID)).
			Join(PAYMENT, PAYMENT.RENTAL_ID.Eq(RENTAL.RENTAL_ID)).
			GroupBy(CATEGORY.CATEGORY_ID, CATEGORY.NAME, CATEGORY.LAST_UPDATE).
			OrderBy(Fieldf("revenue").Desc()),
			func(row *Row) {
				stats := CategoryRevenueStats{
					Category: Category{
						CategoryID:   row.Int(CATEGORY.CATEGORY_ID),
						CategoryName: row.String(CATEGORY.NAME),
						LastUpdate:   row.Time(CATEGORY.LAST_UPDATE),
					},
					Revenue:  row.Float64(NumberFieldf("ROUND(SUM({}), 2)", PAYMENT.AMOUNT).As("revenue")),
					Rank:     row.Int(RankOver(OrderBy(Sum(PAYMENT.AMOUNT).Desc()))),
					Quartile: row.Int(NtileOver(4, OrderBy(Sum(PAYMENT.AMOUNT).Asc()))),
				}
				row.Process(func() { gotAnswer = append(gotAnswer, stats) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotAnswer, wantAnswer); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("Q10", func(t *testing.T) {
		t.Parallel()
		var gotAnswer []MonthlyRentalStats
		wantAnswer := sakilaAnswer10()
		RENTAL := xNEW_RENTAL("")
		FILM_CATEGORY := xNEW_FILM_CATEGORY("")
		CATEGORY := xNEW_CATEGORY("")
		dates := NewRecursiveCTE("dates", []string{"date_value"}, UnionAll(
			SQLite.Select(Fieldf("DATE({})", "2005-03-01")),
			SQLite.Select(Fieldf("DATE(date_value, '+1 month')")).From(Tablef("dates")).Where(Predicatef("date_value < {}", "2006-02-01")),
		))
		months := NewCTE("months", []string{"num", "name"}, Queryf(
			"VALUES {}", RowValues{
				{"01", "January"}, {"02", "February"}, {"03", "March"},
				{"04", "April"}, {"05", "May"}, {"06", "June"},
				{"07", "July"}, {"08", "August"}, {"09", "September"},
				{"10", "October"}, {"11", "November"}, {"12", "December"},
			},
		))
		_, err := Fetch(VerboseLog(sqliteDB), SQLite.
			SelectWith(dates, months).
			From(dates).
			Join(months, months.Field("num").Eq(Fieldf(`strftime('%m', {})`, dates.Field("date_value")))).
			LeftJoin(RENTAL, Predicatef(`strftime('%Y %m', {}) = strftime('%Y %m', {})`, RENTAL.RENTAL_DATE, dates.Field("date_value"))).
			LeftJoin(FILM_CATEGORY, FILM_CATEGORY.FILM_ID.Eq(RENTAL.INVENTORY_ID)).
			LeftJoin(CATEGORY, CATEGORY.CATEGORY_ID.Eq(FILM_CATEGORY.CATEGORY_ID)).
			GroupBy(dates.Field("date_value")).
			OrderBy(dates.Field("date_value")),
			func(row *Row) {
				stats := MonthlyRentalStats{
					Month:       row.String(StringFieldf("strftime('%Y', {}) || ' ' || {}", dates.Field("date_value"), months.Field("name")).As("month")),
					HorrorCount: row.Int64(Count(Case(CATEGORY.NAME).When("Horror", 1)).As("horror_count")),
					ActionCount: row.Int64(Count(Case(CATEGORY.NAME).When("Action", 1)).As("action_count")),
					ComedyCount: row.Int64(Count(Case(CATEGORY.NAME).When("Comedy", 1)).As("comedy_count")),
					ScifiCount:  row.Int64(Count(Case(CATEGORY.NAME).When("Sci-Fi", 1)).As("scifi_count")),
				}
				row.Process(func() { gotAnswer = append(gotAnswer, stats) })
			},
		)
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if diff := testutil.Diff(gotAnswer, wantAnswer); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})
}
