package sq

import (
	"bytes"
	"testing"

	"github.com/bokwoon95/testutil"
)

func Test_JoinTable(t *testing.T) {
	tableInfo := TableInfo{TableName: "users"}
	USERS := struct {
		TableInfo
		USER_ID CustomField
		NAME    CustomField
		EMAIL   CustomField
		AGE     CustomField
	}{
		TableInfo: tableInfo,
		USER_ID:   NewCustomField("user_id", tableInfo),
		NAME:      NewCustomField("name", tableInfo),
		EMAIL:     NewCustomField("email", tableInfo),
		AGE:       NewCustomField("age", tableInfo),
	}

	type TT struct {
		dialect   string
		item      SQLAppender
		wantQuery string
		wantArgs  []interface{}
	}

	assert := func(t *testing.T, tt TT) {
		is := testutil.New(t, testutil.Parallel)
		buf := bufpool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			bufpool.Put(buf)
		}()
		gotArgs, gotParams := []interface{}{}, map[string][]int{}
		tt.item.AppendSQL(tt.dialect, buf, &gotArgs, gotParams)
		is.Equal(tt.wantQuery, buf.String())
		is.Equal(tt.wantArgs, gotArgs)
	}

	t.Run("join", func(t *testing.T) {
		var tt TT
		tt.item = Join(USERS, Eq(1, 1))
		tt.wantQuery = "JOIN users ON ? = ?"
		tt.wantArgs = []interface{}{1, 1}
		assert(t, tt)
	})

	t.Run("left join", func(t *testing.T) {
		var tt TT
		tt.item = LeftJoin(USERS, Eq(1, 1))
		tt.wantQuery = "LEFT JOIN users ON ? = ?"
		tt.wantArgs = []interface{}{1, 1}
		assert(t, tt)
	})

	t.Run("right join", func(t *testing.T) {
		var tt TT
		tt.item = RightJoin(USERS, Eq(1, 1))
		tt.wantQuery = "RIGHT JOIN users ON ? = ?"
		tt.wantArgs = []interface{}{1, 1}
		assert(t, tt)
	})

	t.Run("full join", func(t *testing.T) {
		var tt TT
		tt.item = FullJoin(USERS, Eq(1, 1))
		tt.wantQuery = "FULL JOIN users ON ? = ?"
		tt.wantArgs = []interface{}{1, 1}
		assert(t, tt)
	})

	t.Run("cross join", func(t *testing.T) {
		var tt TT
		tt.item = CrossJoin(USERS)
		tt.wantQuery = "CROSS JOIN users"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("implicit join", func(t *testing.T) {
		var tt TT
		USERS := USERS
		USERS.TableAlias = "u"
		tt.item = ImplicitJoin(USERS)
		tt.wantQuery = ", users AS u"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("custom join", func(t *testing.T) {
		var tt TT
		tt.item = CustomJoin("CROSS JOIN LATERAL", Tablef("unnest({}) WITH ORDINALITY AS uhh(email, seqno)", USERS.EMAIL))
		tt.wantQuery = "CROSS JOIN LATERAL unnest(users.email) WITH ORDINALITY AS uhh(email, seqno)"
		tt.wantArgs = []interface{}{}
		assert(t, tt)
	})

	t.Run("all joins", func(t *testing.T) {
		var tt TT
		tt.item = JoinTables{
			Join(USERS, Eq(1, 1)),
			LeftJoin(USERS, Eq(1, 1)),
			RightJoin(USERS, Eq(1, 1)),
			FullJoin(USERS, Eq(1, 1)),
			CrossJoin(USERS),
			ImplicitJoin(USERS),
			CustomJoin("CROSS JOIN LATERAL", Tablef("unnest({}) WITH ORDINALITY AS uhh(email, seqno)", USERS.EMAIL)),
		}
		tt.wantQuery = "JOIN users ON ? = ?" +
			" LEFT JOIN users ON ? = ?" +
			" RIGHT JOIN users ON ? = ?" +
			" FULL JOIN users ON ? = ?" +
			" CROSS JOIN users" +
			" , users" +
			" CROSS JOIN LATERAL unnest(users.email) WITH ORDINALITY AS uhh(email, seqno)"
		tt.wantArgs = []interface{}{1, 1, 1, 1, 1, 1, 1, 1}
		assert(t, tt)
	})
}
