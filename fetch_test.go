package sq

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/bokwoon95/sq/internal/testutil"
)

func TestSQLiteFetch(t *testing.T) {
	db, err := sql.Open("sqlite3", *sqliteDSNFlag)
	if err != nil {
		t.Fatal(testutil.Callers(), err)
	}

	t.Run("SELECT DISTINCT", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		var gotLastNames []string
		rowCount, err := Fetch(Log(db), SQLite(nil).SelectDistinct().From(ACTOR).OrderBy(ACTOR.LAST_NAME).Limit(4), func(row *Row) {
			lastName := row.String(ACTOR.LAST_NAME)
			row.Process(func() error {
				gotLastNames = append(gotLastNames, lastName)
				return nil
			})
		})
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if rowCount != 4 {
			t.Fatal(testutil.Callers(), fmt.Sprintf("got=%d want=%d", rowCount, 4))
		}
		wantLastNames := []string{"AKROYD", "ALLEN", "ASTAIRE", "BACALL"}
		if diff := testutil.Diff(gotLastNames, wantLastNames); diff != "" {
			t.Fatal(testutil.Callers(), diff)
		}
	})

	t.Run("SELECT EXISTS", func(t *testing.T) {
		t.Parallel()
		ACTOR := xNEW_ACTOR("")
		exists, err := FetchExists(Log(db), SQLite(nil).From(ACTOR).Where(Or(
			ACTOR.FIRST_NAME.EqString("SCARLETT"),
			ACTOR.LAST_NAME.EqString("JOHANSSON"),
		)))
		if err != nil {
			t.Fatal(testutil.Callers(), err)
		}
		if !exists {
			t.Fatal(testutil.Callers(), fmt.Sprintf("got=%v want=%v", exists, true))
		}
	})
}
