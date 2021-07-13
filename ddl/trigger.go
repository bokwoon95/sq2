package ddl

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
}

func getTriggerInfo(dialect, sql string) (tableSchema, tableName, triggerName string, err error) {
	const (
		PRE_TRIGGER = iota
		TRIGGER
		PRE_ON
		ON
	)
	word, rest := "", sql
	state := PRE_TRIGGER
LOOP:
	for rest != "" {
		switch state {
		case PRE_TRIGGER:
			word, rest = popWord(dialect, rest)
			if strings.EqualFold(word, "TRIGGER") {
				state = TRIGGER
			}
			continue
		case TRIGGER:
			if dialect == sq.DialectSQLite {
				words, tmp := popWords(dialect, rest, 3)
				if len(words) == 3 &&
					strings.EqualFold(words[0], "IF") &&
					strings.EqualFold(words[1], "NOT") &&
					strings.EqualFold(words[2], "EXISTS") {
					rest = tmp
				}
			}
			triggerName, rest = popWord(dialect, rest)
			state = PRE_ON
			continue
		case PRE_ON:
			word, rest = popWord(dialect, rest)
			if strings.EqualFold(word, "ON") {
				state = ON
			}
			continue
		case ON:
			tableName, rest = popWord(dialect, rest)
			if i := strings.IndexByte(tableName, '.'); i >= 0 {
				tableSchema, tableName = tableName[:i], tableName[i+1:]
			}
			break LOOP
		}
	}
	if triggerName == "" || tableName == "" {
		return tableSchema, tableName, triggerName, fmt.Errorf("could not find trigger name or table name, did you write the trigger correctly?")
	}
	return tableSchema, tableName, triggerName, nil
}

type TriggerDiff struct {
	TableSchema   string
	TableName     string
	TriggerName   string
	CreateCommand *CreateTriggerCommand
	DropCommand   *DropTriggerCommand
	RenameCommand *RenameTriggerCommand
}

type CreateTriggerCommand struct {
	Trigger Trigger
}

func (cmd *CreateTriggerCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
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

type RenameTriggerCommand struct {
	TableSchema  string
	TableName    string
	RenameToName string
}
