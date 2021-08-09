package ddl

type Migration2 struct {
	Dialect           string
	CreateSchemas     []CreateSchemaCommand
	CreateExtensions  []CreateExtensionCommand
	CreateFunctions   []CreateFunctionCommand
	CreateTables      []CreateTableCommand
	AlterTables       []AlterTableCommand // add & alter columns | add & alter constraints | add indexes
	CreateViews       []CreateViewCommand
	CreateIndexes     []CreateIndexCommand
	CreateTriggers    []CreateTriggerCommand
	CreateForeignKeys []AlterTableCommand
	DropViews         []DropViewCommand
	DropTables        []DropTableCommand
	DropTriggers      []DropTriggerCommand
	DropIndexes       []DropIndexCommand
	DropConstraints   []DropConstraintCommand
	DropColumns       []DropColumnCommand
	DropFunctions     []DropFunctionCommand
	DropExtensions    []DropExtensionCommand
}
