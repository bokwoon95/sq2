package ddl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/bokwoon95/sq"
)

var (
	bufpool  = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
	argspool = sync.Pool{New: func() interface{} { return make([]interface{}, 0) }}
)

const (
	PRIMARY_KEY = "PRIMARY KEY"
	FOREIGN_KEY = "FOREIGN KEY"
	UNIQUE      = "UNIQUE"
	CHECK       = "CHECK"
	INDEX       = "INDEX"
	EXCLUDE     = "EXCLUDE"

	BY_DEFAULT_AS_IDENTITY = "BY DEFAULT AS IDENTITY"
	ALWAYS_AS_IDENTITY     = "ALWAYS AS IDENTITY"

	RESTRICT    = "RESTRICT"
	CASCADE     = "CASCADE"
	NO_ACTION   = "NO ACTION"
	SET_NULL    = "SET NULL"
	SET_DEFAULT = "SET DEFAULT"
)

type CommandType int

const (
	CREATE_SCHEMA CommandType = iota
	RENAME_SCHEMA
	DROP_SCHEMA
	CREATE_TABLE
	RENAME_TABLE
	DROP_TABLE
	ADD_COLUMN
	ALTER_COLUMN
	RENAME_COLUMN
	DROP_COLUMN
	ADD_CONSTRAINT
	RENAME_CONSTRAINT
	DROP_CONSTRAINT
	CREATE_INDEX
	RENAME_INDEX
	DROP_INDEX
	CREATE_FUNCTION
	RENAME_FUNCTION
	DROP_FUNCTION
	CREATE_VIEW
	RENAME_VIEW
	DROP_VIEW
	CREATE_TRIGGER
	RENAME_TRIGGER
	DROP_TRIGGER
	TABLE_DML
)

type Command interface {
	sq.SQLAppender
}

type MigrationCommands struct {
	Dialect            string
	SchemaCommands     []Command
	FunctionCommands   []Command
	TableCommands      []Command
	ViewCommands       []Command
	TriggerCommands    []Command
	ForeignKeyCommands []Command
}

func (m MigrationCommands) WriteOut(w io.Writer) error {
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

func (m MigrationCommands) ExecContext(ctx context.Context, db sq.DB) error {
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

type Exclusions []struct {
	Field    sq.Field
	Operator string
}

func generateName(nameType string, tableName string, columnNames ...string) string {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	buf.WriteString(strings.ReplaceAll(tableName, " ", "_"))
	for _, columnName := range columnNames {
		buf.WriteString("_" + strings.ReplaceAll(columnName, " ", "_"))
	}
	switch nameType {
	case PRIMARY_KEY:
		buf.WriteString("_pkey")
	case FOREIGN_KEY:
		buf.WriteString("_fkey")
	case UNIQUE:
		buf.WriteString("_key")
	case INDEX:
		buf.WriteString("_idx")
	case CHECK:
		buf.WriteString("_check")
	case EXCLUDE:
		buf.WriteString("_excl")
	}
	return buf.String()
}

func defaultColumnType(dialect string, field sq.Field) (columnType string) {
	switch field.(type) {
	case sq.BlobField:
		switch dialect {
		case sq.DialectPostgres:
			return "BYTEA"
		default:
			return "BLOB"
		}
	case sq.BooleanField:
		return "BOOLEAN"
	case sq.JSONField:
		switch dialect {
		case sq.DialectPostgres:
			return "JSONB"
		default:
			return "JSON"
		}
	case sq.NumberField:
		return "INT"
	case sq.StringField:
		switch dialect {
		case sq.DialectPostgres, sq.DialectSQLite:
			return "TEXT"
		default:
			return "VARCHAR(255)"
		}
	case sq.TimeField:
		switch dialect {
		case sq.DialectPostgres:
			return "TIMESTAMPTZ"
		default:
			return "DATETIME"
		}
	}
	return "VARCHAR(255)"
}
