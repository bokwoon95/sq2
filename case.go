package sq

import (
	"bytes"
	"fmt"
)

type PredicateCase struct {
	condition Predicate
	result    interface{}
}

type PredicateCases struct {
	alias    string
	cases    []PredicateCase
	fallback interface{}
}

var _ Field = PredicateCases{}

func (f PredicateCases) GetAlias() string { return f.alias }

func (f PredicateCases) GetName() string { return "" }

func (f PredicateCases) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	buf.WriteString("CASE")
	if len(f.cases) == 0 {
		return fmt.Errorf("CASE: no predicate cases provided")
	}
	var err error
	for i, Case := range f.cases {
		buf.WriteString(" WHEN ")
		err = BufferPrintValue(dialect, buf, args, params, env, excludedTableQualifiers, Case.condition, "")
		if err != nil {
			return fmt.Errorf("CASE #%d WHEN: %w", i+1, err)
		}
		buf.WriteString(" THEN ")
		err = BufferPrintValue(dialect, buf, args, params, env, excludedTableQualifiers, Case.result, "")
		if err != nil {
			return fmt.Errorf("CASE #%d THEN: %w", i+1, err)
		}
	}
	if f.fallback != nil {
		buf.WriteString(" ELSE ")
		err = BufferPrintValue(dialect, buf, args, params, env, excludedTableQualifiers, f.fallback, "")
		if err != nil {
			return fmt.Errorf("CASE ELSE: %w", err)
		}
	}
	buf.WriteString(" END")
	return nil
}

func (f PredicateCases) As(alias string) PredicateCases {
	f.alias = alias
	return f
}

func CaseWhen(predicate Predicate, result interface{}) PredicateCases {
	f := PredicateCases{}
	f.cases = append(f.cases, PredicateCase{
		condition: predicate,
		result:    result,
	})
	return f
}

func (f PredicateCases) When(predicate Predicate, result interface{}) PredicateCases {
	f.cases = append(f.cases, PredicateCase{
		condition: predicate,
		result:    result,
	})
	return f
}

func (f PredicateCases) Else(fallback interface{}) PredicateCases {
	f.fallback = fallback
	return f
}

func (f PredicateCases) GetType() string { return "" }

type SimpleCase struct {
	value  interface{}
	result interface{}
}

type SimpleCases struct {
	alias      string
	expression interface{}
	cases      []SimpleCase
	fallback   interface{}
}

var _ Field = SimpleCases{}

func (f SimpleCases) GetAlias() string { return f.alias }

func (f SimpleCases) GetName() string { return "" }

func (f SimpleCases) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	buf.WriteString("CASE ")
	if len(f.cases) == 0 {
		return fmt.Errorf("no predicate cases provided")
	}
	err := BufferPrintValue(dialect, buf, args, params, env, excludedTableQualifiers, f.expression, "")
	if err != nil {
		return fmt.Errorf("CASE: %w", err)
	}
	for i, Case := range f.cases {
		buf.WriteString(" WHEN ")
		err = BufferPrintValue(dialect, buf, args, params, env, excludedTableQualifiers, Case.value, "")
		if err != nil {
			return fmt.Errorf("CASE WHEN #%d: %w", i+1, err)
		}
		buf.WriteString(" THEN ")
		err = BufferPrintValue(dialect, buf, args, params, env, excludedTableQualifiers, Case.result, "")
		if err != nil {
			return fmt.Errorf("CASE THEN #%d: %w", i+1, err)
		}
	}
	if f.fallback != nil {
		buf.WriteString(" ELSE ")
		err = BufferPrintValue(dialect, buf, args, params, env, excludedTableQualifiers, f.fallback, "")
		if err != nil {
			return fmt.Errorf("CASE ELSE: %w", err)
		}
	}
	buf.WriteString(" END")
	return nil
}

func (f SimpleCases) As(alias string) SimpleCases {
	f.alias = alias
	return f
}

func Case(field Field) SimpleCases { return SimpleCases{expression: field} }

func (f SimpleCases) When(value interface{}, result interface{}) SimpleCases {
	f.cases = append(f.cases, SimpleCase{
		value:  value,
		result: result,
	})
	return f
}

func (f SimpleCases) Else(fallback interface{}) SimpleCases {
	f.fallback = fallback
	return f
}

func (f SimpleCases) GetType() string { return "" }
