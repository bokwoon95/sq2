package ddl2

import (
	"fmt"
	"strings"
)

type Function struct {
	FunctionSchema string `json:",omitempty"`
	FunctionName   string `json:",omitempty"`
	IsDependent    bool   `json:",omitempty"`
	SQL            string `json:",omitempty"`
}

func getFunctionInfo(dialect, sql string) (functionSchema, functionName string, err error) {
	const (
		PRE_FUNCTION = iota
		FUNCTION
	)
	word, rest := "", sql
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
			functionName, rest = popWord(dialect, rest)
			if i := strings.IndexByte(functionName, '.'); i >= 0 {
				functionSchema, functionName = functionName[:i], functionName[i+1:]
			}
			if i := strings.IndexByte(functionName, '('); i >= 0 {
				functionName = functionName[:i]
			}
			break LOOP
		}
	}
	if functionName == "" {
		return functionSchema, functionName, fmt.Errorf("could not find function name, did you write the function correctly?")
	}
	return functionSchema, functionName, nil
}
