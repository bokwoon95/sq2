package sq

import (
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
			Where(ACTOR.ACTOR_ID.IsNotNull(), ACTOR.FIRST_NAME.NeString("")).
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
			Where(ACTOR.ACTOR_ID.IsNotNull(), ACTOR.FIRST_NAME.NeString("")).
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
	tx, err := sqliteDB.Begin()
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	defer tx.Rollback()
	CUSTOMER := xNEW_CUSTOMER("")

	customer1 := Customer{
		StoreID:   1,
		FirstName: "REGINA", LastName: "TATE",
		Email:     "regina_tate@email.com",
		AddressID: 1,
	}
	rowsAffected, lastInsertID, err := Exec(Log(tx), SQLite.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			col.SetInt(CUSTOMER.STORE_ID, customer1.StoreID)
			col.SetString(CUSTOMER.FIRST_NAME, customer1.FirstName)
			col.SetString(CUSTOMER.LAST_NAME, customer1.LastName)
			col.SetString(CUSTOMER.EMAIL, customer1.Email)
			col.SetInt(CUSTOMER.ADDRESS_ID, customer1.AddressID)
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
	customer1.CustomerID = int(lastInsertID)

	exists, err := FetchExists(Log(tx), SQLite.From(CUSTOMER).Where(
		CUSTOMER.CUSTOMER_ID.EqInt(customer1.CustomerID),
		CUSTOMER.STORE_ID.EqInt(customer1.StoreID),
		CUSTOMER.FIRST_NAME.EqString(customer1.FirstName),
		CUSTOMER.LAST_NAME.EqString(customer1.LastName),
		CUSTOMER.EMAIL.EqString(customer1.Email),
		CUSTOMER.ADDRESS_ID.EqInt(customer1.AddressID),
	))
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if !exists {
		t.Fatal(testutil.Callers(), "expected inserted customer %+v to exist", customer1)
	}

	rowsAffected, lastInsertID, err = Exec(Log(tx), SQLite.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			col.SetInt(CUSTOMER.CUSTOMER_ID, customer1.CustomerID)
			col.SetInt(CUSTOMER.STORE_ID, customer1.StoreID)
			col.SetString(CUSTOMER.FIRST_NAME, customer1.FirstName)
			col.SetString(CUSTOMER.LAST_NAME, customer1.LastName)
			col.SetString(CUSTOMER.EMAIL, customer1.Email)
			col.SetInt(CUSTOMER.ADDRESS_ID, customer1.AddressID)
			return nil
		}).
		OnConflict(CUSTOMER.CUSTOMER_ID).DoNothing(),
		ErowsAffected|ElastInsertID,
	)
	if rowsAffected != 0 {
		t.Fatal(testutil.Callers(), "expected an second identical insert to affect 0 rows, got %d instead", rowsAffected)
	}

	customer2 := Customer{
		StoreID: 1, AddressID: 1,
		FirstName: "ANTHONY", LastName: "CURTIS",
		Email: "anthony_curtis@email.com",
	}
	rowsAffected, lastInsertID, err = Exec(Log(tx), SQLite.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			col.SetInt(CUSTOMER.STORE_ID, customer2.StoreID)
			col.SetString(CUSTOMER.FIRST_NAME, customer2.FirstName)
			col.SetString(CUSTOMER.LAST_NAME, customer2.LastName)
			col.SetString(CUSTOMER.EMAIL, customer2.Email)
			col.SetInt(CUSTOMER.ADDRESS_ID, customer2.AddressID)
			return nil
		}).
		OnConflict(CUSTOMER.EMAIL).
		DoUpdateSet(
			AssignExcluded(CUSTOMER.STORE_ID),
			AssignExcluded(CUSTOMER.FIRST_NAME),
			AssignExcluded(CUSTOMER.LAST_NAME),
			AssignExcluded(CUSTOMER.ADDRESS_ID),
		),
		ErowsAffected|ElastInsertID,
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 1 {
		t.Fatal(testutil.Callers(), "expected 1 row to be affected but got %d", rowsAffected)
	}
	customer2.CustomerID = int(lastInsertID)

	exists, err = FetchExists(Log(tx), SQLite.From(CUSTOMER).Where(
		CUSTOMER.CUSTOMER_ID.EqInt(customer2.CustomerID),
		CUSTOMER.STORE_ID.EqInt(customer2.StoreID),
		CUSTOMER.FIRST_NAME.EqString(customer2.FirstName),
		CUSTOMER.LAST_NAME.EqString(customer2.LastName),
		CUSTOMER.EMAIL.EqString(customer2.Email),
		CUSTOMER.ADDRESS_ID.EqInt(customer2.AddressID),
	))
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if !exists {
		t.Fatal(testutil.Callers(), "expected inserted customer %+v to exist", customer1)
	}

	customers := []Customer{
		{
			StoreID: 1, AddressID: 1,
			FirstName: "JULIA", LastName: "HAYWARD",
			Email: "julia_hayward@email.com",
		},
		{
			StoreID: 1, AddressID: 1,
			FirstName: "DUNCAN", LastName: "PEARSON",
			Email: "duncan_pearson@email.com",
		},
		{
			StoreID: 1, AddressID: 1,
			FirstName: "IDA", LastName: "WATKINS",
			Email: "ida_watkins@email.com",
		},
		{
			StoreID: 1, AddressID: 1,
			FirstName: "THOMAS", LastName: "BINDER",
			Email: "thomas_binder@email.com",
		},
		{
			StoreID: 1, AddressID: 1,
			FirstName: "ASTRID", LastName: "SILVA",
			Email: "astrid_silva@email.com",
		},
		{
			StoreID: 1, AddressID: 1,
			FirstName: "HARPER", LastName: "CRAIG",
			Email: "harper_craig@email.com",
		},
		{
			StoreID: 1, AddressID: 1,
			FirstName: "SAMANTHA", LastName: "STEVENSON",
			Email: "samantha_stevenson@email.com",
		},
		{
			StoreID: 1, AddressID: 1,
			FirstName: "PHILIP", LastName: "REID",
			Email: "philip_reid@email.com",
		},
	}
	var customerIDs []int
	rowCount, err := Fetch(Log(tx), SQLite.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			for _, customer := range customers[:4] {
				col.SetInt(CUSTOMER.STORE_ID, customer.StoreID)
				col.SetString(CUSTOMER.FIRST_NAME, customer.FirstName)
				col.SetString(CUSTOMER.LAST_NAME, customer.LastName)
				col.SetString(CUSTOMER.EMAIL, customer.Email)
				col.SetInt(CUSTOMER.ADDRESS_ID, customer.AddressID)
			}
			return nil
		}).
		Returning(CUSTOMER.CUSTOMER_ID),
		func(row *Row) {
			customerID := row.Int(CUSTOMER.CUSTOMER_ID)
			row.Process(func() { customerIDs = append(customerIDs, customerID) })
		},
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowCount != 4 {
		t.Fatal(testutil.Callers(), "expected 4 rows inserted but got %d", rowsAffected)
	}
	for i := 0; i < 4; i++ {
		customers[i].CustomerID = customerIDs[i]
	}

	predicate := And()
	for _, customer := range customers[:4] {
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
		t.Fatal(testutil.Callers(), "expected inserted customers %+v to exist", customers[:4])
	}

	rowCount, err = Fetch(Log(tx), SQLite.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			for _, customer := range customers[4:] {
				col.SetInt(CUSTOMER.STORE_ID, customer.StoreID)
				col.SetString(CUSTOMER.FIRST_NAME, customer.FirstName)
				col.SetString(CUSTOMER.LAST_NAME, customer.LastName)
				col.SetString(CUSTOMER.EMAIL, customer.Email)
				col.SetInt(CUSTOMER.ADDRESS_ID, customer.AddressID)
			}
			return nil
		}).
		OnConflict().DoNothing().
		Returning(CUSTOMER.CUSTOMER_ID),
		func(row *Row) {
			customerID := row.Int(CUSTOMER.CUSTOMER_ID)
			row.Process(func() { customerIDs = append(customerIDs, customerID) })
		},
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if int(rowCount) != len(customers)-4 {
		t.Fatal(testutil.Callers(), "expected %d rows inserted but got %d", len(customers)-4, rowsAffected)
	}
	for i := 4; i < len(customers); i++ {
		customers[i].CustomerID = customerIDs[i]
	}

	predicate = And()
	for _, customer := range customers[4:] {
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
		t.Fatal(testutil.Callers(), "expected inserted customers %+v to exist", customers[:4])
	}
}
