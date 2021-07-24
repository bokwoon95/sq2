package ddl

import (
	"bytes"
	"database/sql"
	"fmt"

	"github.com/bokwoon95/sq"
)

/*
hallelujah: https://github.com/golang/go/issues/45346
type Integer = interface { ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 }
type Enum[T interface{ ~string | Integer; fmt.Stringer}] interface {
	T
	Enumerate() []T
}
for _, enum := range status.Enumerate() {
	if enum == status {
		attributes["selected"] = "true"
	} else {
		attributes["selected"] = "false"
	}
}
if you have the Enumerate() method, you don't necessarily need a driver.Valuer
or sql.Scanner anymore. the library can type switch to figure out the
underlying type, and use reflect to get/set the values accordingly. Of course
scanning would still require looping over every Enumerate() value, so it's
faster to implement your own Scan() method that doesn't use looping.
type ORDERS struct {
	STATUS sq.EnumField[Status]
}
by default SchemaName is empty
*/

// sq.EnumField could be generic on a type, then you can use EqEnum[T
// Enumer](EnumField[T], T) Predicate 🤔
type Enum struct {
	EnumSchema string   `json:",omitempty"`
	EnumName   string   `json:",omitempty"`
	Values     []string `json:",omitempty"`
}

// 🤔 what if NewEnum[T fmt.Stringer](values ...T), and the EnumName was
// reflected from the type? MPAARating -> mpaa_rating
func NewEnum(enumSchema, enumName string, enumValues ...fmt.Stringer) Enum {
	enum := Enum{
		EnumSchema: enumSchema,
		EnumName:   enumName,
		Values:     make([]string, 0, len(enumValues)),
	}
	for _, value := range enumValues {
		enum.Values = append(enum.Values, value.String())
	}
	return enum
}

type CreateEnumCommand struct {
	Enum Enum
}

func (cmd *CreateEnumCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect != sq.DialectPostgres {
		return fmt.Errorf("%s does not support creating enums", dialect)
	}
	buf.WriteString("CREATE TYPE ")
	if cmd.Enum.EnumSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Enum.EnumSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Enum.EnumName) + " AS ENUM (")
	for i, value := range cmd.Enum.Values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(`'` + sq.EscapeQuote(value, '\'') + `'`)
	}
	buf.WriteString(")")
	return nil
}

type AddEnumValueCommand struct {
	AddIfNotExists  bool
	EnumSchema      string
	EnumName        string
	NewValue        string
	NeighbourValue  sql.NullString
	BeforeNeighbour bool
}

func (cmd *AddEnumValueCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect != sq.DialectPostgres {
		return fmt.Errorf("%s does not support creating enums", dialect)
	}
	return nil
}
