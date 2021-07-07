package ddl3

type DropSchemaCommand struct {
	Valid       bool     // 1 byte
	DoIfExists  bool     // 1 byte
	SchemaNames []string // 24 bytes
	DoCascade   bool     // 1 byte
} // 27 bytes

type DropTableCommand struct {
	Valid        bool
	DoIfExists   bool
	TableSchemas []string
	TableNames   []string
	DoCascade    bool
}

type DropColumnCommand struct {
	Valid       bool
	DoIfExists  bool
	TableSchema string
	TableName   string
	ColumnName  string
	DoCascade   bool
}

type DropConstraintCommand struct {
	Valid            bool
	DoIfExists       bool
	TableSchema      string
	TableName        string
	ConstraintSchema string
	ConstraintName   string
	DoCascade        bool
}

type DropIndexCommand struct {
	Valid          bool
	DoConcurrently bool
	DoIfExists     bool
	TableSchemas   []string
	TableNames     []string
	IndexSchemas   []string
	IndexNames     []string
	DoCascade      bool
}

type DropViewCommand struct {
	Valid       bool
	DoIfExists  bool
	ViewSchemas []string
	ViewNames   []string
	DoCascade   bool
}

type DropFunctionCommand struct {
	Valid          bool
	DoIfExists     bool
	FunctionSchema string
	FunctionName   string
	ArgModes       []string
	ArgNames       []string
	ArgTypes       []string
	DoCascade      bool
}

type DropTriggerCommand struct {
	Valid       bool
	DoIfExists  bool
	TableSchema string
	TableName   string
	TriggerName string
	DoCascade   bool
}
