package ddl3

type CreateFunctionCommand struct {
	Valid    bool
	Dialect  string
	Function Function
}
