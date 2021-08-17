package sq

import "time"

const (
	Trailers        = "Trailers"
	Commentaries    = "Commentaries"
	DeletedScenes   = "Deleted Scenes"
	BehindTheScenes = "Behind the Scenes"
)

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
	ReleaseYear     int       `json:"release_year"`
	RentalDuration  int       `json:"rental_duration"`  // The length of the rental period, in days.
	RentalRate      float64   `json:"rental_rate"`      // The cost to rent the film for the period specified in the rental_duration column.
	Length          int       `json:"length"`           // The duration of the film, in minutes.
	ReplacementCost float64   `json:"replacement_cost"` // The amount charged to the customer if the film is not returned or is returned in a damaged state.
	Rating          string    `json:"rating"`
	SpecialFeatures []string  `json:"special_features"` //  Lists which common special features are included on the DVD.
	Actors          []Actor   `json:"actors"`
	LastUpdate      time.Time `json:"last_update"`
	Audience        string    `json:"audience"`
	LengthType      string    `json:"length_type"`
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
	Month       string `json:"month"`
	HorrorCount int64  `json:"horror_count"`
	ActionCount int64  `json:"action_count"`
	ComedyCount int64  `json:"comedy_count"`
	ScifiCount  int64  `json:"scifi_count"`
}

type Customer struct {
	CustomerID int       `json:"customer_id"`
	StoreID    int       `json:"store_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"name"`
	AddressID  int       `json:"address_id"`
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
	Revenue  float64  `json:"revenue"`
	Rank     int      `json:"rank"`
	Quartile int      `json:"quartile"`
}

func datetime(year, month, day, hour, min, sec int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)
}

// Q1) Find all distinct actor last names ordered by last name. Show only the
// top 5 results.
func sakilaAnswer1() []string {
	return []string{"AKROYD", "ALLEN", "ASTAIRE", "BACALL", "BAILEY"}
}

// Q2) Find if there is any actor with first name 'SCARLETT' or last name
// 'JOHANSSON'.
func sakilaAnswer2() bool { return true }

// Q3) Find the number of distinct actor last names.
func sakilaAnswer3() int { return 121 }

// Q4) Find all actors whose last name contain the letters 'GEN', ordered by
// actor_id.
func sakilaAnswer4() []Actor {
	return []Actor{
		{ActorID: 14, FirstName: "VIVIEN", LastName: "BERGEN", LastUpdate: datetime(2006, 2, 15, 4, 34, 33)},
		{ActorID: 41, FirstName: "JODIE", LastName: "DEGENERES", LastUpdate: datetime(2006, 2, 15, 4, 34, 33)},
		{ActorID: 107, FirstName: "GINA", LastName: "DEGENERES", LastUpdate: datetime(2006, 2, 15, 4, 34, 33)},
		{ActorID: 166, FirstName: "NICK", LastName: "DEGENERES", LastUpdate: datetime(2006, 2, 15, 4, 34, 33)},
	}
}

// Q5) Find actor last names that occur only once in the database, ordered by
// last_name. Show only the top 5 results.
func sakilaAnswer5() []string {
	return []string{"ASTAIRE", "BACALL", "BALE", "BALL", "BARRYMORE"}
}

// Q6) Find all the cities (and their country) of the where country is either
// 'Egypt', 'Greece' or 'Puerto Rico', ordered by country_name and city_name.
func sakilaAnswer6() []City {
	return []City{
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
	}
}

