package ddl3

import "github.com/bokwoon95/sq"

type View struct {
	ViewSchema     string
	ViewName       string
	IsMaterialized bool
	IsRecursive    bool
	Contents       string
}

type IView interface {
	sq.SchemaTable
	DDL(dialect string, v *V) sq.Query
}

type V struct {
	isMaterialized bool
	isRecursive    bool
}

func (v *V) Materialized() { v.isMaterialized = true }

func (v *V) Recursive() { v.isRecursive = true }
