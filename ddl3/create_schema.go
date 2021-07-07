package ddl3

type CreateSchemaCommand struct {
	Valid         bool
	Dialect       string
	DoIfNotExists bool
	SchemaName    string
	Authorization string
}
