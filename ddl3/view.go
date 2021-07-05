package ddl3

import "github.com/bokwoon95/sq"

type View struct {
}

// all IViews can be converted into Views. An IView is simply a struct
// container that contains the view definition as a method. The struct itself
// can be used as a table in sq queries.
// the vw *ddl.Vw can be used to register additional attributes on the view,
// like IsMaterialized or IsRecursive.
type IView interface {
	sq.SchemaTable
	// TODO: extra argument that can be used to register certain view
	// porperties like MATERIALIZED or RECURSIVE.
	View(dialect string) sq.Query
	// TODO: what if we did DDL(dialect string, v *ddl.V) sq.Query instead?
	// That way a struct can only either be a table or view. For
	// TableValuedFunctions it would be DDL(dialect string, fn *ddl.Fn)
}
