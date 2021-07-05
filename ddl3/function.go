package ddl3

import (
	"io"

	"github.com/bokwoon95/sq"
)

type FunctionObject struct {
	sq.SchemaTable
}

type Function interface {
	GetSchema() string
	GetName() string
	GetArgs() (argModes, argNames, argtypes []string)
	GetSource() io.Reader
}
