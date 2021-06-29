package sq

import (
	"bytes"
	"fmt"
)

type RowValue []interface{}

func (r RowValue) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	buf.WriteString("(")
	var err error
	for i, value := range r {
		if i > 0 {
			buf.WriteString(", ")
		}
		err = BufferPrintValue(dialect, buf, args, params, excludedTableQualifiers, value, "")
		if err != nil {
			return fmt.Errorf("rowvalue #%d: %w", i+1, err)
		}
	}
	buf.WriteString(")")
	return nil
}

func (r RowValue) GetName() string { return "" }

func (r RowValue) GetAlias() string { return "" }

func (r RowValue) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	return r.AppendSQLExclude(dialect, buf, args, params, nil)
}

func (r RowValue) In(v interface{}) Predicate { return In(r, v) }

type RowValues []RowValue

func (rs RowValues) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	var err error
	for i, r := range rs {
		if i > 0 {
			buf.WriteString(", ")
		}
		err = r.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return fmt.Errorf("rowvalues #%d: %w", i+1, err)
		}
	}
	return nil
}
