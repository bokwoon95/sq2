package sq

import (
	"bytes"
	"fmt"
)

type JoinType string

const (
	JoinTypeInner JoinType = "JOIN"
	JoinTypeLeft  JoinType = "LEFT JOIN"
	JoinTypeRight JoinType = "RIGHT JOIN"
	JoinTypeFull  JoinType = "FULL JOIN"
	JoinTypeCross JoinType = "CROSS JOIN"
)

type JoinTable struct {
	JoinType     JoinType
	Table        Table
	OnPredicates VariadicPredicate
}

var _ SQLAppender = JoinTable{}

func Join(table Table, predicates ...Predicate) JoinTable {
	return JoinTable{
		JoinType:     JoinTypeInner,
		Table:        table,
		OnPredicates: VariadicPredicate{Predicates: predicates},
	}
}

func LeftJoin(table Table, predicates ...Predicate) JoinTable {
	return JoinTable{
		JoinType:     JoinTypeLeft,
		Table:        table,
		OnPredicates: VariadicPredicate{Predicates: predicates},
	}
}

func RightJoin(table Table, predicates ...Predicate) JoinTable {
	return JoinTable{
		JoinType:     JoinTypeRight,
		Table:        table,
		OnPredicates: VariadicPredicate{Predicates: predicates},
	}
}

func FullJoin(table Table, predicates ...Predicate) JoinTable {
	return JoinTable{
		JoinType:     JoinTypeFull,
		Table:        table,
		OnPredicates: VariadicPredicate{Predicates: predicates},
	}
}

func CrossJoin(table Table) JoinTable {
	return JoinTable{
		JoinType: JoinTypeCross,
		Table:    table,
	}
}

func CustomJoin(joinType JoinType, table Table, predicates ...Predicate) JoinTable {
	return JoinTable{
		JoinType:     joinType,
		Table:        table,
		OnPredicates: VariadicPredicate{Predicates: predicates},
	}
}

func (join JoinTable) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if join.JoinType == "" {
		join.JoinType = JoinTypeInner
	}
	if len(join.OnPredicates.Predicates) == 0 &&
		(join.JoinType == JoinTypeInner ||
			join.JoinType == JoinTypeRight ||
			join.JoinType == JoinTypeFull) {
		return fmt.Errorf("%s requires at least one predicate specified", join.JoinType)
	}
	if dialect == DialectSQLite && (join.JoinType == JoinTypeRight || join.JoinType == JoinTypeFull) {
		return fmt.Errorf("sqlite does not support %s", join.JoinType)
	}
	buf.WriteString(string(join.JoinType) + " ")
	if join.Table == nil {
		return fmt.Errorf("joining on a nil table")
	}
	err := join.Table.AppendSQL(dialect, buf, args, params)
	if err != nil {
		return err
	}
	if tableAlias := join.Table.GetAlias(); tableAlias != "" {
		buf.WriteString(" AS ")
		buf.WriteString(QuoteIdentifier(dialect, tableAlias))
	}
	if len(join.OnPredicates.Predicates) > 0 {
		buf.WriteString(" ON ")
		join.OnPredicates.Toplevel = true
		err = join.OnPredicates.AppendSQLExclude(dialect, buf, args, params, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

type JoinTables []JoinTable

var _ SQLAppender = JoinTables{}

func (joins JoinTables) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	var err error
	for i, join := range joins {
		if i > 0 {
			buf.WriteString(" ")
		}
		err = join.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
		}
	}
	return nil
}
