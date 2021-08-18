package sq

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
)

func Test_SQLiteInsertQuery(t *testing.T) {
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

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = SQLite.
			InsertInto(ACTOR).
			InsertInto(ACTOR).
			Columns(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
			Values("bob", "the builder").
			Values("alice", "in wonderland").
			With(NewCTE("cte", []string{"n"}, Queryf("SELECT 1")))
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" INSERT INTO actor AS a (first_name, last_name)" +
			" VALUES ($1, $2), ($3, $4)"
		tt.wantArgs = []interface{}{"bob", "the builder", "alice", "in wonderland"}
		assert(t, tt)
	})

	t.Run("INSERT with RETURNING", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = SQLite.
			InsertWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			InsertInto(ACTOR).
			Valuesx(func(c *Column) error {
				// bob
				c.SetString(ACTOR.FIRST_NAME, "bob")
				c.SetString(ACTOR.LAST_NAME, "the builder")
				// alice
				c.SetString(ACTOR.FIRST_NAME, "alice")
				c.SetString(ACTOR.LAST_NAME, "in wonderland")
				return nil
			}).
			Returning(ACTOR.ACTOR_ID)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" INSERT INTO actor AS a (first_name, last_name)" +
			" VALUES ($1, $2), ($3, $4)" +
			" RETURNING a.actor_id"
		tt.wantArgs = []interface{}{"bob", "the builder", "alice", "in wonderland"}
		assert(t, tt)
	})

	t.Run("INSERT ignore duplicates", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = SQLite.
			InsertWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			InsertInto(ACTOR).
			Valuesx(func(c *Column) error {
				// bob
				c.SetInt64(ACTOR.ACTOR_ID, 1)
				c.SetString(ACTOR.FIRST_NAME, "bob")
				c.SetString(ACTOR.LAST_NAME, "the builder")
				// alice
				c.SetInt64(ACTOR.ACTOR_ID, 2)
				c.SetString(ACTOR.FIRST_NAME, "alice")
				c.SetString(ACTOR.LAST_NAME, "in wonderland")
				return nil
			}).
			OnConflict(ACTOR.ACTOR_ID).
			Where(And(ACTOR.ACTOR_ID.IsNotNull(), ACTOR.FIRST_NAME.NeString(""))).
			DoNothing()
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" INSERT INTO actor AS a (actor_id, first_name, last_name)" +
			" VALUES ($1, $2, $3), ($4, $5, $6)" +
			" ON CONFLICT (actor_id)" +
			" WHERE actor_id IS NOT NULL AND first_name <> $7" +
			" DO NOTHING"
		tt.wantArgs = []interface{}{int64(1), "bob", "the builder", int64(2), "alice", "in wonderland", ""}
		assert(t, tt)
	})

	t.Run("upsert", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("a")
		tt.item = SQLite.
			InsertWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			InsertInto(ACTOR).
			Valuesx(func(c *Column) error {
				// bob
				c.SetInt64(ACTOR.ACTOR_ID, 1)
				c.SetString(ACTOR.FIRST_NAME, "bob")
				c.SetString(ACTOR.LAST_NAME, "the builder")
				// alice
				c.SetInt64(ACTOR.ACTOR_ID, 2)
				c.SetString(ACTOR.FIRST_NAME, "alice")
				c.SetString(ACTOR.LAST_NAME, "in wonderland")
				return nil
			}).
			OnConflict(ACTOR.ACTOR_ID).
			Where(And(ACTOR.ACTOR_ID.IsNotNull(), ACTOR.FIRST_NAME.NeString(""))).
			DoUpdateSet(AssignExcluded(ACTOR.FIRST_NAME), AssignExcluded(ACTOR.LAST_NAME)).
			Where(ACTOR.LAST_UPDATE.IsNotNull(), ACTOR.LAST_NAME.NeString(""))
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" INSERT INTO actor AS a (actor_id, first_name, last_name)" +
			" VALUES ($1, $2, $3), ($4, $5, $6)" +
			" ON CONFLICT (actor_id)" +
			" WHERE actor_id IS NOT NULL AND first_name <> $7" +
			" DO UPDATE SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name" +
			" WHERE a.last_update IS NOT NULL AND a.last_name <> $8"
		tt.wantArgs = []interface{}{int64(1), "bob", "the builder", int64(2), "alice", "in wonderland", "", ""}
		assert(t, tt)
	})

	t.Run("INSERT from SELECT", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR1, ACTOR2 := xNEW_ACTOR("a1"), xNEW_ACTOR("a2")
		tt.item = SQLite.
			InsertWith(NewCTE("cte", []string{"n"}, Queryf("SELECT 1"))).
			InsertInto(ACTOR1).
			Columns(ACTOR1.FIRST_NAME, ACTOR1.LAST_NAME).
			Select(SQLite.
				Select(ACTOR2.FIRST_NAME, ACTOR2.LAST_NAME).
				From(ACTOR2).
				Where(ACTOR2.ACTOR_ID.In([]int64{1, 2})),
			)
		tt.wantQuery = "WITH cte (n) AS (SELECT 1)" +
			" INSERT INTO actor AS a1 (first_name, last_name)" +
			" SELECT a2.first_name, a2.last_name" +
			" FROM actor AS a2" +
			" WHERE a2.actor_id IN ($1, $2)"
		tt.wantArgs = []interface{}{int64(1), int64(2)}
		assert(t, tt)
	})
}

