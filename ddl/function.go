package ddl

import (
	"bytes"
	"fmt"
	"strings"
)

type Function struct {
	FunctionSchema string   `json:",omitempty"`
	FunctionName   string   `json:",omitempty"`
	ArgModes       []string `json:",omitempty"`
	ArgNames       []string `json:",omitempty"`
	ArgTypes       []string `json:",omitempty"`
	SQL            string   `json:",omitempty"`
}

func (fun *Function) populateFunctionInfo(dialect string) error {
	const (
		PRE_FUNCTION = iota
		FUNCTION
	)
	word, rest := "", fun.SQL
	state := PRE_FUNCTION
LOOP:
	for rest != "" {
		switch state {
		case PRE_FUNCTION:
			word, rest = popWord(dialect, rest)
			if strings.EqualFold(word, "FUNCTION") {
				state = FUNCTION
			}
			continue
		case FUNCTION:
			fun.FunctionName, _ = popWord(dialect, rest)
			if i := strings.IndexByte(fun.FunctionName, '.'); i >= 0 {
				fun.FunctionSchema, fun.FunctionName = fun.FunctionName[:i], fun.FunctionName[i+1:]
			}
			if i := strings.IndexByte(fun.FunctionName, '('); i >= 0 {
				fun.FunctionName = fun.FunctionName[:i]
			}
			i := strings.IndexByte(rest, '(')
			if i < 0 {
				return fmt.Errorf("opening bracket for args not found")
			}
			j := strings.IndexByte(rest, ')')
			if j < 0 {
				return fmt.Errorf("closing bracket for args not found")
			}
			rawArgs := strings.Split(rest[i+1:j], ",")
			fun.ArgModes = make([]string, len(rawArgs))
			fun.ArgNames = make([]string, len(rawArgs))
			fun.ArgTypes = make([]string, len(rawArgs))
			for i, rawArg := range rawArgs {
				words, _ := popWords(dialect, rawArg, 3)
				if len(words) == 1 {
					argType := words[0]
					if j := strings.IndexByte(argType, '='); j >= 0 {
						argType = strings.TrimSpace(argType[:j])
					}
					continue
				}
				if strings.EqualFold(words[0], "IN") ||
					strings.EqualFold(words[0], "OUT") ||
					strings.EqualFold(words[0], "INTOUT") ||
					strings.EqualFold(words[0], "VARIADIC") {
					fun.ArgModes[i], words = words[0], words[1:]
				}
				fun.ArgNames[i], fun.ArgTypes[i] = words[0], words[1]
			}
			break LOOP
		}
	}
	if fun.SQL != "" && fun.FunctionName == "" {
		return fmt.Errorf("could not find function name, did you write the function correctly?")
	}
	return nil
}

type CreateFunctionCommand struct {
	Function Function
}

func (cmd CreateFunctionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString(cmd.Function.SQL)
	return nil
}

type DropFunctionCommand struct {
	DropIfExists bool
	Function     Function
	DropCascade  bool
}

type RenameFunctionCommand struct {
	Function     Function
	RenameToName string
}
