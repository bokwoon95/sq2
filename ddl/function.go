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
	ReturnType     string   `json:",omitempty"`
	SQL            string   `json:",omitempty"`
	IsIndependent  bool     `json:",omitempty"`
	Ignore         bool     `json:",omitempty"`
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
			rawArgs := strings.TrimSpace(remainder[i+1 : j])
			if rawArgs == "" {
				if token, tmp := popIdentifierToken(dialect, remainder[j+1:]); strings.EqualFold(token, "RETURNS") {
					fun.ReturnType, _ = popIdentifierToken(dialect, tmp)
				}
				break LOOP
			}
			args := splitArgs(rawArgs)
			fun.ArgModes = make([]string, len(args))
			fun.ArgNames = make([]string, len(args))
			fun.ArgTypes = make([]string, len(args))
			for i, arg := range args {
				tokens, _ := popIdentifierTokens(dialect, arg, -1)
				if len(tokens) == 0 {
					return fmt.Errorf("argument #%d ('%s') is invalid", i+1, arg)
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
					return fmt.Errorf("argument #%d ('%s', %#v) is invalid", i+1, arg, tokens)
				}
				// TODO: some argTypes may be split into multiple words, like
				// TIMESTAMP WITH TIME ZONE (how troublesome!)
				// known types:
				// TIMESTAMP WITH TIME ZONE
				// name TIMESTAMP WITH TIME ZONE
				// IN TIMESTAMP WITH TIME ZONE
				// IN name TIMESTAMP WITH TIME ZONE
				// name IN TIMESTAMP WITH TIME ZONE
				// unknown types:
				// emp
				// name emp
				// IN emp
				// IN name emp
				// name IN emp
				// start from the back and try to figure out if it is a known
				// type (that may span multiple tokens). if unknown type, then
				// just take the last token as the type. then look at the
				// number of tokens left: 0, 1 or 2? anything more, raise an
				// error. else you can run the loop below.
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
			if token, tmp := popIdentifierToken(dialect, remainder[j+1:]); strings.EqualFold(token, "RETURNS") {
				fun.ReturnType, _ = popIdentifierToken(dialect, tmp)
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
	if dialect == sq.DialectSQLite {
		return fmt.Errorf("sqlite does not support functions")
	}
	buf.WriteString(cmd.Function.SQL)
	return nil
}

type DropFunctionCommand struct {
	DropIfExists bool
	Function     Function
	DropCascade  bool
}

func (cmd DropFunctionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect == sq.DialectSQLite {
		return fmt.Errorf("sqlite does not support functions")
	}
	buf.WriteString("DROP FUNCTION ")
	if cmd.DropIfExists {
		buf.WriteString("IF EXISTS ")
	}
	if cmd.Function.FunctionSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Function.FunctionSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Function.FunctionName))
	if dialect == sq.DialectPostgres {
		buf.WriteString("(" + strings.Join(cmd.Function.ArgTypes, ", ") + ")")
	}
	if cmd.DropCascade {
		buf.WriteString(" CASCADE")
	}
	buf.WriteString(";")
	return nil
}

type RenameFunctionCommand struct {
	Function     Function
	RenameToName string
}

func (cmd RenameFunctionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect == sq.DialectSQLite || dialect == sq.DialectMySQL {
		return fmt.Errorf("%s does not support renaming functions", dialect)
	}
	buf.WriteString("ALTER FUNCTION ")
	if cmd.Function.FunctionSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Function.FunctionSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Function.FunctionName))
	if dialect == sq.DialectPostgres {
		buf.WriteString("(" + strings.Join(cmd.Function.ArgTypes, ", ") + ")")
	}
	buf.WriteString(" RENAME TO " + sq.QuoteIdentifier(dialect, cmd.RenameToName) + ";")
	return nil
}
