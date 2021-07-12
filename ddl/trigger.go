package ddl

import (
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

type Trigger struct {
	TableSchema string
	TableName   string
	TriggerName string
	SQL         string
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