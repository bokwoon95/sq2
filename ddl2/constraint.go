package ddl2

type Constraint struct {
	ConstraintSchema    string
	ConstraintName      string
	ConstraintType      string
	TableSchema         string
	TableName           string
	Columns             []string
	ReferencesSchema    string
	ReferencesTable     string
	ReferencesColumns   []string
	OnUpdate            string
	OnDelete            string
	MatchOption         string
	CheckExpr           string
	IsDeferrable        bool
	IsInitiallyDeferred bool
}
