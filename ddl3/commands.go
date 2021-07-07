package ddl3

import (
	"io"

	"github.com/bokwoon95/sq"
)

type Command interface {
	sq.SQLAppender
	// TODO: there's no point in making it an SQL appender. Every command
	// already embeds the dialect. The only method you need on a command should
	// be `ToSQL() (string, error)`.
	// TODO: what if... the command didn't embed the dialect then? Would it make AppendSQL viable?
	GetType() string
}

type Commands []Command

func (cmds Commands) WriteOut(w io.Writer) error {
	return nil
}

func (cmds Commands) ExecDB(db sq.Queryer) error {
	return nil
}

func AutoMigrate(db sq.Queryer, opts ...CatalogOption) error {
	return nil
}
