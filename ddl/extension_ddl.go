package ddl

import (
	"bytes"
	"fmt"

	"github.com/bokwoon95/sq"
)

type CreateExtensionCommand struct {
	CreateIfNotExists bool
	Extension         [2]string
	WithSchema        string
	CreateCascade     bool
}

func (cmd *CreateExtensionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect != sq.DialectPostgres {
		return fmt.Errorf("%s does not support (creating) extensions", dialect)
	}
	buf.WriteString("CREATE EXTENSION ")
	if cmd.CreateIfNotExists {
		buf.WriteString("IF NOT EXISTS ")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Extension[0]))
	if cmd.WithSchema != "" || cmd.Extension[1] != "" || cmd.CreateCascade {
		buf.WriteString(" WITH")
		if cmd.WithSchema != "" {
			buf.WriteString(" SCHEMA" + sq.QuoteIdentifier(dialect, cmd.WithSchema))
		}
		if cmd.Extension[1] != "" {
			buf.WriteString(` VERSION '` + sq.EscapeQuote(cmd.Extension[1], '\'') + `'`)
		}
		if cmd.CreateCascade {
			buf.WriteString(" CASCADE")
		}
	}
	return nil
}

type DropExtensionCommand struct {
	DropIfExists   bool
	ExtensionNames []string
	DropCascade    bool
}

func (cmd *DropExtensionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect != sq.DialectPostgres {
		return fmt.Errorf("%s does not support (dropping) extensions", dialect)
	}
	buf.WriteString("DROP EXTENSION ")
	if cmd.DropIfExists {
		buf.WriteString("IF EXISTS ")
	}
	for i, extensionName := range cmd.ExtensionNames {
		if i == 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(sq.QuoteIdentifier(dialect, extensionName))
	}
	if cmd.DropCascade {
		buf.WriteString(" CASCADE")
	}
	return nil
}
