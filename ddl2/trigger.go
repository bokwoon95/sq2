package ddl2

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

type Trigger struct {
	TableSchema string `json:",omitempty"`
	TableName   string `json:",omitempty"`
	TriggerName string `json:",omitempty"`
	SQL         string `json:",omitempty"`
	Ignore      bool   `json:",omitempty"`
}

func (trg *Trigger) populateTriggerInfo(dialect string) error {
	const (
		PRE_TRIGGER = iota
		TRIGGER
		PRE_ON
		ON
	)
	word, rest := "", trg.SQL
	state := PRE_TRIGGER
LOOP:
	for rest != "" {
		switch state {
		case PRE_TRIGGER:
			word, rest, _ = popIdentifierToken(dialect, rest)
			if strings.EqualFold(word, "TRIGGER") {
				state = TRIGGER
			}
			continue
		case TRIGGER:
			if dialect == sq.DialectSQLite {
				words, tmp, _ := popIdentifierTokens(dialect, rest, 3)
				if len(words) == 3 &&
					strings.EqualFold(words[0], "IF") &&
					strings.EqualFold(words[1], "NOT") &&
					strings.EqualFold(words[2], "EXISTS") {
					rest = tmp
				}
			}
			trg.TriggerName, rest, _ = popIdentifierToken(dialect, rest)
			state = PRE_ON
			continue
		case PRE_ON:
			word, rest, _ = popIdentifierToken(dialect, rest)
			if strings.EqualFold(word, "ON") {
				state = ON
			}
			continue
		case ON:
			trg.TableName, rest, _ = popIdentifierToken(dialect, rest)
			if i := strings.IndexByte(trg.TableName, '.'); i >= 0 {
				trg.TableSchema, trg.TableName = trg.TableName[:i], trg.TableName[i+1:]
			}
			break LOOP
		}
	}
	if trg.SQL != "" && (trg.TriggerName == "" || trg.TableName == "") {
		return fmt.Errorf("could not find trigger name or table name, did you write the trigger correctly?")
	}
	return nil
}

type CreateTriggerCommand struct {
	Trigger Trigger
}

func (cmd CreateTriggerCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString(cmd.Trigger.SQL)
	return nil
}

type DropTriggerCommand struct {
	DropIfExists bool
	TableSchema  string
	TableName    string
	TriggerName  string
	DropCascade  bool
}

func (cmd DropTriggerCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("DROP TRIGGER ")
	if cmd.DropIfExists {
		buf.WriteString("IF EXISTS ")
	}
	switch dialect {
	case sq.DialectPostgres:
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TriggerName) + " ON ")
		if cmd.TableSchema != "" {
			buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TableSchema) + ".")
		}
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TableName))
		if cmd.DropCascade {
			buf.WriteString(" CASCADE")
		}
	default:
		if cmd.TableSchema != "" {
			buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TableSchema) + ".")
		}
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TriggerName))
	}
	return nil
}

type RenameTriggerCommand struct {
	TableSchema  string
	TableName    string
	TriggerName  string
	RenameToName string
}

func (cmd RenameTriggerCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect == sq.DialectSQLite || dialect == sq.DialectMySQL {
		return fmt.Errorf("%s does not support renaming triggers", dialect)
	}
	buf.WriteString("ALTER TRIGGER " + sq.QuoteIdentifier(dialect, cmd.TriggerName) + " ON ")
	if cmd.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TableName) + " RENAME TO " + sq.QuoteIdentifier(dialect, cmd.RenameToName))
	return nil
}
