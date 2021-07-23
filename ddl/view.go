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
	Indexes        []*Index   `json:",omitempty"`
	Triggers       []*Trigger `json:",omitempty"`
	SQL            string    `json:",omitempty"`
	Ignore         bool      `json:",omitempty"`
	indexCache     map[string]int
	triggerCache   map[[3]string]int
}

func (view *View) CachedIndexPosition(indexName string) (indexPosition int) {
	if indexName == "" {
		return -1
	}
	indexPosition, ok := view.indexCache[indexName]
	if !ok {
		return -1
	}
	if indexPosition < 0 || indexPosition >= len(view.Indexes) || view.Indexes[indexPosition].IndexName != indexName {
		delete(view.indexCache, indexName)
		return -1
	}
	return indexPosition
}

func (view *View) AppendIndex(index *Index) (indexPosition int) {
	view.Indexes = append(view.Indexes, index)
	if view.indexCache == nil {
		view.indexCache = make(map[string]int)
	}
	indexPosition = len(view.Indexes) - 1
	view.indexCache[index.IndexName] = indexPosition
	return indexPosition
}

func (view *View) RefreshIndexCache() {
	if view.indexCache == nil && len(view.Indexes) > 0 {
		view.indexCache = make(map[string]int)
	}
	for i, index := range view.Indexes {
		view.indexCache[index.IndexName] = i
	}
}

func (view *View) CachedTriggerPosition(tableSchema, tableName, triggerName string) (triggerPosition int) {
	if triggerName == "" {
		return -1
	}
	key := [3]string{tableSchema, tableName, triggerName}
	triggerPosition, ok := view.triggerCache[key]
	if !ok {
		return -1
	}
	if triggerPosition < 0 || triggerPosition >= len(view.Triggers) {
		delete(view.triggerCache, key)
		return -1
	}
	trigger := view.Triggers[triggerPosition]
	if trigger.TableSchema != tableSchema || trigger.TableName != tableName || trigger.TriggerName != triggerName {
		delete(view.triggerCache, key)
		return -1
	}
	return triggerPosition
}

func (view *View) AppendTrigger(trigger *Trigger) (triggerPosition int) {
	view.Triggers = append(view.Triggers, trigger)
	if view.triggerCache == nil {
		view.triggerCache = make(map[[3]string]int)
	}
	key := [3]string{trigger.TableSchema, trigger.TableName, trigger.TriggerName}
	triggerPosition = len(view.Triggers) - 1
	view.triggerCache[key] = triggerPosition
	return triggerPosition
}

func (view *View) RefreshTriggerCache() {
	if view.triggerCache == nil && len(view.Triggers) > 0 {
		view.triggerCache = make(map[[3]string]int)
	}
	for i, trigger := range view.Triggers {
		key := [3]string{trigger.TableSchema, trigger.TableName, trigger.TriggerName}
		view.triggerCache[key] = i
	}
}

func (view *View) createOrUpdateIndex(indexName string, columns []string, exprs []string) (indexPosition int, err error) {
	if indexName == "" {
		return -1, fmt.Errorf("indexName cannot be empty")
	}
	if indexPosition = view.CachedIndexPosition(indexName); indexPosition >= 0 {
		index := view.Indexes[indexPosition]
		index.TableSchema = view.ViewSchema
		index.TableName = view.ViewName
		index.Columns = columns
		index.Exprs = exprs
		view.Indexes[indexPosition] = index
	} else {
		indexPosition = view.AppendIndex(&Index{
			TableSchema: view.ViewSchema,
			TableName:   view.ViewName,
			IndexName:   indexName,
			Columns:     columns,
			Exprs:       exprs,
		})
	}
	return indexPosition, nil
}

type CreateViewCommand struct {
	CreateOrReplace   bool
	CreateIfNotExists bool
	View              View
}

func (cmd CreateViewCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("CREATE ")
	if cmd.CreateOrReplace {
		if dialect != sq.DialectPostgres && dialect != sq.DialectMySQL {
			return fmt.Errorf("%s does not support CREATE OR REPLACE VIEW", dialect)
		}
		if dialect == sq.DialectPostgres && cmd.View.IsMaterialized {
			return fmt.Errorf("postgres does not allow CREATE OR REPLACE VIEW for Materialized Views")
		}
		buf.WriteString("OR REPLACE ")
	}
	if cmd.View.IsMaterialized {
		buf.WriteString("MATERIALIZED ")
	}
	buf.WriteString("VIEW ")
	if cmd.CreateIfNotExists {
		if dialect != sq.DialectSQLite && dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support CREATE VIEW IF NOT EXISTS", dialect)
		}
		if dialect == sq.DialectPostgres && !cmd.View.IsMaterialized {
			return fmt.Errorf("postgres does not allow CREATE VIEW IF NOT EXISTS for Non-Materialized Views")
		}
		buf.WriteString("IF NOT EXISTS ")
	}
	if cmd.View.ViewSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.View.ViewSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.View.ViewName) + " AS " + cmd.View.SQL + ";")
	return nil
}

type DropViewCommand struct {
	DropIfExists   bool
	IsMaterialized bool
	ViewSchemas    []string
	ViewNames      []string
	DropCascade    bool
}

func (cmd DropViewCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("DROP ")
	if cmd.IsMaterialized {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support (dropping) materialized views", dialect)
		}
		buf.WriteString("MATERIALIZED ")
	}
	buf.WriteString("VIEW ")
	if cmd.DropIfExists {
		buf.WriteString("IF EXISTS ")
	}
	if len(cmd.ViewNames) > 1 && dialect == sq.DialectSQLite {
		return fmt.Errorf("sqlite does not support dropping multiple views in one command")
	}
	for i, viewSchema := range cmd.ViewSchemas {
		if i > 0 {
			buf.WriteString(", ")
		}
		if viewSchema != "" {
			buf.WriteString(sq.QuoteIdentifier(dialect, viewSchema) + ".")
		}
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.ViewNames[i]))
	}
	if cmd.DropCascade {
		if dialect == sq.DialectSQLite {
			return fmt.Errorf("sqlite does not support DROP VIEW CASCADE")
		}
		buf.WriteString(" CASCADE")
	}
	buf.WriteString(";")
	return nil
}

type RenameViewCommand struct {
	AlterViewIfExists bool
	ViewSchema        string
	ViewName          string
	RenameToName      string
}

func (cmd RenameViewCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect == sq.DialectSQLite {
		return fmt.Errorf("sqlite does not support renaming views")
	}
	buf.WriteString("ALTER VIEW ")
	if cmd.AlterViewIfExists {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support ALTER VIEW IF EXISTS", dialect)
		}
		buf.WriteString("IF EXISTS ")
	}
	if cmd.ViewSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.ViewSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.ViewName) + " RENAME TO " + sq.QuoteIdentifier(dialect, cmd.RenameToName) + ";")
	return nil
}
