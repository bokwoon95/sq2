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
	Indexes        []Index   `json:",omitempty"`
	Triggers       []Trigger `json:",omitempty"`
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
	if indexPosition < 0 || indexPosition >= len(view.Indexes) {
		delete(view.indexCache, indexName)
		return -1
	}
	index := view.Indexes[indexPosition]
	if index.IndexName != indexName || index.Ignore {
		delete(view.indexCache, indexName)
		return -1
	}
	return indexPosition
}

func (view *View) AppendIndex(index Index) (indexPosition int) {
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
		if view.Ignore {
			continue
		}
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
	if trigger.TableSchema != tableSchema || trigger.TableName != tableName || trigger.TriggerName != triggerName || trigger.Ignore {
		delete(view.triggerCache, key)
		return -1
	}
	return triggerPosition
}

func (view *View) AppendTrigger(trigger Trigger) (triggerPosition int) {
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
		if trigger.Ignore {
			continue
		}
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
		indexPosition = view.AppendIndex(Index{
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
		if dialect == sq.DialectPostgres && cmd.View.IsMaterialized {
			return fmt.Errorf("postgres does not allow CREATE OR REPLACE VIEW for Materialized Views")
		}
		if dialect != sq.DialectPostgres && dialect != sq.DialectMySQL {
			return fmt.Errorf("%s does not support CREATE OR REPLACE VIEW", dialect)
		}
		buf.WriteString("OR REPLACE ")
	}
	if cmd.View.IsMaterialized {
		buf.WriteString("MATERIALIZED ")
	}
	buf.WriteString("VIEW ")
	if cmd.CreateIfNotExists {
		if dialect == sq.DialectPostgres && !cmd.View.IsMaterialized {
			return fmt.Errorf("postgres does not allow CREATE VIEW IF NOT EXISTS for Non-Materialized Views")
		}
		if dialect != sq.DialectSQLite && dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support CREATE VIEW IF NOT EXISTS", dialect)
		}
		buf.WriteString("IF NOT EXISTS ")
	}
	if cmd.View.ViewSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.View.ViewSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.View.ViewName) + " AS " + cmd.View.SQL)
	return nil
}

type DropViewCommand struct {
	DropIfExists   bool
	IsMaterialized bool
	ViewSchemas    []string
	ViewNames      []string
	DropCascade    bool
}

func (cmd *DropViewCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("DROP ")
	if cmd.IsMaterialized {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support (dropping) materialized views", dialect)
		}
		buf.WriteString("MATERIALIZED ")
	}
	buf.WriteString("VIEW ")
	if cmd.DropIfExists {
		buf.WriteString("IF EXISTS")
		if len(cmd.ViewNames) == 1 {
			buf.WriteString(" ")
		}
	}
	if len(cmd.ViewNames) > 1 && dialect == sq.DialectSQLite {
		return fmt.Errorf("sqlite does not support dropping multiple views in one command")
	}
	for i, viewSchema := range cmd.ViewSchemas {
		if i > 0 {
			buf.WriteString("\n    ,")
		} else if len(cmd.ViewNames) > 1 {
			buf.WriteString("\n    ")
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
		if len(cmd.ViewNames) > 1 {
			buf.WriteString("\n")
		} else {
			buf.WriteString(" ")
		}
		buf.WriteString("CASCADE")
	}
	return nil
}

func decomposeDropViewCommandSQLite(dropViewCmd *DropViewCommand) []Command {
	dropTableCmds := make([]DropViewCommand, 0, len(dropViewCmd.ViewNames))
	for i, viewName := range dropViewCmd.ViewNames {
		dropTableCmds = append(dropTableCmds, DropViewCommand{
			DropIfExists: dropViewCmd.DropIfExists,
			ViewSchemas:  []string{dropViewCmd.ViewSchemas[i]},
			ViewNames:    []string{viewName},
		})
	}
	cmds := make([]Command, len(dropViewCmd.ViewNames))
	for i := range dropTableCmds {
		cmds[i] = &dropTableCmds[i]
	}
	return cmds
}

func decomposeDropViewCommandSQLite2(dropViewCmd DropViewCommand) []DropViewCommand {
	dropViewCmds := make([]DropViewCommand, 0, len(dropViewCmd.ViewNames))
	for i, viewName := range dropViewCmd.ViewNames {
		dropViewCmds = append(dropViewCmds, DropViewCommand{
			DropIfExists: dropViewCmd.DropIfExists,
			ViewSchemas:  []string{dropViewCmd.ViewSchemas[i]},
			ViewNames:    []string{viewName},
		})
	}
	return dropViewCmds
}
