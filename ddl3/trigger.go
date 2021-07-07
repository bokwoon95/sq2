package ddl3

type Trigger struct {
	TableSchema string
	TableName   string
	TriggerName string
	SQL         string
}

func getTriggerInfo(sql string) (tableSchema, tableName, triggerName string, err error) {
	return "", "", "", nil
}

// catalog.LoadTriggerFS(triggerName string, fsys fs.FS, filename string)
// catalog.LoadTrigger(ddl.Trigger{TriggerName: "", Contents: ""})
