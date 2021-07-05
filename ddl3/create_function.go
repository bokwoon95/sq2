package ddl3

type CreateFunctionCommand struct {
	Dialect     string
	DoOrReplace bool
	Function    Object
}
