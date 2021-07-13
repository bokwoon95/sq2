package ddl

import (
	"bytes"
	"fmt"

	"github.com/bokwoon95/sq"
)

type View struct {
	ViewSchema     string    `json:",omitempty"`
	ViewName       string    `json:",omitempty"`
	IsMaterialized bool      `json:",omitempty"`
	Columns        []string  `json:",omitempty"` // can the field names of a view be fetched using sql?
	Indexes        []Index   `json:",omitempty"`
	Triggers       []Trigger `json:",omitempty"`
	Query          string    `json:",omitempty"`
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
	view.Query = buf.String()
	if len(args) > 0 {
		view.Query, err = sq.Sprintf(dialect, view.Query, args)
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
		view.Columns = append(view.Columns, fieldName)
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

type ViewDiff struct {
	ViewSchema     string
	ViewName       string
	CreateCommand  *CreateViewCommand
	DropCommand    *DropViewCommand
	RenameCommand  *RenameViewCommand
	ReplaceCommand *RenameViewCommand
	TriggerDiffs   []TriggerDiff
}

type CreateViewCommand struct {
	View View
}

func (cmd *CreateViewCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString(cmd.View.Query)
	return nil
}

type DropViewCommand struct {
	DropIfExists bool
	ViewSchemas  []string
	ViewNames    []string
	DropCascade  bool
}

type RenameViewCommand struct {
	AlterViewIfExists bool
	ViewSchema        string
	ViewName          string
	RenameToName      string
}
