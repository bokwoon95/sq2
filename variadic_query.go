package sq

import (
	"bytes"
	"fmt"
)

type VariadicQueryOperator string

const (
	QueryUnion        VariadicQueryOperator = "UNION"
	QueryUnionAll     VariadicQueryOperator = "UNION ALL"
	QueryIntersect    VariadicQueryOperator = "INTERSECT"
	QueryIntersectAll VariadicQueryOperator = "INTERSECT ALL"
	QueryExcept       VariadicQueryOperator = "EXCEPT"
	QueryExceptAll    VariadicQueryOperator = "EXCEPT ALL"
)

type VariadicQuery struct {
	TopLevel bool
	Operator VariadicQueryOperator
	Queries  []Query
}

var _ SQLAppender = VariadicQuery{}

func (vq VariadicQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	var err error
	if vq.Operator == "" {
		vq.Operator = QueryUnion
	}
	if len(vq.Queries) == 0 {
		return nil
	}
	if len(vq.Queries) == 1 {
		switch q := vq.Queries[0].(type) {
		case nil:
			return fmt.Errorf("VariadicQuery query #1 is nil")
		case VariadicQuery:
			q.TopLevel = vq.TopLevel
			err = q.AppendSQL(dialect, buf, args, params, nil)
			if err != nil {
				return err
			}
		default:
			err = q.AppendSQL(dialect, buf, args, params, nil)
			if err != nil {
				return err
			}
		}
		return nil
	}
	if !vq.TopLevel {
		buf.WriteString("(")
	}
	for i, q := range vq.Queries {
		if i > 0 {
			buf.WriteString(" " + string(vq.Operator) + " ")
		}
		switch q := q.(type) {
		case nil:
			return fmt.Errorf("VariadicQuery query #%d is nil", i+1)
		case VariadicQuery:
			q.TopLevel = false
			err = q.AppendSQL(dialect, buf, args, params, nil)
			if err != nil {
				return fmt.Errorf("VariadicQuery query #%d: %w", i+1, err)
			}
		default:
			err = q.AppendSQL(dialect, buf, args, params, nil)
			if err != nil {
				return fmt.Errorf("VariadicQuery query #%d: %w", i+1, err)
			}
		}
	}
	if !vq.TopLevel {
		buf.WriteString(")")
	}
	return nil
}

func (vq VariadicQuery) SetFetchableFields(fields []Field) (Query, error) {
	return vq, ErrNonFetchableQuery
}

func (vq VariadicQuery) GetFetchableFields() ([]Field, error) {
	if len(vq.Queries) == 0 {
		return nil, nil
	}
	return vq.Queries[0].GetFetchableFields()
}

func (vq VariadicQuery) GetDialect() string {
	if len(vq.Queries) == 0 {
		return ""
	}
	return vq.Queries[0].GetDialect()
}

func Union(queries ...Query) VariadicQuery {
	return VariadicQuery{TopLevel: true, Operator: QueryUnion, Queries: queries}
}

func UnionAll(queries ...Query) VariadicQuery {
	return VariadicQuery{TopLevel: true, Operator: QueryUnionAll, Queries: queries}
}

func Intersect(queries ...Query) VariadicQuery {
	return VariadicQuery{TopLevel: true, Operator: QueryIntersect, Queries: queries}
}

func IntersectAll(queries ...Query) VariadicQuery {
	return VariadicQuery{TopLevel: true, Operator: QueryIntersectAll, Queries: queries}
}

func Except(queries ...Query) VariadicQuery {
	return VariadicQuery{TopLevel: true, Operator: QueryExcept, Queries: queries}
}

func ExceptAll(queries ...Query) VariadicQuery {
	return VariadicQuery{TopLevel: true, Operator: QueryExceptAll, Queries: queries}
}
