package sq

import "time"

type Actor struct {
	ActorID    int       `json:"actor_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	LastUpdate time.Time `json:"last_update"`
}

type Film struct {
	FilmID          int       `json:"film_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	ReleaseYear     string    `json:"release_year"`
	RentalDuration  int       `json:"rental_duration"`
	RentalRate      int       `json:"rental_rate"`
	Length          int       `json:"length"`
	ReplacementCost int       `json:"replacement_cost"`
	Rating          string    `json:"rating"`
	SpecialFeatures []string  `json:"special_features"`
	Actors          []Actor   `json:"actors"`
	Audience        string    `json:"audience"`
	LengthType      string    `json:"length_type"`
	LastUpdate      time.Time `json:"last_update"`
}

type Rental struct {
	RentalID   int       `json:"rental_id"`
	RentalDate time.Time `json:"rental_date"`
	Inventory  Inventory `json:"inventory"`
	Customer   Customer  `json:"customer"`
}

type Staff struct {
	StaffID    int       `json:"staff_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Picture    []byte    `json:"picture"`
	Email      string    `json:"email"`
	Active     bool      `json:"active"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	LastUpdate time.Time `json:"last_update"`
}

type Inventory struct {
	InventoryID int       `json:"inventory_id"`
	Film        Film      `json:"film"`
	LastUpdate  time.Time `json:"last_update"`
}

type Store struct {
	StoreID     int         `json:"store_id"`
	Inventories []Inventory `json:"inventories"`
	LastUpdate  time.Time   `json:"last_update"`
}

type City struct {
	Country    Country   `json:"country"`
	CityID     int       `json:"city_id"`
	CityName   string    `json:"city"`
	LastUpdate time.Time `json:"last_update"`
}

type Country struct {
	CountryID   int       `json:"country_id"`
	CountryName string    `json:"country"`
	LastUpdate  time.Time `json:"last_update"`
}

type MonthlyRentalStats struct {
	Month       time.Time `json:"month"`
	HorrorCount int64     `json:"horror_count"`
	ActionCount int64     `json:"action_count"`
	ComedyCount int64     `json:"comedy_count"`
	ScifiCount  int64     `json:"scifi_count"`
}

type Customer struct {
	CustomerID int       `json:"customer_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"name"`
	Active     bool      `json:"active"`
	CreateDate time.Time `json:"create_date"`
	LastUpdate time.Time `json:"last_update"`
}

type CustomerRentalStats struct {
	Customer    Customer `json:"customer"`
	RentalCount int64    `json:"rental_count"`
}

type FilmActorStats struct {
	Film       Film  `json:"film"`
	ActorCount int64 `json:"actor_count"`
}

type Category struct {
	CategoryID   int       `json:"category_id"`
	CategoryName string    `json:"category_name"`
	LastUpdate   time.Time `json:"last_update"`
}

type CategoryRevenueStats struct {
	Category Category `json:"category"`
	Revenue  int64    `json:"revenue"`
	Rank     int      `json:"rank"`
	Quartile int      `json:"quartile"`
}

type TestSuiteAnswers struct {
	Answer01 []string
	Answer02 bool
	Answer03 int
	Answer04 []Actor
	Answer05 []string
	Answer06 []City
	Answer07 []Film
	Answer08 []FilmActorStats
	Answer09 []CategoryRevenueStats
	Answer10 []MonthlyRentalStats
	Answer11 []Store
	Answer12 []Film
}

func NewTestSuiteAnswers() TestSuiteAnswers {
	datetime := func(year, month, day, hour, min, sec int) time.Time {
		return time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
	}
	return TestSuiteAnswers{
		Answer01: []string{"AKROYD", "ALLEN", "ASTAIRE", "BACALL", "BAILEY"},
		Answer02: true,
		Answer03: 121,
		Answer04: []Actor{
			{ActorID: 14, FirstName: "VIVIEN", LastName: "BERGEN", LastUpdate: datetime(2006, 2, 15, 4, 34, 33)},
			{ActorID: 41, FirstName: "JODIE", LastName: "DEGENERES", LastUpdate: datetime(2006, 2, 15, 4, 34, 33)},
			{ActorID: 107, FirstName: "GINA", LastName: "DEGENERES", LastUpdate: datetime(2006, 2, 15, 4, 34, 33)},
			{ActorID: 166, FirstName: "NICK", LastName: "DEGENERES", LastUpdate: datetime(2006, 2, 15, 4, 34, 33)},
		},
		Answer05: []string{"ASTAIRE", "BACALL", "BALE", "BALL", "BARRYMORE"},
		Answer06: []City{
			{
				Country: Country{CountryID: 29, CountryName: "Egypt", LastUpdate: datetime(2006, 2, 15, 4, 44, 0)},
				CityID:  79, CityName: "Bilbays", LastUpdate: datetime(2006, 02, 15, 4, 45, 25),
			},
			{
				Country: Country{CountryID: 29, CountryName: "Egypt", LastUpdate: datetime(2006, 2, 15, 4, 44, 0)},
				CityID:  216, CityName: "Idfu", LastUpdate: datetime(2006, 02, 15, 4, 45, 25),
			},
			{
				Country: Country{CountryID: 29, CountryName: "Egypt", LastUpdate: datetime(2006, 2, 15, 4, 44, 0)},
				CityID:  337, CityName: "Mit Ghamr", LastUpdate: datetime(2006, 02, 15, 4, 45, 25),
			},
			{
				Country: Country{CountryID: 29, CountryName: "Egypt", LastUpdate: datetime(2006, 2, 15, 4, 44, 0)},
				CityID:  421, CityName: "Qalyub", LastUpdate: datetime(2006, 02, 15, 4, 45, 25),
			},
			{
				Country: Country{CountryID: 29, CountryName: "Egypt", LastUpdate: datetime(2006, 2, 15, 4, 44, 0)},
				CityID:  465, CityName: "Sawhaj", LastUpdate: datetime(2006, 02, 15, 4, 45, 25),
			},
			{
				Country: Country{CountryID: 29, CountryName: "Egypt", LastUpdate: datetime(2006, 2, 15, 4, 44, 0)},
				CityID:  476, CityName: "Shubra al-Khayma", LastUpdate: datetime(2006, 02, 15, 4, 45, 25),
			},
			{
				Country: Country{CountryID: 39, CountryName: "Greece", LastUpdate: datetime(2006, 2, 15, 4, 44, 0)},
				CityID:  38, CityName: "Athenai", LastUpdate: datetime(2006, 02, 15, 4, 45, 25),
			},
			{
				Country: Country{CountryID: 39, CountryName: "Greece", LastUpdate: datetime(2006, 2, 15, 4, 44, 0)},
				CityID:  401, CityName: "Patras", LastUpdate: datetime(2006, 02, 15, 4, 45, 25),
			},
			{
				Country: Country{CountryID: 77, CountryName: "Puerto Rico", LastUpdate: datetime(2006, 2, 15, 4, 44, 0)},
				CityID:  32, CityName: "Arecibo", LastUpdate: datetime(2006, 02, 15, 4, 45, 25),
			},
			{
				Country: Country{CountryID: 77, CountryName: "Puerto Rico", LastUpdate: datetime(2006, 2, 15, 4, 44, 0)},
				CityID:  411, CityName: "Ponce", LastUpdate: datetime(2006, 02, 15, 4, 45, 25),
			},
		},
		Answer07: []Film{},
		Answer08: []FilmActorStats{},
		Answer09: []CategoryRevenueStats{},
		Answer10: []MonthlyRentalStats{},
		Answer11: []Store{},
		Answer12: []Film{},
	}
}
