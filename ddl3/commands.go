package ddl3

import (
	"io"

	"github.com/bokwoon95/sq"
)

type Command interface {
	ToSQL(dialect string) (string, error)
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
