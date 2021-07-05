package ddl3

import "github.com/bokwoon95/sq"

type ViewSource interface {
	sq.SchemaTable
	// TODO: extra argument that can be used to register certain view
	// porperties like MATERIALIZED or RECURSIVE.
	View(dialect string) sq.Query
}
