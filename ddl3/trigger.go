package ddl3

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bokwoon95/sq"
)

type Trigger struct {
	TableSchema string
	TableName   string
	TriggerName string
	SQL         string
}

func popWord(s string) (word, rest string) {
	s = strings.TrimLeft(s, " \t\n\v\f\r\u0085\u00A0")
	if s == "" {
		return "", ""
	}
	var splitAt int
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		splitAt = i
		if unicode.IsSpace(r) {
			splitAt -= size
			break
		}
	}
	return s[:splitAt], s[splitAt:]
}

func popWords(s string, num int) (words []string, rest string) {
	word, rest := "", s
	for i := 0; i < num && rest != ""; i++ {
		word, rest = popWord(rest)
		words = append(words, word)
	}
	return words, rest
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
			word, rest = popWord(rest)
			if strings.EqualFold(word, "TRIGGER") {
				state = TRIGGER
			}
			continue
		case TRIGGER:
			if dialect == sq.DialectSQLite {
				words, tmp := popWords(rest, 3)
				if len(words) == 3 &&
					strings.EqualFold(words[0], "IF") &&
					strings.EqualFold(words[1], "NOT") &&
					strings.EqualFold(words[2], "EXISTS") {
					rest = tmp
				}
			}
			triggerName, rest = popWord(rest)
			state = PRE_ON
			continue
		case PRE_ON:
			word, rest = popWord(rest)
			if strings.EqualFold(word, "ON") {
				state = ON
			}
			continue
		case ON:
			tableName, rest = popWord(rest)
			if i := strings.IndexByte(tableName, '.'); i >= 0 {
				tableSchema, tableName = tableName[:i], tableName[i+1:]
			}
			break LOOP
		}
	}
	return tableSchema, tableName, triggerName, nil
}

// catalog.LoadTriggerFS(triggerName string, fsys fs.FS, filename string)
// catalog.LoadTrigger(ddl.Trigger{TriggerName: "", Contents: ""})
