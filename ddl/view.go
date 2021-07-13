package ddl

import (
	"fmt"
)

type View struct {
	ViewSchema     string    `json:",omitempty"`
	ViewName       string    `json:",omitempty"`
	IsMaterialized bool      `json:",omitempty"`
	Columns        []string  `json:",omitempty"`
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
