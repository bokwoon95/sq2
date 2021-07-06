package ddl3

import "github.com/bokwoon95/sq"

type View struct {
	ViewSchema string
	ViewName   string
	Contents   string
}

type IView interface {
	sq.SchemaTable
	DDL(dialect string, v *V) sq.Query
}

type V struct {
	doOrReplace    bool
	isMaterialized bool
	isRecursive    bool
}

// TODO: maybe these should take in a bool instead?

func (v *V) Materialized() { v.isMaterialized = true }

func (v *V) Recursive() { v.isRecursive = true }

func (v *V) CreateOrReplace() { v.doOrReplace = true }