// Q7) List films together with their audience and length type, ordered by
// title.
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
func sakilaAnswer7() []Film {
	return []Film{
		{
			FilmID: 1, Title: "ACADEMY DINOSAUR",
			Description: "A Epic Drama of a Feminist And a Mad Scientist who must Battle a Teacher in The Canadian Rockies",
			ReleaseYear: 2006, RentalDuration: 6, RentalRate: 0.99, Length: 86, ReplacementCost: 20.99, Rating: "PG",
			SpecialFeatures: []string{DeletedScenes, BehindTheScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			Audience: "teens", LengthType: "medium",
		},
		{
			FilmID: 2, Title: "ACE GOLDFINGER",
			Description: "A Astounding Epistle of a Database Administrator And a Explorer who must Find a Car in Ancient China",
			ReleaseYear: 2006, RentalDuration: 3, RentalRate: 4.99, Length: 48, ReplacementCost: 12.99, Rating: "G",
			SpecialFeatures: []string{Trailers, DeletedScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			Audience: "family", LengthType: "short",
		},
		{
			FilmID: 3, Title: "ADAPTATION HOLES",
			Description: "A Astounding Reflection of a Lumberjack And a Car who must Sink a Lumberjack in A Baloon Factory",
			ReleaseYear: 2006, RentalDuration: 7, RentalRate: 2.99, Length: 50, ReplacementCost: 18.99, Rating: "NC-17",
			SpecialFeatures: []string{Trailers, DeletedScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			Audience: "adults", LengthType: "short",
		},
		{
			FilmID: 4, Title: "AFFAIR PREJUDICE",
			Description: "A Fanciful Documentary of a Frisbee And a Lumberjack who must Chase a Monkey in A Shark Tank",
			ReleaseYear: 2006, RentalDuration: 5, RentalRate: 2.99, Length: 117, ReplacementCost: 26.99, Rating: "G",
			SpecialFeatures: []string{Commentaries, BehindTheScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			Audience: "family", LengthType: "medium",
		},
		{
			FilmID: 5, Title: "AFRICAN EGG",
			Description: "A Fast-Paced Documentary of a Pastry Chef And a Dentist who must Pursue a Forensic Psychologist in The Gulf of Mexico",
			ReleaseYear: 2006, RentalDuration: 6, RentalRate: 2.99, Length: 130, ReplacementCost: 22.99, Rating: "G",
			SpecialFeatures: []string{DeletedScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			Audience: "family", LengthType: "long",
		},
		{
			FilmID: 6, Title: "AGENT TRUMAN",
			Description: "A Intrepid Panorama of a Robot And a Boy who must Escape a Sumo Wrestler in Ancient China",
			ReleaseYear: 2006, RentalDuration: 3, RentalRate: 2.99, Length: 169, ReplacementCost: 17.99, Rating: "PG",
			SpecialFeatures: []string{DeletedScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			Audience: "teens", LengthType: "long",
		},
		{
			FilmID: 7, Title: "AIRPLANE SIERRA",
			Description: "A Touching Saga of a Hunter And a Butler who must Discover a Butler in A Jet Boat",
			ReleaseYear: 2006, RentalDuration: 6, RentalRate: 4.99, Length: 62, ReplacementCost: 28.99, Rating: "PG-13",
			SpecialFeatures: []string{Trailers, DeletedScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			Audience: "teens", LengthType: "medium",
		},
		{
			FilmID: 8, Title: "AIRPORT POLLOCK",
			Description: "A Epic Tale of a Moose And a Girl who must Confront a Monkey in Ancient India",
			ReleaseYear: 2006, RentalDuration: 6, RentalRate: 4.99, Length: 54, ReplacementCost: 15.99, Rating: "R",
			SpecialFeatures: []string{Trailers}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			Audience: "adults", LengthType: "short",
		},
		{
			FilmID: 9, Title: "ALABAMA DEVIL",
			Description: "A Thoughtful Panorama of a Database Administrator And a Mad Scientist who must Outgun a Mad Scientist in A Jet Boat",
			ReleaseYear: 2006, RentalDuration: 3, RentalRate: 2.99, Length: 114, ReplacementCost: 21.99, Rating: "PG-13",
			SpecialFeatures: []string{Trailers, DeletedScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			Audience: "teens", LengthType: "medium",
		},
		{
			FilmID: 10, Title: "ALADDIN CALENDAR",
			Description: "A Action-Packed Tale of a Man And a Lumberjack who must Reach a Feminist in Ancient China",
			ReleaseYear: 2006, RentalDuration: 6, RentalRate: 4.99, Length: 63, ReplacementCost: 24.99, Rating: "NC-17",
			SpecialFeatures: []string{Trailers, DeletedScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			Audience: "adults", LengthType: "medium",
		},
	}
}

// Q8) Find the films whose total number of actors is above the average,
// ordered by descending actor count. Show only the top 10 results.
func sakilaAnswer8() []FilmActorStats {
	return []FilmActorStats{
		{
			Film: Film{
				FilmID: 9, Title: "ALABAMA DEVIL",
				Description: "A Thoughtful Panorama of a Database Administrator And a Mad Scientist who must Outgun a Mad Scientist in A Jet Boat",
				ReleaseYear: 2006, RentalDuration: 3, RentalRate: 2.99, Length: 114, ReplacementCost: 21.99, Rating: "PG-13",
				SpecialFeatures: []string{Trailers, DeletedScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			},
			ActorCount: 24,
		},
		{
			Film: Film{
				FilmID: 1, Title: "ACADEMY DINOSAUR",
				Description: "A Epic Drama of a Feminist And a Mad Scientist who must Battle a Teacher in The Canadian Rockies",
				ReleaseYear: 2006, RentalDuration: 6, RentalRate: 0.99, Length: 86, ReplacementCost: 20.99, Rating: "PG",
				SpecialFeatures: []string{DeletedScenes, BehindTheScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			},
			ActorCount: 23,
		},
		{
			Film: Film{
				FilmID: 14, Title: "ALICE FANTASIA",
				Description: "A Emotional Drama of a A Shark And a Database Administrator who must Vanquish a Pioneer in Soviet Georgia",
				ReleaseYear: 2006, RentalDuration: 6, RentalRate: 0.99, Length: 94, ReplacementCost: 23.99, Rating: "NC-17",
				SpecialFeatures: []string{Trailers, DeletedScenes, BehindTheScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			},
			ActorCount: 21,
		},
		{
			Film: Film{
				FilmID: 6, Title: "AGENT TRUMAN",
				Description: "A Intrepid Panorama of a Robot And a Boy who must Escape a Sumo Wrestler in Ancient China",
				ReleaseYear: 2006, RentalDuration: 3, RentalRate: 2.99, Length: 169, ReplacementCost: 17.99, Rating: "PG",
				SpecialFeatures: []string{DeletedScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			},
			ActorCount: 20,
		},
		{
			Film: Film{
				FilmID: 8, Title: "AIRPORT POLLOCK",
				Description: "A Epic Tale of a Moose And a Girl who must Confront a Monkey in Ancient India",
				ReleaseYear: 2006, RentalDuration: 6, RentalRate: 4.99, Length: 54, ReplacementCost: 15.99, Rating: "R",
				SpecialFeatures: []string{Trailers}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			},
			ActorCount: 20,
		},
		{
			Film: Film{
				FilmID: 12, Title: "ALASKA PHANTOM",
				Description: "A Fanciful Saga of a Hunter And a Pastry Chef who must Vanquish a Boy in Australia",
				ReleaseYear: 2006, RentalDuration: 6, RentalRate: 0.99, Length: 136, ReplacementCost: 22.99, Rating: "PG",
				SpecialFeatures: []string{Commentaries, DeletedScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			},
			ActorCount: 19,
		},
		{
			Film: Film{
				FilmID: 13, Title: "ALI FOREVER",
				Description: "A Action-Packed Drama of a Dentist And a Crocodile who must Battle a Feminist in The Canadian Rockies",
				ReleaseYear: 2006, RentalDuration: 4, RentalRate: 4.99, Length: 150, ReplacementCost: 21.99, Rating: "PG",
				SpecialFeatures: []string{DeletedScenes, BehindTheScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			},
			ActorCount: 19,
		},
		{
			Film: Film{
				FilmID: 7, Title: "AIRPLANE SIERRA",
				Description: "A Touching Saga of a Hunter And a Butler who must Discover a Butler in A Jet Boat",
				ReleaseYear: 2006, RentalDuration: 6, RentalRate: 4.99, Length: 62, ReplacementCost: 28.99, Rating: "PG-13",
				SpecialFeatures: []string{Trailers, DeletedScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			},
			ActorCount: 17,
		},
		{
			Film: Film{
				FilmID: 11, Title: "ALAMO VIDEOTAPE",
				Description: "A Boring Epistle of a Butler And a Cat who must Fight a Pastry Chef in A MySQL Convention",
				ReleaseYear: 2006, RentalDuration: 6, RentalRate: 0.99, Length: 126, ReplacementCost: 16.99, Rating: "G",
				SpecialFeatures: []string{Commentaries, BehindTheScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			},
			ActorCount: 17,
		},
		{
			Film: Film{
				FilmID: 3, Title: "ADAPTATION HOLES",
				Description: "A Astounding Reflection of a Lumberjack And a Car who must Sink a Lumberjack in A Baloon Factory",
				ReleaseYear: 2006, RentalDuration: 7, RentalRate: 2.99, Length: 50, ReplacementCost: 18.99, Rating: "NC-17",
				SpecialFeatures: []string{Trailers, DeletedScenes}, LastUpdate: datetime(2006, 2, 15, 5, 03, 42),
			},
			ActorCount: 16,
		},
	}
}

// Q9) List the film categories and their total revenue (rounded to 2 decimal
// places), ordered by descending revenue. Include the rank of that category
// and the quartile it belongs to, relative to the other categories.
func sakilaAnswer9() []CategoryRevenueStats {
	return []CategoryRevenueStats{
		{
			Category: Category{CategoryID: 15, CategoryName: "Sports", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  5314.21, Rank: 1, Quartile: 4,
		},
		{
			Category: Category{CategoryID: 14, CategoryName: "Sci-Fi", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  4756.98, Rank: 2, Quartile: 4,
		},
		{
			Category: Category{CategoryID: 2, CategoryName: "Animation", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  4656.30, Rank: 3, Quartile: 4,
		},
		{
			Category: Category{CategoryID: 7, CategoryName: "Drama", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  4587.39, Rank: 4, Quartile: 4,
		},
		{
			Category: Category{CategoryID: 5, CategoryName: "Comedy", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  4383.58, Rank: 5, Quartile: 3,
		},
		{
			Category: Category{CategoryID: 1, CategoryName: "Action", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  4375.85, Rank: 6, Quartile: 3,
		},
		{
			Category: Category{CategoryID: 13, CategoryName: "New", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  4352.61, Rank: 7, Quartile: 3,
		},
		{
			Category: Category{CategoryID: 10, CategoryName: "Games", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  4281.33, Rank: 8, Quartile: 3,
		},
		{
			Category: Category{CategoryID: 9, CategoryName: "Foreign", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  4270.67, Rank: 9, Quartile: 2,
		},
		{
			Category: Category{CategoryID: 8, CategoryName: "Family", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  4235.03, Rank: 10, Quartile: 2,
		},
		{
			Category: Category{CategoryID: 6, CategoryName: "Documentary", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  4217.52, Rank: 11, Quartile: 2,
		},
		{
			Category: Category{CategoryID: 11, CategoryName: "Horror", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  3722.54, Rank: 12, Quartile: 2,
		},
		{
			Category: Category{CategoryID: 3, CategoryName: "Children", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  3655.55, Rank: 13, Quartile: 1,
		},
		{
			Category: Category{CategoryID: 4, CategoryName: "Classics", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  3639.59, Rank: 14, Quartile: 1,
		},
		{
			Category: Category{CategoryID: 16, CategoryName: "Travel", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  3549.64, Rank: 15, Quartile: 1,
		},
		{
			Category: Category{CategoryID: 12, CategoryName: "Music", LastUpdate: datetime(2006, 02, 15, 4, 46, 27)},
			Revenue:  3417.72, Rank: 16, Quartile: 1,
		},
	}
}

// Q10) Find the total number of 'Horror' films, 'Action' films, 'Comedy' films
// and 'Sci-Fi' films rented out every month between '2005-03-01' and
// '2006-02-01', ordered by month. Months with 0 rentals should also be
// included.
func sakilaAnswer10() []MonthlyRentalStats {
	return []MonthlyRentalStats{
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
	}
}

func thoraTempleFilmIDs() []int {
	return []int{
		5, 10, 49, 80, 116, 121,
		149, 346, 419, 462, 465,
		474, 537, 538, 544, 714,
		879, 912, 945, 958, 993,
	}
}
