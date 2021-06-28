package sq_test

import (
	"testing"

	. "github.com/bokwoon95/sq"
)

func Test_SQLiteUpdateQuery(t *testing.T) {
	type TT struct {
		dialect   string
		item      Query
		wantQuery string
		wantArgs  []interface{}
	}
}
