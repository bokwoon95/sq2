package sq

import (
	"strings"
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
)

func Test_MySQLInsertQuery(t *testing.T) {
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
		ACTOR := xNEW_ACTOR("")
		tt.item = MySQL.
			InsertInto(ACTOR).
			Columns(ACTOR.FIRST_NAME, ACTOR.LAST_NAME).
			Values("bob", "the builder").
			Values("alice", "in wonderland").
			AsRow("NEW").AsColumns("fname", "lname")
		tt.wantQuery = "INSERT INTO actor (first_name, last_name)" +
			" VALUES (?, ?), (?, ?) AS NEW (fname, lname)"
		tt.wantArgs = []interface{}{"bob", "the builder", "alice", "in wonderland"}
		assert(t, tt)
	})

	t.Run("INSERT ignore duplicates (INSERT IGNORE)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("")
		tt.item = MySQL.
			InsertIgnoreInto(ACTOR).
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
			})
		tt.wantQuery = "INSERT IGNORE INTO actor (actor_id, first_name, last_name)" +
			" VALUES (?, ?, ?), (?, ?, ?)"
		tt.wantArgs = []interface{}{int64(1), "bob", "the builder", int64(2), "alice", "in wonderland"}
		assert(t, tt)
	})

	t.Run("INSERT ignore duplicates (ON DUPLICATE KEY UPDATE)", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("")
		tt.item = MySQL.
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
			OnDuplicateKeyUpdate(
				AssignSelf(ACTOR.FIRST_NAME),
				AssignSelf(ACTOR.LAST_NAME),
			)
		tt.wantQuery = "INSERT INTO actor (actor_id, first_name, last_name)" +
			" VALUES (?, ?, ?), (?, ?, ?)" +
			" ON DUPLICATE KEY UPDATE first_name = first_name, last_name = last_name"
		tt.wantArgs = []interface{}{int64(1), "bob", "the builder", int64(2), "alice", "in wonderland"}
		assert(t, tt)
	})

	t.Run("upsert", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR := xNEW_ACTOR("")
		tt.item = MySQL.
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
			AsRow("NEW").
			OnDuplicateKeyUpdate(
				AssignAlias(ACTOR.FIRST_NAME, "NEW"),
				AssignAlias(ACTOR.LAST_NAME, "NEW"),
			)
		tt.wantQuery = "INSERT INTO actor (actor_id, first_name, last_name)" +
			" VALUES (?, ?, ?), (?, ?, ?) AS NEW" +
			" ON DUPLICATE KEY UPDATE first_name = NEW.first_name, last_name = NEW.last_name"
		tt.wantArgs = []interface{}{int64(1), "bob", "the builder", int64(2), "alice", "in wonderland"}
		assert(t, tt)
	})

	// TODO: mysql docs say to use a derived table to avoid using VALUES() for
	// ON DUPLICATE KEY UPDATE. Can I just directly alias my SELECT-ed fields
	// instead? Need to test it out. Not important though because I'm not going
	// to be using it here.
	t.Run("INSERT from SELECT", func(t *testing.T) {
		t.Parallel()
		var tt TT
		ACTOR1, ACTOR2 := xNEW_ACTOR(""), xNEW_ACTOR("a2")
		tt.item = MySQL.
			InsertInto(ACTOR1).
			Columns(ACTOR1.FIRST_NAME, ACTOR1.LAST_NAME).
			Select(MySQL.
				Select(ACTOR2.FIRST_NAME, ACTOR2.LAST_NAME).
				From(ACTOR2).
				Where(ACTOR2.ACTOR_ID.In([]int64{1, 2})),
			)
		tt.wantQuery = "INSERT INTO actor (first_name, last_name)" +
			" SELECT a2.first_name, a2.last_name" +
			" FROM actor AS a2" +
			" WHERE a2.actor_id IN (?, ?)"
		tt.wantArgs = []interface{}{int64(1), int64(2)}
		assert(t, tt)
	})
}

