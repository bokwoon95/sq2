package ddl2

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

type CreateExtensionCommand struct {
	CreateIfNotExists bool
	Extension         string
	CreateCascade     bool
}

func (cmd *CreateExtensionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect != sq.DialectPostgres {
		return fmt.Errorf("%w dialect=%s feature=extensions", ErrUnsupportedFeature, dialect)
	}
	buf.WriteString("CREATE EXTENSION ")
	if cmd.CreateIfNotExists {
		buf.WriteString("IF NOT EXISTS ")
	}
	extname, extversion := cmd.Extension, ""
	if i := strings.IndexByte(cmd.Extension, '@'); i >= 0 {
		extname, extversion = cmd.Extension[:i], cmd.Extension[i+1:]
	}
	buf.WriteString(extname)
	if extversion != "" || cmd.CreateCascade {
		buf.WriteString(" WITH")
		if extversion != "" {
			buf.WriteString(" VERSION '" + sq.EscapeQuote(extversion, '\'') + "'")
		}
		if cmd.CreateCascade {
			buf.WriteString(" CASCADE")
		}
	}
	return nil
}
