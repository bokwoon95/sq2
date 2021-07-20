package ddl

import (
	"context"
	"fmt"
	"io"

	"github.com/bokwoon95/sq"
)

type MigrationCommands struct {
	Dialect            string
	SchemaCommands     []Command
	FunctionCommands   []Command
	TableCommands      []Command
	ViewCommands       []Command
	TriggerCommands    []Command
	ForeignKeyCommands []Command
}

func (m *MigrationCommands) WriteSQL(w io.Writer) error {
	var written bool
	for _, cmds := range [][]Command{
		m.SchemaCommands,
		m.FunctionCommands,
		m.TableCommands,
		m.ViewCommands,
		m.TriggerCommands,
		m.ForeignKeyCommands,
	} {
		for _, cmd := range cmds {
			query, args, _, err := sq.ToSQL(m.Dialect, cmd)
			if err != nil {
				return fmt.Errorf("command: %s: %w", query, err)
			}
			if len(args) > 0 {
				query, err = sq.Sprintf(m.Dialect, query, args)
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

func (m *MigrationCommands) Exec(db sq.DB) error {
	return m.ExecContext(context.Background(), db)
}

func (m *MigrationCommands) ExecContext(ctx context.Context, db sq.DB) error {
	for _, cmds := range [][]Command{
		m.SchemaCommands,
		m.FunctionCommands,
		m.TableCommands,
		m.ViewCommands,
		m.TriggerCommands,
		m.ForeignKeyCommands,
	} {
		for _, cmd := range cmds {
			query, args, _, err := sq.ToSQL(m.Dialect, cmd)
			if err != nil {
				return fmt.Errorf("command: %s: %w", query, err)
			}
			_, err = db.ExecContext(ctx, query, args...)
			if err != nil {
				return fmt.Errorf("command: %s: %w", query, err)
			}
		}
	}
	return nil
}
