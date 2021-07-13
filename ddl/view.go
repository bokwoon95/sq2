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
	indexesCache   map[string]int
	triggersCache  map[[3]string]int
}

func (view *View) CachedIndexPosition(indexName string) (indexPosition int) {
	if indexName == "" {
		return -1
	}
	indexPosition, ok := view.indexesCache[indexName]
	if !ok {
		return -1
	}
	if indexPosition < 0 || indexPosition >= len(view.Indexes) || view.Indexes[indexPosition].IndexName != indexName {
		delete(view.indexesCache, indexName)
		return -1
	}
	return indexPosition
}

func (view *View) AppendIndex(index Index) (indexPosition int) {
	view.Indexes = append(view.Indexes, index)
	if view.indexesCache == nil {
		view.indexesCache = make(map[string]int)
	}
	indexPosition = len(view.Indexes) - 1
	view.indexesCache[index.IndexName] = indexPosition
	return indexPosition
}

func (view *View) RefreshIndexesCache() {
	for i, index := range view.Indexes {
		if view.indexesCache == nil {
			view.indexesCache = make(map[string]int)
		}
		view.indexesCache[index.IndexName] = i
	}
}

func (view *View) CachedTriggerPosition(tableSchema, tableName, triggerName string) (triggerPosition int) {
	if triggerName == "" {
		return -1
	}
	key := [3]string{tableSchema, tableName, triggerName}
	triggerPosition, ok := view.triggersCache[key]
	if !ok {
		return -1
	}
	if triggerPosition < 0 || triggerPosition >= len(view.Triggers) {
		delete(view.triggersCache, key)
		return -1
	}
	trigger := view.Triggers[triggerPosition]
	if trigger.TableSchema != tableSchema || trigger.TableName != tableName || trigger.TriggerName != triggerName {
		delete(view.triggersCache, key)
		return -1
	}
	return triggerPosition
}

func (view *View) AppendTrigger(trigger Trigger) (triggerPosition int) {
	view.Triggers = append(view.Triggers, trigger)
	if view.triggersCache == nil {
		view.triggersCache = make(map[[3]string]int)
	}
	key := [3]string{trigger.TableSchema, trigger.TableName, trigger.TriggerName}
	triggerPosition = len(view.Triggers) - 1
	view.triggersCache[key] = triggerPosition
	return triggerPosition
}

func (view *View) RefreshTriggersCache() {
	for i, trigger := range view.Triggers {
		if view.triggersCache == nil {
			view.triggersCache = make(map[[3]string]int)
		}
		key := [3]string{trigger.TableSchema, trigger.TableName, trigger.TriggerName}
		view.triggersCache[key] = i
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
			IndexType:   "BTREE",
			Columns:     columns,
			Exprs:       exprs,
		})
	}
	return indexPosition, nil
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
	CreateOrReplace   bool
	CreateIfNotExists bool
	View              View
}

func (cmd *CreateViewCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("CREATE ")
	if cmd.CreateOrReplace {
		if cmd.View.IsMaterialized && dialect == sq.DialectPostgres {
			return fmt.Errorf("postgres MATERIALIZED VIEWs cannot be REPLACE-d, you have to drop it and recreate")
		}
		buf.WriteString("OR REPLACE ")
	}
	if cmd.View.IsMaterialized {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support CREATE MATERIALIZED VIEW", dialect)
		}
		buf.WriteString("MATERIALIZED ")
	}
	buf.WriteString("VIEW ")
	if cmd.CreateIfNotExists {
		if dialect != sq.DialectSQLite {
			return fmt.Errorf("%s does not support CREATE VIEW IF NOT EXISTS", dialect)
		}
		buf.WriteString("IF NOT EXISTS ")
	}
	if cmd.View.ViewSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.View.ViewSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.View.ViewName) + " AS " + cmd.View.Query + ";")
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
