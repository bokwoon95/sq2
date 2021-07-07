package ddl3

type CreateSchemaCommand struct {
	Valid         bool   // 1 byte
	DoIfNotExists bool   // 1 byte
	SchemaName    string // 16 bytes
	Authorization string // 16 bytes
} // 34 bytes
