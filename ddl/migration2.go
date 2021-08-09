package ddl

type Migration2 struct {
	Dialect                  string
	CreateSchemasCommands    []CreateSchemaCommand
	CreateExtensionCommands  []CreateExtensionCommand
	CreateFunctionCommands   []CreateFunctionCommand
	CreateTableCommands      []CreateTableCommand
	AlterTableCommands       []AlterTableCommand // add & alter columns | add & alter constraints | add indexes
	CreateViewCommands       []CreateViewCommand
	CreateIndexeCommands     []CreateIndexCommand
	CreateTriggerCommands    []CreateTriggerCommand
	CreateForeignKeyCommands []AlterTableCommand
	DropViewCommands         []DropViewCommand
	DropTableCommands        []DropTableCommand
	DropTriggerCommands      []DropTriggerCommand
	DropIndexeCommands       []DropIndexCommand
	DropConstraintCommands   []DropConstraintCommand
	DropColumnCommands       []DropColumnCommand
	DropFunctionCommands     []DropFunctionCommand
	DropExtensionCommands    []DropExtensionCommand
}
