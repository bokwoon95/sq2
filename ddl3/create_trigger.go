package ddl3

type CreateTriggerCommand struct {
	Valid   bool
	Dialect string
	Trigger Trigger
}
