package sq

import (
	"bytes"
	"fmt"
)

func Not(predicate Predicate) Predicate {
	return predicate.Not()
}

type CustomPredicate struct {
	Alias    string
	Format   string
	Values   []interface{}
	Negative bool
}

var _ Predicate = CustomPredicate{}

func Predicatef(format string, values ...interface{}) CustomPredicate {
	return CustomPredicate{Format: format, Values: values}
}

func (p CustomPredicate) As(alias string) CustomPredicate {
	p.Alias = alias
	return p
}

func (p CustomPredicate) GetAlias() string { return p.Alias }

func (p CustomPredicate) GetName() string { return "" }

func (p CustomPredicate) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	if p.Negative {
		buf.WriteString("NOT ")
	}
	return BufferPrintf(dialect, buf, args, params, env, excludedTableQualifiers, p.Format, p.Values)
}

func (p CustomPredicate) Not() Predicate {
	p.Negative = !p.Negative
	return p
}

type VariadicPredicate struct {
	// Toplevel indicates if the variadic predicate is the top level predicate
	// i.e. it does not need enclosing brackets
	Toplevel   bool
	Alias      string
	Or         bool
	Predicates []Predicate
	Negative   bool
}

func (p VariadicPredicate) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	var err error
	switch len(p.Predicates) {
	case 0:
		return nil
	case 1:
		if p.Negative {
			buf.WriteString("NOT ")
		}
		switch v := p.Predicates[0].(type) {
		case nil:
			return fmt.Errorf("predicate #1 is nil")
		case VariadicPredicate:
			if !p.Toplevel {
				buf.WriteString("(")
			}
			v.Toplevel = true
			err = v.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
			if err != nil {
				return err
			}
			if !p.Toplevel {
				buf.WriteString(")")
			}
		default:
			err = p.Predicates[0].AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
			if err != nil {
				return err
			}
		}
	default:
		if p.Negative {
			buf.WriteString("NOT ")
		}
		if !p.Toplevel {
			buf.WriteString("(")
		}
		for i, predicate := range p.Predicates {
			if i > 0 {
				if p.Or {
					buf.WriteString(" OR ")
				} else {
					buf.WriteString(" AND ")
				}
			}
			if predicate == nil {
				return fmt.Errorf("predicate #%d is nil", i+1)
			}
			err = predicate.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
			if err != nil {
				return fmt.Errorf("predicate #%d: %w", i+1, err)
			}
		}
		if !p.Toplevel {
			buf.WriteString(")")
		}
	}
	return nil
}

func (p VariadicPredicate) Not() Predicate {
	p.Negative = !p.Negative
	return p
}

func (p VariadicPredicate) GetAlias() string { return p.Alias }

func (p VariadicPredicate) GetName() string { return "" }

func (p VariadicPredicate) Append(predicates ...Predicate) VariadicPredicate {
	p.Predicates = append(p.Predicates, predicates...)
	return p
}

func And(predicates ...Predicate) VariadicPredicate {
	return VariadicPredicate{Or: false, Predicates: predicates}
}

func Or(predicates ...Predicate) VariadicPredicate {
	return VariadicPredicate{Or: true, Predicates: predicates}
}

func IsNull(f Field) Predicate { return Predicatef("{} IS NULL", f) }

func IsNotNull(f Field) Predicate { return Predicatef("{} IS NOT NULL", f) }

// TODO: if any of the values are queries, wrap them in braces

func cmp(a, b interface{}, operator string) Predicate {
	_, ok1 := a.(Query)
	_, ok2 := b.(Query)
	if !ok1 && !ok2 {
		return Predicatef("{} "+operator+" {}", a, b)
	} else if !ok1 && ok2 {
		return Predicatef("{} "+operator+" ({})", a, b)
	} else if ok1 && !ok2 {
		return Predicatef("({}) "+operator+" {}", a, b)
	} else {
		return Predicatef("({}) "+operator+" ({})", a, b)
	}
}

func Eq(a, b interface{}) Predicate { return cmp(a, b, "=") }

func Ne(a, b interface{}) Predicate { return cmp(a, b, "<>") }

func Gt(a, b interface{}) Predicate { return cmp(a, b, ">") }

func Ge(a, b interface{}) Predicate { return cmp(a, b, ">=") }

func Lt(a, b interface{}) Predicate { return cmp(a, b, "<") }

func Le(a, b interface{}) Predicate { return cmp(a, b, "<=") }

func In(a, b interface{}) Predicate {
	if b, ok := b.(RowValue); ok {
		return Predicatef("{} IN {}", a, b)
	}
	return Predicatef("{} IN ({})", a, b)
}

func Exists(q Query) Predicate { return Predicatef("EXISTS ({})", q) }

func NotExists(q Query) Predicate { return Predicatef("NOT EXISTS ({})", q) }
