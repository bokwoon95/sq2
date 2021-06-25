package sq

import (
	"testing"
	"time"
)

func Test_Column(t *testing.T) {
	USERS := struct {
		USER_ID    NumberField
		IS_ACTIVE  BooleanField
		NAME       StringField
		AGE        NumberField
		SCORE      NumberField
		CREATED_AT TimeField
	}{
		USER_ID:    NewNumberField("user_id", TableInfo{}),
		IS_ACTIVE:  NewBooleanField("is_active", TableInfo{}),
		NAME:       NewStringField("name", TableInfo{}),
		AGE:        NewNumberField("age", TableInfo{}),
		SCORE:      NewNumberField("score", TableInfo{}),
		CREATED_AT: NewTimeField("created_at", TableInfo{}),
	}

	t.Run("ColumnModeInsert", func(t *testing.T) {
		col := NewColumn(ColumnModeInsert)
		users := []struct {
			UserID    int64
			IsActive  bool
			Name      string
			Age       int
			Score     float64
			CreatedAt time.Time
		}{
			{UserID: 1, IsActive: true, Name: "bob", Age: 27, Score: 89.9, CreatedAt: time.Unix(0, 0)},
			{UserID: 2, IsActive: true, Name: "alice", Age: 24, Score: 90.0, CreatedAt: time.Unix(1, 0)},
			{UserID: 3, IsActive: true, Name: "mallory", Age: 35, Score: 80.0, CreatedAt: time.Unix(2, 0)},
			{UserID: 4, IsActive: false, Name: "eve", Age: 21, Score: 79.9, CreatedAt: time.Unix(3, 0)},
		}
		columnmapper := func(col *Column) error {
			for _, user := range users {
				col.SetInt64(USERS.USER_ID, user.UserID)
				col.SetBool(USERS.IS_ACTIVE, user.IsActive)
				col.SetString(USERS.NAME, user.Name)
				col.SetInt(USERS.AGE, user.Age)
				col.SetFloat64(USERS.SCORE, user.Score)
				col.SetTime(USERS.CREATED_AT, user.CreatedAt)
			}
			return nil
		}
		columnmapper(col)
		gotFields, gotRowValues := ColumnInsertResult(col)
		wantFields := Fields{USERS.USER_ID, USERS.IS_ACTIVE, USERS.NAME, USERS.AGE, USERS.SCORE, USERS.CREATED_AT}
		wantRowValues := RowValues{
			{int64(1), true, "bob", 27, 89.9, time.Unix(0, 0)},
			{int64(2), true, "alice", 24, 90.0, time.Unix(1, 0)},
			{int64(3), true, "mallory", 35, 80.0, time.Unix(2, 0)},
			{int64(4), false, "eve", 21, 79.9, time.Unix(3, 0)},
		}
		if diff := testdiff(gotFields, wantFields); diff != "" {
			t.Error(testcallers(), diff)
		}
		if diff := testdiff(gotRowValues, wantRowValues); diff != "" {
			t.Error(testcallers(), diff)
		}
	})
}
