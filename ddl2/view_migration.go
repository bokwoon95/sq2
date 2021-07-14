package ddl2

import (
	"bytes"
	"fmt"

	"github.com/bokwoon95/sq"
)

type ViewMigration struct {
	ViewSchema        string
	ViewName          string
	CreateCommand     *CreateViewCommand
	DropCommand       *DropViewCommand
	RenameCommand     *RenameViewCommand
	ReplaceCommand    *RenameViewCommand
	TriggerMigrations []TriggerMigration
	// IndexMigrations []IndexMigration
}

type CreateViewCommand struct {
	CreateOrReplace   bool
	CreateIfNotExists bool
	View              View
}

var _ Command = &CreateViewCommand{}

func (cmd *CreateViewCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("CREATE ")
	if cmd.CreateOrReplace {
		if cmd.View.IsMaterialized && dialect == sq.DialectPostgres {
			return fmt.Errorf("postgres MATERIALIZED VIEWs cannot be REPLACE-d, you have to drop it and recreate")
		}
		buf.WriteString("OR REPLACE ")
	}
	if cmd.View.IsMaterialized {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support CREATE MATERIALIZED VIEW", dialect)
		}
		buf.WriteString("MATERIALIZED ")
	}
	buf.WriteString("VIEW ")
	if cmd.CreateIfNotExists {
		if dialect != sq.DialectSQLite {
			return fmt.Errorf("%s does not support CREATE VIEW IF NOT EXISTS", dialect)
		}
		buf.WriteString("IF NOT EXISTS ")
	}
	if cmd.View.ViewSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.View.ViewSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.View.ViewName) + " AS " + cmd.View.Query + ";")
	return nil
}

type DropViewCommand struct {
	DropIfExists bool
	ViewSchemas  []string
	ViewNames    []string
	DropCascade  bool
}

type RenameViewCommand struct {
	AlterViewIfExists bool
	ViewSchema        string
	ViewName          string
	RenameToName      string
}
