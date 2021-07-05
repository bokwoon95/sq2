package ddl3

type AddColumnCommand struct {
	Dialect         string
	DoIfTableExists bool
	TableSchema     string
	TableName       string
	DoIfNotExists   bool
	Column          Column
	Constraints     []Constraint // instead of a fat slice, we just need a specialized set of constraints for SQLite
	// For SQLite, if INTEGER and PRIMARY KEY and AUOTINCREMENT, we place them
	// tgt. If other constraints, we treat it as special case. No need for an
	// entire constraints slice, which will likely take up a lot of space.
}
