package sq

import (
	"bytes"
	"fmt"
)

type CTE2 struct {
	query          Query
	initErr        error
	fetchFieldsErr error
	initErrMsg     string
	cteName        string
	cteAlias       string
	isRecursive    bool
}

var _ Table = CTE2{}

func (cte CTE2) GetAlias() string { return cte.cteAlias }

func (cte CTE2) GetName() string { return cte.cteName }

func (cte CTE2) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if cte.initErrMsg != "" {
		return fmt.Errorf("sq: %s", cte.initErrMsg)
	} else if cte.fetchFieldsErr != nil {
		return fmt.Errorf("sq: CTE %s", cte.initErrMsg)
	}
	buf.WriteString(cte.cteName)
	return nil
}

func newCTE2(recursive bool, name string, columns []string, query Query) CTE2 {
	var cte CTE2
	if name == "" {
		cte.initErr = fmt.Errorf("sq: CTE name cannot be empty")
		return cte
	}
	if query == nil {
		cte.initErr = fmt.Errorf("sq: CTE query cannot be nil")
		return cte
	}
	fieldNames := make([]string, len(columns))
	copy(fieldNames, columns)
	if len(fieldNames) == 0 {
		fields, err := query.GetFetchableFields()
		if err != nil {
			cte.fetchFieldsErr = err
			return cte
		}
		if len(fields) == 0 {
			cte.initErr = fmt.Errorf("sq: CTE query does not return any fields")
			return cte
		}
	}
	return cte
}

type CTEField2 struct {
	cteErrMsg error
	valid     bool
	info      FieldInfo
}

var _ Field = CTEField2{}

func (f CTEField2) GetAlias() string { return f.info.FieldAlias }

func (f CTEField2) GetName() string { return f.info.FieldName }

func (f CTEField2) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	if !f.valid {
		tableQualifier := f.info.TableName
		if f.info.TableAlias != "" {
			tableQualifier = f.info.TableAlias
		}
		if f.cteErrMsg != nil {
			return fmt.Errorf("sq: CTE field %s.%s is invalid because the CTE is invalid: %w", tableQualifier, f.info.FieldName, f.cteErrMsg)
		} else {
			return fmt.Errorf("sq: CTE field %s.%s does not exist", tableQualifier, f.info.FieldName)
		}
	}
	return f.info.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
}

func (f CTEField2) As(alias string) CTEField2 {
	f.info.FieldAlias = alias
	return f
}

func (f CTEField2) Asc() CTEField2 {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = false
	return f
}

func (f CTEField2) Desc() CTEField2 {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = true
	return f
}

func (f CTEField2) NullsLast() CTEField2 {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = false
	return f
}

func (f CTEField2) NullsFirst() CTEField2 {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = true
	return f
}

func (f CTEField2) IsNull() Predicate { return IsNull(f) }

func (f CTEField2) IsNotNull() Predicate { return IsNotNull(f) }

func (f CTEField2) In(v interface{}) Predicate { return In(f, v) }

func (f CTEField2) Eq(v interface{}) Predicate { return Eq(f, v) }

func (f CTEField2) Ne(v interface{}) Predicate { return Ne(f, v) }

func (f CTEField2) Gt(v interface{}) Predicate { return Gt(f, v) }

func (f CTEField2) Ge(v interface{}) Predicate { return Ge(f, v) }

func (f CTEField2) Lt(v interface{}) Predicate { return Lt(f, v) }

func (f CTEField2) Le(v interface{}) Predicate { return Le(f, v) }
