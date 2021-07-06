package ddl3

type Trigger struct {
	TableSchema string
	TableName   string
	TriggerName string
	SQL         string
}

// catalog.LoadTriggerFS(triggerName string, fsys fs.FS, filename string)
// catalog.LoadTrigger(ddl.Trigger{TriggerName: "", Contents: ""})