func TestSQLiteSakilaInsert(t *testing.T) {
	if testing.Short() {
		return
	}
	tx, err := sqliteDB.Begin()
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	defer tx.Rollback()
	CUSTOMER := xNEW_CUSTOMER("")
	regina := Customer{StoreID: 1, AddressID: 1, FirstName: "REGINA", LastName: "TATE", Email: "regina_tate@email.com"}
	customers := []Customer{
		{StoreID: 1, AddressID: 1, FirstName: "JULIA", LastName: "HAYWARD", Email: "julia_hayward@email.com"},
		{StoreID: 1, AddressID: 1, FirstName: "DUNCAN", LastName: "PEARSON", Email: "duncan_pearson@email.com"},
		{StoreID: 1, AddressID: 1, FirstName: "IDA", LastName: "WATKINS", Email: "ida_watkins@email.com"},
		{StoreID: 1, AddressID: 1, FirstName: "THOMAS", LastName: "BINDER", Email: "thomas_binder@email.com"},
	}
	// {StoreID: 1, AddressID: 1, FirstName: "ASTRID", LastName: "SILVA", Email: "astrid_silva@email.com"},
	// {StoreID: 1, AddressID: 1, FirstName: "HARPER", LastName: "CRAIG", Email: "harper_craig@email.com"},
	// {StoreID: 1, AddressID: 1, FirstName: "SAMANTHA", LastName: "STEVENSON", Email: "samantha_stevenson@email.com"},
	// {StoreID: 1, AddressID: 1, FirstName: "PHILIP", LastName: "REID", Email: "philip_reid@email.com"},

	// add regina
	rowsAffected, lastInsertID, err := Exec(Log(tx), SQLite.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			col.SetInt(CUSTOMER.STORE_ID, regina.StoreID)
			col.SetString(CUSTOMER.FIRST_NAME, regina.FirstName)
			col.SetString(CUSTOMER.LAST_NAME, regina.LastName)
			col.SetString(CUSTOMER.EMAIL, regina.Email)
			col.SetInt(CUSTOMER.ADDRESS_ID, regina.AddressID)
			return nil
		}),
		ErowsAffected|ElastInsertID,
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 1 {
		t.Fatal(testutil.Callers(), "expected 1 row to be affected but got %d", rowsAffected)
	}
	regina.CustomerID = int(lastInsertID)

	// ensure regina exists
	exists, err := FetchExists(Log(tx), SQLite.From(CUSTOMER).Where(
		CUSTOMER.CUSTOMER_ID.EqInt(regina.CustomerID),
		CUSTOMER.STORE_ID.EqInt(regina.StoreID),
		CUSTOMER.FIRST_NAME.EqString(regina.FirstName),
		CUSTOMER.LAST_NAME.EqString(regina.LastName),
		CUSTOMER.EMAIL.EqString(regina.Email),
		CUSTOMER.ADDRESS_ID.EqInt(regina.AddressID),
	))
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if !exists {
		t.Fatal(testutil.Callers(), "expected inserted customer %+v to exist", regina)
	}

	// add regina again and check that ON CONFLICT DO NOTHING kicks in
	rowsAffected, lastInsertID, err = Exec(Log(tx), SQLite.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			col.SetInt(CUSTOMER.CUSTOMER_ID, regina.CustomerID)
			col.SetInt(CUSTOMER.STORE_ID, regina.StoreID)
			col.SetString(CUSTOMER.FIRST_NAME, regina.FirstName)
			col.SetString(CUSTOMER.LAST_NAME, regina.LastName)
			col.SetString(CUSTOMER.EMAIL, regina.Email)
			col.SetInt(CUSTOMER.ADDRESS_ID, regina.AddressID)
			return nil
		}).
		OnConflict().DoNothing(),
		ErowsAffected|ElastInsertID,
	)
	if rowsAffected != 0 {
		t.Fatal(testutil.Callers(), "expected an second identical insert to affect 0 rows, got %d instead", rowsAffected)
	}

	// modify and upsert regina
	regina.FirstName = regina.FirstName[:1] + strings.ToLower(regina.FirstName[1:])
	regina.LastName = regina.LastName[:1] + strings.ToLower(regina.LastName[1:])
	rowsAffected, lastInsertID, err = Exec(Log(tx), SQLite.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			col.SetInt(CUSTOMER.CUSTOMER_ID, regina.CustomerID)
			col.SetInt(CUSTOMER.STORE_ID, regina.StoreID)
			col.SetString(CUSTOMER.FIRST_NAME, regina.FirstName)
			col.SetString(CUSTOMER.LAST_NAME, regina.LastName)
			col.SetString(CUSTOMER.EMAIL, regina.Email)
			col.SetInt(CUSTOMER.ADDRESS_ID, regina.AddressID)
			return nil
		}).
		OnConflict(CUSTOMER.CUSTOMER_ID).
		DoUpdateSet(
			AssignExcluded(CUSTOMER.STORE_ID),
			AssignExcluded(CUSTOMER.FIRST_NAME),
			AssignExcluded(CUSTOMER.LAST_NAME),
			AssignExcluded(CUSTOMER.EMAIL),
			AssignExcluded(CUSTOMER.ADDRESS_ID),
		),
		ErowsAffected|ElastInsertID,
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 1 {
		t.Fatal(testutil.Callers(), "expected 1 row to be upserted but got %d", rowsAffected)
	}

	// ensure the modified regina exists
	exists, err = FetchExists(Log(tx), SQLite.From(CUSTOMER).Where(
		CUSTOMER.CUSTOMER_ID.EqInt(regina.CustomerID),
		CUSTOMER.STORE_ID.EqInt(regina.StoreID),
		CUSTOMER.FIRST_NAME.EqString(regina.FirstName),
		CUSTOMER.LAST_NAME.EqString(regina.LastName),
		CUSTOMER.EMAIL.EqString(regina.Email),
		CUSTOMER.ADDRESS_ID.EqInt(regina.AddressID),
	))
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if !exists {
		t.Fatal(testutil.Callers(), "expected inserted customer %+v to exist", regina)
	}

	// add the first 2 customers
	var customerIDs []int
	rowCount, err := Fetch(Log(tx), SQLite.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			for _, customer := range customers[:2] {
				col.SetInt(CUSTOMER.STORE_ID, customer.StoreID)
				col.SetString(CUSTOMER.FIRST_NAME, customer.FirstName)
				col.SetString(CUSTOMER.LAST_NAME, customer.LastName)
				col.SetString(CUSTOMER.EMAIL, customer.Email)
				col.SetInt(CUSTOMER.ADDRESS_ID, customer.AddressID)
			}
			return nil
		}),
		func(row *Row) {
			customerID := row.Int(CUSTOMER.CUSTOMER_ID)
			row.Process(func() { customerIDs = append(customerIDs, customerID) })
		},
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowCount != 2 {
		t.Fatal(testutil.Callers(), "expected 2 rows inserted but got %d", rowCount)
	}
	for i := 0; i < 2; i++ {
		customers[i].CustomerID = customerIDs[i]
	}

	// ensure the first 2 customers exist
	predicate := And()
	for _, customer := range customers[:2] {
		predicate = predicate.Append(Exists(SQLite.
			SelectOne().
			From(CUSTOMER).
			Where(
				CUSTOMER.CUSTOMER_ID.EqInt(customer.CustomerID),
				CUSTOMER.STORE_ID.EqInt(customer.StoreID),
				CUSTOMER.FIRST_NAME.EqString(customer.FirstName),
				CUSTOMER.LAST_NAME.EqString(customer.LastName),
				CUSTOMER.EMAIL.EqString(customer.Email),
				CUSTOMER.ADDRESS_ID.EqInt(customer.AddressID),
			),
		))
	}
	_, err = Fetch(Log(tx), SQLite.Select(), func(row *Row) { exists = row.Bool(predicate) })
	if !exists {
		t.Fatal(testutil.Callers(), "expected inserted customers %+v to exist", customers[:2])
	}

	// add the first 2 customers again and ensure ON CONFLICT DO NOTHING kicks in
	rowCount, err = Fetch(Log(tx), SQLite.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			for _, customer := range customers[:2] {
				col.SetInt(CUSTOMER.CUSTOMER_ID, customer.CustomerID)
				col.SetInt(CUSTOMER.STORE_ID, customer.StoreID)
				col.SetString(CUSTOMER.FIRST_NAME, customer.FirstName)
				col.SetString(CUSTOMER.LAST_NAME, customer.LastName)
				col.SetString(CUSTOMER.EMAIL, customer.Email)
				col.SetInt(CUSTOMER.ADDRESS_ID, customer.AddressID)
			}
			return nil
		}).
		OnConflict().DoNothing(),
		func(row *Row) {
			_ = row.Int(CUSTOMER.CUSTOMER_ID)
		},
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowCount != 0 {
		t.Fatal(testutil.Callers(), "expected an second identical insert to affect 0 rows, got %d instead", rowCount)
	}

	// add all 4 customers and check that only the last 2 customers got added
	rowCount, err = Fetch(Log(tx), SQLite.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			for _, customer := range customers {
				col.SetInt(CUSTOMER.STORE_ID, customer.StoreID)
				col.SetString(CUSTOMER.FIRST_NAME, customer.FirstName)
				col.SetString(CUSTOMER.LAST_NAME, customer.LastName)
				col.SetString(CUSTOMER.EMAIL, customer.Email)
				col.SetInt(CUSTOMER.ADDRESS_ID, customer.AddressID)
			}
			return nil
		}).
		OnConflict().DoNothing(),
		func(row *Row) {
			customerID := row.Int(CUSTOMER.CUSTOMER_ID)
			row.Process(func() { customerIDs = append(customerIDs, customerID) })
		},
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if int(rowCount) != 2 {
		t.Fatal(testutil.Callers(), "expected 2 rows inserted but got %d", rowCount)
	}
	for i := 2; i < 4; i++ {
		customers[i].CustomerID = customerIDs[i]
	}

	// check that all 4 customers exist
	predicate = And()
	for _, customer := range customers {
		predicate = predicate.Append(Exists(SQLite.
			SelectOne().
			From(CUSTOMER).
			Where(
				CUSTOMER.CUSTOMER_ID.EqInt(customer.CustomerID),
				CUSTOMER.STORE_ID.EqInt(customer.StoreID),
				CUSTOMER.FIRST_NAME.EqString(customer.FirstName),
				CUSTOMER.LAST_NAME.EqString(customer.LastName),
				CUSTOMER.EMAIL.EqString(customer.Email),
				CUSTOMER.ADDRESS_ID.EqInt(customer.AddressID),
			),
		))
	}
	_, err = Fetch(Log(tx), SQLite.Select(), func(row *Row) { exists = row.Bool(predicate) })
	if !exists {
		t.Fatalf(testutil.Callers()+" expected inserted customers %+v to exist", customers[:4])
	}

	// modify and upsert the first 2 customers
	for i, customer := range customers[:2] {
		customers[i].FirstName = customer.FirstName[:1] + strings.ToLower(customer.FirstName[1:])
		customers[i].LastName = customer.LastName[:1] + strings.ToLower(customer.LastName[1:])
	}
	rowsAffected, _, err = Exec(Log(tx), SQLite.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			for _, customer := range customers[:2] {
				col.SetInt(CUSTOMER.CUSTOMER_ID, customer.CustomerID)
				col.SetInt(CUSTOMER.STORE_ID, customer.StoreID)
				col.SetString(CUSTOMER.FIRST_NAME, customer.FirstName)
				col.SetString(CUSTOMER.LAST_NAME, customer.LastName)
				col.SetString(CUSTOMER.EMAIL, customer.Email)
				col.SetInt(CUSTOMER.ADDRESS_ID, customer.AddressID)
			}
			return nil
		}).
		OnConflict(CUSTOMER.CUSTOMER_ID).
		DoUpdateSet(
			AssignExcluded(CUSTOMER.STORE_ID),
			AssignExcluded(CUSTOMER.FIRST_NAME),
			AssignExcluded(CUSTOMER.LAST_NAME),
			AssignExcluded(CUSTOMER.EMAIL),
			AssignExcluded(CUSTOMER.ADDRESS_ID),
		),
		ErowsAffected,
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 2 {
		t.Fatal(testutil.Callers(), fmt.Sprintf("expected 2 rows to be upserted but got %d", rowsAffected))
	}

	// check that all 4 customers (including the modified 2) exist
	predicate = And()
	for _, customer := range customers {
		predicate = predicate.Append(Exists(SQLite.
			SelectOne().
			From(CUSTOMER).
			Where(
				CUSTOMER.CUSTOMER_ID.EqInt(customer.CustomerID),
				CUSTOMER.STORE_ID.EqInt(customer.StoreID),
				CUSTOMER.FIRST_NAME.EqString(customer.FirstName),
				CUSTOMER.LAST_NAME.EqString(customer.LastName),
				CUSTOMER.EMAIL.EqString(customer.Email),
				CUSTOMER.ADDRESS_ID.EqInt(customer.AddressID),
			),
		))
	}
	_, err = Fetch(Log(tx), SQLite.Select(), func(row *Row) { exists = row.Bool(predicate) })
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if !exists {
		t.Fatal(testutil.Callers(), "expected inserted customers %+v to exist", customers[:4])
	}

	// Customer 'MARY SMITH' rents the film 'ACADEMY DINOSAUR' from staff 'Mike
	// Hillyer' at Store 1 on 9th of August 2021 4pm. Insert a rental record
	// representing that transaction (INSERT with SELECT).
	STAFF := xNEW_STAFF("")
	FILM := xNEW_FILM("")
	INVENTORY := xNEW_INVENTORY("")
	STORE := xNEW_STORE("")
	RENTAL := xNEW_RENTAL("")
	customer_id := NewSubquery("customer_id", SQLite.
		Select(CUSTOMER.CUSTOMER_ID).
		From(CUSTOMER).
		Where(RowValue{CUSTOMER.FIRST_NAME, CUSTOMER.LAST_NAME}.Eq(RowValue{"MARY", "SMITH"})).
		Limit(1),
	)
	staff_id := NewSubquery("staff_id", SQLite.
		Select(STAFF.STAFF_ID).
		From(STAFF).
		Where(
			STAFF.STORE_ID.EqInt(1),
			RowValue{STAFF.FIRST_NAME, STAFF.LAST_NAME}.Eq(RowValue{"Mike", "Hillyer"}),
		).
		Limit(1),
	)
	var rentalID int
	_, err = Fetch(Log(tx), SQLite.
		InsertInto(RENTAL).
		Columns(RENTAL.INVENTORY_ID, RENTAL.CUSTOMER_ID, RENTAL.STAFF_ID, RENTAL.RENTAL_DATE).
		Select(SQLite.
			Select(
				INVENTORY.INVENTORY_ID,
				Value(customer_id),
				Value(staff_id),
				Value(datetime(2021, 8, 9, 16, 0, 0)).As("rental_date"),
			).
			From(FILM).
			Join(INVENTORY, INVENTORY.FILM_ID.Eq(FILM.FILM_ID)).
			Join(STORE, STORE.STORE_ID.Eq(INVENTORY.STORE_ID)).
			Where(
				FILM.TITLE.EqString("ACADEMY DINOSAUR"),
				STORE.STORE_ID.EqInt(1),
				Not(Exists(SQLite.
					SelectOne().
					From(RENTAL).
					Where(RENTAL.INVENTORY_ID.Eq(INVENTORY.INVENTORY_ID), RENTAL.RENTAL_DATE.IsNull()),
				)),
			).
			OrderBy(INVENTORY.INVENTORY_ID).
			Limit(1),
		),
		func(row *Row) {
			rentalID = row.Int(RENTAL.RENTAL_ID)
		},
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}

	// check that the rentalID returned has the attributes we want
	exists, err = FetchExists(Log(tx), SQLite.
		From(RENTAL).
		Join(CUSTOMER, CUSTOMER.CUSTOMER_ID.Eq(RENTAL.CUSTOMER_ID)).
		Join(STAFF, STAFF.STAFF_ID.Eq(RENTAL.STAFF_ID)).
		Join(INVENTORY, INVENTORY.INVENTORY_ID.Eq(RENTAL.INVENTORY_ID)).
		Join(FILM, FILM.FILM_ID.Eq(INVENTORY.FILM_ID)).
		Where(
			RENTAL.RENTAL_ID.EqInt(rentalID),
			RowValue{CUSTOMER.FIRST_NAME, CUSTOMER.LAST_NAME}.Eq(RowValue{"MARY", "SMITH"}),
			STAFF.STORE_ID.EqInt(1),
			RowValue{STAFF.FIRST_NAME, STAFF.LAST_NAME}.Eq(RowValue{"Mike", "Hillyer"}),
			FILM.TITLE.EqString("ACADEMY DINOSAUR"),
			RENTAL.RENTAL_DATE.EqTime(datetime(2021, 8, 9, 16, 0, 0)),
		),
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if !exists {
		t.Fatalf(testutil.Callers()+"record record with rental_id %d does not have the attributes we want", rentalID)
	}
}
