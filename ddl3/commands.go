package ddl3

import (
	"io"

	"github.com/bokwoon95/sq"
)

type Command interface {
	sq.SQLAppender
}

type Commands []Command

func (cmds Commands) WriteOut(w io.Writer) error {
	return nil
}

func (cmds Commands) ExecDB(db sq.DB) error {
	return nil
}

func AutoMigrate(db sq.DB, opts ...CatalogOption) error {
	return nil
}
