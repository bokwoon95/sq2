package ddl

import "fmt"

type DDLEnum interface {
	fmt.Stringer
	EnumerateStringer() []fmt.Stringer
}

type Enum struct {
	EnumName   string
	EnumLabels []string
}
