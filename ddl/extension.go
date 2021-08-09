package ddl

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

func (cmd CreateExtensionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
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
	buf.WriteString(sq.QuoteIdentifier(dialect, extname))
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

type DropExtensionCommand struct {
	DropIfExists bool
	Extensions   []string
	DropCascade  bool
}

func (cmd DropExtensionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect != sq.DialectPostgres {
		return fmt.Errorf("%w dialect=%s feature=extensions", ErrUnsupportedFeature, dialect)
	}
	buf.WriteString("DROP EXTENSION ")
	if cmd.DropIfExists {
		buf.WriteString("IF EXISTS ")
	}
	for i, extension := range cmd.Extensions {
		if i > 0 {
			buf.WriteString(", ")
		}
		if n := strings.IndexByte(extension, '@'); n >= 0 {
			extension = extension[:n]
		}
		buf.WriteString(sq.QuoteIdentifier(dialect, extension))
	}
	if cmd.DropCascade {
		buf.WriteString(" CASCADE")
	}
	return nil
}
