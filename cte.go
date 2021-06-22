package sq

import (
	"bytes"
	"fmt"
	"strings"
)

type CTEField struct {
	valid        bool
	query        Query
	cteRecursive bool
	cteName      string
	cteAlias     string
	columns      []string
	fieldName    string
	fieldAlias   string
}

var _ Field = CTEField{}

func (f CTEField) GetAlias() string { return f.fieldAlias }

func (f CTEField) GetName() string { return f.fieldName }

func (f CTEField) As(alias string) CTEField {
	f.fieldAlias = alias
	return f
}

func (f CTEField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	if !f.valid {
		return fmt.Errorf("sq: referenced nonexistent CTEField")
	}
	if f.fieldName == "" {
		return fmt.Errorf("sq: CTEField has no name")
	}
	tableQualifier := f.cteName
	if f.cteAlias != "" {
		tableQualifier = f.cteAlias
	}
	buf.WriteString(tableQualifier + "." + f.fieldName)
	return nil
}

type CTE map[string]CTEField

var _ Table = CTE{}

func newCTE(recursive bool, name string, columns []string, query Query) (CTE, error) {
	if name == "" {
		return nil, fmt.Errorf("sq: CTE name cannot be empty")
	}
	if query == nil {
		return nil, fmt.Errorf("sq: CTE query cannot be nil")
	}
	cte := CTE{"": {
		query:        query,
		cteRecursive: recursive,
		cteName:      name,
		columns:      columns,
	}}
	fieldNames := make([]string, len(columns))
	copy(fieldNames, columns)
	if len(fieldNames) == 0 {
		fields, err := query.GetFetchableFields()
		if len(fields) == 0 {
			return nil, fmt.Errorf("sq: CTE query does not return any fields")
		}
		if err != nil {
			return nil, fmt.Errorf("sq: error fetching fields for CTE: %w", err)
		}
		for i, field := range fields {
			if field == nil {
				return nil, fmt.Errorf("sq: field #%d in CTE query is nil", i+1)
			}
			fieldName := field.GetAlias()
			if fieldName == "" {
				fieldName = field.GetName()
			}
			if fieldName == "" {
				return nil, fmt.Errorf("sq: field #%d in CTE query has no name and no alias", i+1)
			}
			fieldNames = append(fieldNames, fieldName)
		}
	}
	for _, fieldName := range fieldNames {
		cte[fieldName] = CTEField{
			valid:     true,
			cteName:   name,
			fieldName: fieldName,
		}
	}
	return cte, nil
}

func NewCTE(name string, columns []string, query Query) (CTE, error) {
	return newCTE(false, name, columns, query)
}

func NewRecursiveCTE(name string, columns []string, query Query) (CTE, error) {
	return newCTE(true, name, columns, query)
}

func (cte CTE) GetAlias() string { return cte[""].cteAlias }

func (cte CTE) GetName() string { return cte[""].cteName }

func (cte CTE) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if len(cte) == 0 {
		return fmt.Errorf("sq: empty CTE")
	}
	metafield, ok := cte[""]
	if !ok {
		return fmt.Errorf("sq: empty CTE")
	}
	if metafield.cteName == "" {
		return fmt.Errorf("sq: CTE has no name")
	}
	buf.WriteString(metafield.cteName)
	return nil
}

func (cte CTE) As(alias string) CTE {
	metafield := cte[""]
	metafield.cteAlias = alias
	cte2 := CTE{"": metafield}
	for fieldName := range cte {
		if fieldName == "" {
			continue
		}
		cte2[fieldName] = CTEField{
			cteName:   metafield.cteName,
			fieldName: fieldName,
		}
	}
	return cte2
}

type CTEs []CTE

var _ SQLAppender = CTEs{}

func (ctes CTEs) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	var hasRecursiveCTE bool
	for _, cte := range ctes {
		if cte[""].cteRecursive {
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
		metafield := cte[""]
		if metafield.cteName == "" {
			return fmt.Errorf("CTE #%d has no name", i+1)
		}
		buf.WriteString(metafield.cteName)
		if len(metafield.columns) > 0 {
			buf.WriteString(" (")
			buf.WriteString(strings.Join(metafield.columns, ", "))
			buf.WriteString(")")
		}
		buf.WriteString(" AS (")
		switch query := metafield.query.(type) {
		case nil:
			return fmt.Errorf("CTE #%d has no query", i+1)
		case VariadicQuery:
			query.TopLevel = true
			err := query.AppendSQL(dialect, buf, args, params)
			if err != nil {
				return fmt.Errorf("CTE #%d failed to build query: %w", i+1, err)
			}
		default:
			err := query.AppendSQL(dialect, buf, args, params)
			if err != nil {
				return fmt.Errorf("CTE #%d failed to build query: %w", i+1, err)
			}
		}
	}
	buf.WriteString(" ")
	return nil
}
