package ddl3

import (
	"io"

	"github.com/bokwoon95/sq"
)

type Command interface {
	sq.SQLAppender
	GetType() string
}

type Commands []Command

func (cmds Commands) WriteOut(w io.Writer) error {
	return nil
}

func (cmds Commands) ExecDB(db sq.Queryer) error {
	return nil
}

func AutoMigrate(db sq.Queryer, tables []sq.Table, items ...interface{}) error {
	return nil
}