package sq

import (
	"bytes"
	"fmt"
	"strings"
)

type CTEField struct {
	stickyErr error
	valid     bool
	info      FieldInfo
}

var _ Field = CTEField{}

func (f CTEField) GetAlias() string { return f.info.FieldAlias }

func (f CTEField) GetName() string { return f.info.FieldName }

func (f CTEField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	if !f.valid {
		tableQualifier := f.info.TableName
		if f.info.TableAlias != "" {
			tableQualifier = f.info.TableAlias
		}
		if f.stickyErr != nil {
			return fmt.Errorf("sq: CTE field %s.%s invalid due to CTE error: %w", tableQualifier, f.info.FieldName, f.stickyErr)
		} else {
			return fmt.Errorf("sq: CTE field %s.%s does not exist", tableQualifier, f.info.FieldName)
		}
	}
	return f.info.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
}

func (f CTEField) As(alias string) CTEField {
	f.info.FieldAlias = alias
	return f
}

func (f CTEField) Asc() CTEField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = false
	return f
}

func (f CTEField) Desc() CTEField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = true
	return f
}

func (f CTEField) NullsLast() CTEField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = false
	return f
}

func (f CTEField) NullsFirst() CTEField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = true
	return f
}

func (f CTEField) IsNull() Predicate { return IsNull(f) }

func (f CTEField) IsNotNull() Predicate { return IsNotNull(f) }

func (f CTEField) In(v interface{}) Predicate { return In(f, v) }

func (f CTEField) Eq(v interface{}) Predicate { return Eq(f, v) }

func (f CTEField) Ne(v interface{}) Predicate { return Ne(f, v) }

func (f CTEField) Gt(v interface{}) Predicate { return Gt(f, v) }

func (f CTEField) Ge(v interface{}) Predicate { return Ge(f, v) }

func (f CTEField) Lt(v interface{}) Predicate { return Lt(f, v) }

func (f CTEField) Le(v interface{}) Predicate { return Le(f, v) }

type CTE struct {
	query       Query
	stickyErr   error
	isRecursive bool
	cteName     string
	cteAlias    string
	columns     []string
	fieldNames  map[string]struct{}
}

var _ Table = CTE{}

func (cte CTE) GetAlias() string { return cte.cteAlias }

func (cte CTE) GetName() string { return cte.cteName }

func (cte CTE) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if cte.stickyErr != nil {
		return cte.stickyErr
	}
	buf.WriteString(cte.cteName)
	return nil
}

func NewCTE(name string, columns []string, query Query) CTE {
	return newCTE(false, name, columns, query)
}

func NewRecursiveCTE(name string, columns []string, query Query) CTE {
	return newCTE(true, name, columns, query)
}

func newCTE(recursive bool, name string, columns []string, query Query) CTE {
	cte := CTE{
		query:      query,
		cteName:    name,
		columns:    columns,
		fieldNames: make(map[string]struct{}),
	}
	if name == "" {
		cte.stickyErr = fmt.Errorf("sq: CTE name cannot be empty")
		return cte
	}
	if query == nil {
		cte.stickyErr = fmt.Errorf("sq: CTE query cannot be nil")
		return cte
	}
	fieldNames := make([]string, len(columns))
	copy(fieldNames, columns)
	if len(fieldNames) == 0 {
		fields, err := query.GetFetchableFields()
		if err != nil {
			cte.stickyErr = err
			return cte
		}
		if len(fields) == 0 {
			cte.stickyErr = fmt.Errorf("sq: CTE %s does not return any fields", name)
			return cte
		}
		for i, field := range fields {
			fieldName := field.GetAlias()
			if fieldName == "" {
				fieldName = field.GetName()
			}
			if fieldName == "" {
				cte.stickyErr = fmt.Errorf("sq: CTE %s field #%d needs a name or an alias", name, i+1)
				return cte
			}
			fieldNames = append(fieldNames, fieldName)
		}
	}
	for _, fieldName := range fieldNames {
		cte.fieldNames[fieldName] = struct{}{}
	}
	return cte
}

func (cte CTE) As(alias string) CTE {
	if cte.stickyErr != nil {
		return cte
	}
	if cte.cteName == "" {
		cte.stickyErr = fmt.Errorf("sq: CTE name cannot be empty")
		return cte
	}
	if cte.query == nil {
		cte.stickyErr = fmt.Errorf("sq: CTE query cannot be nil")
		return cte
	}
	if len(cte.fieldNames) == 0 {
		cte.stickyErr = fmt.Errorf("sq: CTE %s does not return any fields", cte.cteName)
		return cte
	}
	cte.cteAlias = alias
	return cte
}

func (cte CTE) Field(fieldName string) CTEField {
	_, ok := cte.fieldNames[fieldName]
	return CTEField{
		stickyErr: cte.stickyErr,
		valid:     ok,
		info: FieldInfo{
			TableName:  cte.cteName,
			TableAlias: cte.cteAlias,
			FieldName:  fieldName,
		},
	}
}

type CTEs []CTE

var _ SQLAppender = CTEs{}

func (ctes CTEs) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	var hasRecursiveCTE bool
	for _, cte := range ctes {
		if cte.isRecursive {
			hasRecursiveCTE = true
			break
		}
	}
	if hasRecursiveCTE {
		buf.WriteString("WITH RECURSIVE ")
	} else {
		buf.WriteString("WITH ")
	}
	for i, cte := range ctes {
		if i > 0 {
			buf.WriteString(", ")
		}
		if cte.cteName == "" {
			return fmt.Errorf("sq: CTE #%d has no name", i+1)
		}
		buf.WriteString(cte.cteName)
		if len(cte.columns) > 0 {
			buf.WriteString(" (")
			buf.WriteString(strings.Join(cte.columns, ", "))
			buf.WriteString(")")
		}
		buf.WriteString(" AS (")
		switch query := cte.query.(type) {
		case nil:
			return fmt.Errorf("sq: CTE #%d has no query", i+1)
		case VariadicQuery:
			query.TopLevel = true
			err := query.AppendSQL(dialect, buf, args, params)
			if err != nil {
				return fmt.Errorf("sq: CTE #%d failed to build query: %w", i+1, err)
			}
		default:
			err := query.AppendSQL(dialect, buf, args, params)
			if err != nil {
				return fmt.Errorf("sq: CTE #%d failed to build query: %w", i+1, err)
			}
		}
	}
	buf.WriteString(") ")
	return nil
}
