package ddl

import (
	"bytes"
	"fmt"

	"github.com/bokwoon95/sq"
)

type View struct {
	ViewSchema string
	ViewName   string
	FieldNames []string
	SQL        string
}

func (view *View) loadQuery(q sq.Query, v *V) error {
	buf := bufpool.Get().(*bytes.Buffer)
	args := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		args = args[:0]
		bufpool.Put(buf)
		argspool.Put(args)
	}()
	buf.WriteString("CREATE ")
	if v.CreateOrReplace {
		buf.WriteString("OR REPLACE ")
	}
	buf.WriteString("VIEW ")
	if v.CreateIfNotExists {
		buf.WriteString("IF NOT EXISTS ")
	}
	dialect := q.GetDialect()
	if view.ViewSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, view.ViewSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, view.ViewName) + " AS ")
	err := q.AppendSQL(dialect, buf, &args, make(map[string][]int))
	if err != nil {
		return err
	}
	buf.WriteString(";")
	view.SQL = buf.String()
	if len(args) > 0 {
		view.SQL, err = sq.Sprintf(dialect, view.SQL, args)
		if err != nil {
			return err
		}
	}
	fields, err := q.GetFetchableFields()
	if err != nil {
		return fmt.Errorf("fetching view fields: %w", err)
	}
	for i, field := range fields {
		fieldName := field.GetAlias()
		if fieldName == "" {
			fieldName = field.GetName()
		}
		if fieldName == "" {
			return fmt.Errorf("view query %s field #%d has no name and no alias", view.ViewName, i+1)
		}
		view.FieldNames = append(view.FieldNames, fieldName)
	}
	return nil
}

type DDLView interface {
	sq.SchemaTable
	DDL(dialect string, v *V) sq.Query
}

type V struct {
	CreateOrReplace   bool
	CreateIfNotExists bool
	IsMaterialized    bool
	IsRecursive       bool
	Triggers          []Trigger
}
