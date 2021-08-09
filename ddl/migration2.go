package ddl

type Migration2 struct {
	Dialect           string
	CreateSchemas     []CreateSchemaCommand
	CreateExtensions  []CreateExtensionCommand
	CreateFunctions   []CreateFunctionCommand
	CreateTables      []CreateTableCommand
	AlterTables       []AlterTableCommand
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
