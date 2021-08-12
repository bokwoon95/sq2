package ddl

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"

	"github.com/bokwoon95/sq"
)

type Function struct {
	FunctionSchema string `json:",omitempty"`
	FunctionName   string `json:",omitempty"`
	RawArgs        string `json:",omitempty"`
	ReturnType     string `json:",omitempty"`
	SQL            string `json:",omitempty"`
	Comment        string `json:",omitempty"`
	Ignore         bool   `json:",omitempty"`
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
			token, remainder, _ = popIdentifierToken(dialect, remainder)
			if strings.EqualFold(token, "FUNCTION") {
				state = FUNCTION
			}
			continue
		case FUNCTION:
			fun.FunctionName, _, _ = popIdentifierToken(dialect, remainder)
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
			fun.RawArgs = strings.TrimSpace(remainder[i+1 : j])
			if token, tmp, _ := popIdentifierToken(dialect, remainder[j+1:]); strings.EqualFold(token, "RETURNS") {
				fun.ReturnType, _, _ = popIdentifierToken(dialect, tmp)
			}
			break LOOP
		}
	}
	if fun.SQL != "" && fun.FunctionName == "" {
		return fmt.Errorf("could not find function name, did you write the function correctly?")
	}
	return nil
}

func FilesToFunctions(dialect string, fsys fs.FS, filenames ...string) ([]Function, error) {
	var functions []Function
	for _, filename := range filenames {
		b, err := fs.ReadFile(fsys, filename)
		if err != nil {
			return nil, fmt.Errorf("reading file %s: %w", filename, err)
		}
		function := Function{SQL: string(b)}
		err = function.populateFunctionInfo(dialect)
		if err != nil {
			return nil, fmt.Errorf("populating function from %s: %w", filename, err)
		}
		functions = append(functions, function)
	}
	return functions, nil
}

type CreateFunctionCommand struct {
	Function Function
	Ignore   bool
}

func (cmd CreateFunctionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	if cmd.Ignore {
		return nil
	}
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
	Ignore       bool
}

func (cmd DropFunctionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	if cmd.Ignore {
		return nil
	}
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
		buf.WriteString("(" + cmd.Function.RawArgs + ")")
	}
	if cmd.DropCascade {
		buf.WriteString(" CASCADE")
	}
	return nil
}
