package ddl

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
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
	if dialect != sq.DialectPostgres {
		return nil
	}
	state := PRE_FUNCTION
	token, remainder := "", fun.SQL
LOOP:
	for remainder != "" {
		switch state {
		case PRE_FUNCTION:
			token, remainder = popIdentifierToken(dialect, remainder)
			if strings.EqualFold(token, "FUNCTION") {
				state = FUNCTION
			}
			continue
		case FUNCTION:
			fun.FunctionName, _ = popIdentifierToken(dialect, remainder)
			if i := strings.IndexByte(fun.FunctionName, '.'); i >= 0 {
				fun.FunctionSchema, fun.FunctionName = fun.FunctionName[:i], fun.FunctionName[i+1:]
			}
			if i := strings.IndexByte(fun.FunctionName, '('); i >= 0 {
				fun.FunctionName = fun.FunctionName[:i]
			}
			i := strings.IndexByte(remainder, '(')
			if i < 0 {
				return fmt.Errorf("opening bracket for args not found")
			}
			j := strings.IndexByte(remainder, ')')
			if j < 0 {
				return fmt.Errorf("closing bracket for args not found")
			}
			if i+1 == j {
				break LOOP
			}
			rawArgs := splitArgs(remainder[i+1 : j])
			fun.ArgModes = make([]string, len(rawArgs))
			fun.ArgNames = make([]string, len(rawArgs))
			fun.ArgTypes = make([]string, len(rawArgs))
			for i, rawArg := range rawArgs {
				tokens, _ := popIdentifierTokens(dialect, rawArg, 4)
				if len(tokens) == 0 {
					return fmt.Errorf("argument #%d ('%s') is invalid", i+1, rawArg)
				}
				// This loop filters out the tokens we are not interested in.
				// we are only interested in tokens that contain the ArgMode,
				// ArgName or ArgType.
				for j := len(tokens) - 1; j >= 0; j-- {
					// If a token is DEFAULT or starts with '=', everything
					// after it (including the token itself) is filtered out.
					if strings.EqualFold(tokens[j], "DEFAULT") || tokens[j][0] == '=' {
						tokens = tokens[:j]
						break
					}
					// If a token contains '=' (but does not start with it) it
					// is assumed to contain the ArgType. Everything after the
					// token is filtered out. We will strip the extraneous
					// characters after the '=' later on down below.
					if strings.IndexByte(tokens[j], '=') >= 0 {
						tokens = tokens[:j+1]
						break
					}
				}
				if len(tokens) == 0 {
					return fmt.Errorf("argument #%d ('%s', %#v) is invalid", i+1, rawArg, tokens)
				}
				fun.ArgTypes[i] = tokens[len(tokens)-1]
				tokens = tokens[:len(tokens)-1]
				for _, token := range tokens {
					if strings.EqualFold(token, "IN") ||
						strings.EqualFold(token, "OUT") ||
						strings.EqualFold(token, "INOUT") ||
						strings.EqualFold(token, "VARIADIC") {
						fun.ArgModes[i] = token
					} else {
						fun.ArgNames[i] = token
					}
				}
				if j := strings.IndexByte(fun.ArgTypes[i], '='); j >= 0 {
					fun.ArgTypes[i] = fun.ArgTypes[i][:j]
				}
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

func (cmd DropFunctionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	return nil
}

type RenameFunctionCommand struct {
	Function     Function
	RenameToName string
}