func TestMySQLSakilaInsert(t *testing.T) {
	if testing.Short() {
		return
	}
	tx, err := mysqlDB.Begin()
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	defer tx.Rollback()
	CUSTOMER := xNEW_CUSTOMER("")
	regina := Customer{StoreID: 1, AddressID: 1, FirstName: "REGINA", LastName: "TATE", Email: "regina_tate@email.com"}

	ensureCustomerExists := func(customer Customer) {
		exists, err := FetchExists(Log(tx), MySQL.From(CUSTOMER).Where(
			CUSTOMER.CUSTOMER_ID.EqInt(customer.CustomerID),
			CUSTOMER.STORE_ID.EqInt(customer.StoreID),
			CUSTOMER.FIRST_NAME.EqString(customer.FirstName),
			CUSTOMER.LAST_NAME.EqString(customer.LastName),
			CUSTOMER.EMAIL.EqString(customer.Email),
			CUSTOMER.ADDRESS_ID.EqInt(customer.AddressID),
		))
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if !exists {
			t.Fatalf(testutil.Callers()+"expected inserted customer %+v to exist", customer)
		}
	}

	// add regina
	rowsAffected, lastInsertID, err := Exec(Log(tx), MySQL.
		InsertInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			col.SetInt(CUSTOMER.STORE_ID, regina.StoreID)
			col.SetString(CUSTOMER.FIRST_NAME, regina.FirstName)
			col.SetString(CUSTOMER.LAST_NAME, regina.LastName)
			col.SetString(CUSTOMER.EMAIL, regina.Email)
			col.SetInt(CUSTOMER.ADDRESS_ID, regina.AddressID)
			return nil
		}),
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 1 {
		t.Fatal(testutil.Callers(), "expected 1 row to be affected but got %d", rowsAffected)
	}
	regina.CustomerID = int(lastInsertID)

	// ensure regina exists
	ensureCustomerExists(regina)

	// add regina again (using INSERT IGNORE) and check that nothing got added
	rowsAffected, lastInsertID, err = Exec(Log(tx), MySQL.
		InsertIgnoreInto(CUSTOMER).
		Valuesx(func(col *Column) error {
			col.SetInt(CUSTOMER.CUSTOMER_ID, regina.CustomerID)
			col.SetInt(CUSTOMER.STORE_ID, regina.StoreID)
			col.SetString(CUSTOMER.FIRST_NAME, regina.FirstName)
			col.SetString(CUSTOMER.LAST_NAME, regina.LastName)
			col.SetString(CUSTOMER.EMAIL, regina.Email)
			col.SetInt(CUSTOMER.ADDRESS_ID, regina.AddressID)
			return nil
		}),
	)
	if rowsAffected != 0 {
		t.Fatal(testutil.Callers(), "expected an second identical insert to affect 0 rows, got %d instead", rowsAffected)
	}

	// add regina again (using ON DUPLICATE KEY UPDATE with self assignment) and check that nothing got added
	rowsAffected, lastInsertID, err = Exec(Log(tx), MySQL.
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
		OnDuplicateKeyUpdate(AssignSelf(CUSTOMER.CUSTOMER_ID)),
	)
	if rowsAffected != 0 {
		t.Fatal(testutil.Callers(), "expected an second identical insert to affect 0 rows, got %d instead", rowsAffected)
	}

	// modify and upsert regina (using VALUES)
	regina.FirstName = regina.FirstName[:1] + strings.ToLower(regina.FirstName[1:])
	regina.LastName = regina.LastName[:1] + strings.ToLower(regina.LastName[1:])
	rowsAffected, lastInsertID, err = Exec(Log(tx), MySQL.
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
		OnDuplicateKeyUpdate(
			AssignValues(CUSTOMER.STORE_ID),
			AssignValues(CUSTOMER.FIRST_NAME),
			AssignValues(CUSTOMER.LAST_NAME),
			AssignValues(CUSTOMER.EMAIL),
			AssignValues(CUSTOMER.ADDRESS_ID),
		),
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	// for some reason rowsAffected is always 2, I suspect because MySQL first
	// does an INSERT followed by an UPDATE so it counts as 2 affected rows
	if rowsAffected != 2 {
		t.Fatal(testutil.Callers(), "expected 2 row to be upserted but got %d", rowsAffected)
	}

	// ensure the modified regina exists
	ensureCustomerExists(regina)

	// modify and upsert regina (using MySQL 8 row aliases)
	regina.FirstName = strings.ToUpper(regina.FirstName)
	regina.LastName = strings.ToUpper(regina.LastName)
	rowsAffected, lastInsertID, err = Exec(Log(tx), MySQL.
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
		AsRow("NEW").
		OnDuplicateKeyUpdate(
			AssignAlias(CUSTOMER.STORE_ID, "NEW"),
			AssignAlias(CUSTOMER.FIRST_NAME, "NEW"),
			AssignAlias(CUSTOMER.LAST_NAME, "NEW"),
			AssignAlias(CUSTOMER.EMAIL, "NEW"),
			AssignAlias(CUSTOMER.ADDRESS_ID, "NEW"),
		),
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 2 {
		t.Fatal(testutil.Callers(), "expected 2 row to be upserted but got %d", rowsAffected)
	}

	// ensure the modified regina exists
	ensureCustomerExists(regina)

	// modify and upsert regina (using MySQL 8 row aliases as well as column aliases)
	regina.FirstName = regina.FirstName[:1] + strings.ToLower(regina.FirstName[1:])
	regina.LastName = regina.LastName[:1] + strings.ToLower(regina.LastName[1:])
	rowsAffected, lastInsertID, err = Exec(Log(tx), MySQL.
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
		AsRow("NEW").AsColumns("cid", "sid", "fname", "lname", "email", "aid").
		OnDuplicateKeyUpdate(
			Assign(CUSTOMER.STORE_ID, Literal("sid")),
			Assign(CUSTOMER.FIRST_NAME, Literal("fname")),
			Assign(CUSTOMER.LAST_NAME, Literal("lname")),
			Assign(CUSTOMER.EMAIL, Literal("NEW.email")),
			Assign(CUSTOMER.ADDRESS_ID, Literal("aid")),
		),
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}
	if rowsAffected != 2 {
		t.Fatal(testutil.Callers(), "expected 2 row to be upserted but got %d", rowsAffected)
	}

	// ensure the modified regina exists
	ensureCustomerExists(regina)

	// Customer 'MARY SMITH' rents the film 'ACADEMY DINOSAUR' from staff 'Mike
	// Hillyer' at Store 1 on 9th of August 2021 4pm. Insert a rental record
	// representing that transaction (INSERT with SELECT).
	STAFF := xNEW_STAFF("")
	FILM := xNEW_FILM("")
	INVENTORY := xNEW_INVENTORY("")
	STORE := xNEW_STORE("")
	RENTAL := xNEW_RENTAL("")
	customer_id := NewSubquery("customer_id", MySQL.
		Select(CUSTOMER.CUSTOMER_ID).
		From(CUSTOMER).
		Where(RowValue{CUSTOMER.FIRST_NAME, CUSTOMER.LAST_NAME}.Eq(RowValue{"MARY", "SMITH"})).
		Limit(1),
	)
	staff_id := NewSubquery("staff_id", MySQL.
		Select(STAFF.STAFF_ID).
		From(STAFF).
		Where(
			STAFF.STORE_ID.EqInt(1),
			RowValue{STAFF.FIRST_NAME, STAFF.LAST_NAME}.Eq(RowValue{"Mike", "Hillyer"}),
		).
		Limit(1),
	)
	_, rentalID, err := Exec(Log(tx), MySQL.
		InsertInto(RENTAL).
		Columns(RENTAL.INVENTORY_ID, RENTAL.CUSTOMER_ID, RENTAL.STAFF_ID, RENTAL.RENTAL_DATE).
		Select(MySQL.
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
				Not(Exists(MySQL.
					SelectOne().
					From(RENTAL).
					Where(RENTAL.INVENTORY_ID.Eq(INVENTORY.INVENTORY_ID), RENTAL.RENTAL_DATE.IsNull()),
				)),
			).
			OrderBy(INVENTORY.INVENTORY_ID).
			Limit(1),
		),
	)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}

	// check that the rentalID returned has the attributes we want
	exists, err := FetchExists(Log(tx), MySQL.
		From(RENTAL).
		Join(CUSTOMER, CUSTOMER.CUSTOMER_ID.Eq(RENTAL.CUSTOMER_ID)).
		Join(STAFF, STAFF.STAFF_ID.Eq(RENTAL.STAFF_ID)).
		Join(INVENTORY, INVENTORY.INVENTORY_ID.Eq(RENTAL.INVENTORY_ID)).
		Join(FILM, FILM.FILM_ID.Eq(INVENTORY.FILM_ID)).
		Where(
			RENTAL.RENTAL_ID.EqInt64(rentalID),
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
