package ddl3

import (
	"fmt"
	"io"

	"github.com/bokwoon95/sq"
)

type Command interface {
	sq.SQLAppender
}

func (cmdset CommandSet) WriteOut(w io.Writer) error {
	var written bool
	for _, cmds := range [][]Command{
		cmdset.SchemaCommands,
		cmdset.FunctionCommands,
		cmdset.TableCommands,
		cmdset.ViewCommands,
		cmdset.TableFunctionCommands,
		cmdset.ForeignKeyCommands,
	} {
		for _, cmd := range cmds {
			query, args, _, err := sq.ToSQL(cmdset.Dialect, cmd)
			if err != nil {
				return fmt.Errorf("command: %s: %w", query, err)
			}
			if len(args) > 0 {
				query, err = sq.Sprintf(cmdset.Dialect, query, args)
				if err != nil {
					return fmt.Errorf("command: %s: %w", query, err)
				}
			}
			if !written {
				written = true
			} else {
				io.WriteString(w, "\n\n")
			}
			io.WriteString(w, query)
		}
	}
	return nil
}

type CommandSet struct {
	Dialect               string
	SchemaCommands        []Command
	FunctionCommands      []Command
	TableCommands         []Command
	ViewCommands          []Command
	TableFunctionCommands []Command
	ForeignKeyCommands    []Command
}

func (cmdset CommandSet) ExecDB(db sq.DB) error {
	return nil
}

func AutoMigrate(db sq.DB, opts ...CatalogOption) error {
	return nil
}
