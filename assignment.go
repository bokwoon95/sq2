package sq

import (
	"bytes"
	"fmt"
)

type Assignment struct {
	LHS interface{}
	RHS interface{}
}

var _ SQLExcludeAppender = Assignment{}

// TODO: rename this to Set, for parity with (XXXField).Set()
func Assign(LHS, RHS interface{}) Assignment {
	return Assignment{LHS: LHS, RHS: RHS}
}

func (a Assignment) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	err := BufferPrintValue(dialect, buf, args, params, nil, excludedTableQualifiers, a.LHS, "")
	if err != nil {
		return err
	}
	buf.WriteString(" = ")
	switch a.RHS.(type) {
	case Query:
		buf.WriteString("(")
		err = BufferPrintValue(dialect, buf, args, params, env, excludedTableQualifiers, a.RHS, "")
		if err != nil {
			return err
		}
		buf.WriteString(")")
	default:
		err = BufferPrintValue(dialect, buf, args, params, env, excludedTableQualifiers, a.RHS, "")
		if err != nil {
			return err
		}
	}
	return nil
}

type Assignments []Assignment

var _ SQLExcludeAppender = Assignments{}

func (as Assignments) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	var err error
	for i, a := range as {
		if i > 0 {
			buf.WriteString(", ")
		}
		err = a.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
		if err != nil {
			return fmt.Errorf("assignment #%d: %w", i+1, err)
		}
	}
	return nil
}

// TODO: SetExcluded? sounds way better. SetAlias is bad, but AssignAlias sounds just as confusing.
func AssignExcluded(field Field) Assignment {
	name := field.GetName()
	return Assign(Literal(name), Literal("EXCLUDED."+name))
}

// TODO: should the argument order be swapped instead? alias then field?
func AssignAlias(field Field, alias string) Assignment {
	name := field.GetName()
	return Assign(Literal(name), Literal(alias+"."+name))
}

func AssignSelf(field Field) Assignment {
	name := field.GetName()
	return Assign(Literal(name), Literal(name))
}

func AssignValues(field Field) Assignment {
	name := field.GetName()
	return Assign(Literal(name), Literal("VALUES("+name+")"))
}
