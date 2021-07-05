package ddl3

type CreateSchemaCommand struct {
	Dialect       string
	DoIfNotExists bool
	SchemaName    string
	Authorization string
}
