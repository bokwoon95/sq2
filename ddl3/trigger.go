package ddl3

import "io"

type TriggerSource interface {
	GetSource() io.Reader
}

// catalog.LoadTriggerFS(triggerName string, fsys fs.FS, filename string)
// catalog.LoadTrigger(ddl.Trigger{TriggerName: "", Contents: ""})
