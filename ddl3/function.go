package ddl3

import "strings"

/*
I no longer care about one Function object being able to contain every
overloaded function. If the user has multiple overloaded functions, there will
be multiple Function structs with the same function schema and function name.
TODO: Instead of CachedFunctionIndex(functionSchema, functionName string) int,
there will be CachedFunctionIndices(functionSchema, functionName string) []int
catalog.LoadFunctionFS(functionSchema, functionName string, fsys fs.FS, filename string)
catalog.LoadFunction(Function{
	FunctionSchema: "",
	FunctionName: "",
	Contents: "",
})
*/

// GetFunctionName -> functionSchema, functionName, (functionArgs?)

type Function struct {
	FunctionSchema string
	FunctionName   string
	ContainsTable  bool
	SQL            string
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
			break LOOP
		}
	}
	return functionSchema, functionName, nil
}

// all IFunctions can be converted into Functions. An IFunction is simply a
// struct container that contains the function definition in a method. The
// struct itself can be used as a table in sq queries (i.e. a table-valued
// function).
// the fn *ddl.Fn object passed to the Function() method can be used to
// register statements, one by one. This allows the user to reuse sq queries
// inside table valued functions. This means that it Select queries must
// additionally support the INTO clause.
// alternatively, the user can just call fn.FromFS(fsys fs.FS, filename string)
// type IFunction interface {
// 	GetSchema() string
// 	GetName() string
// 	GetArgs() (argModes, argNames, argtypes []string)
// 	GetSource() io.Reader
// }

// NOT: To play it safe, do not implement IFunction first. Users can only define
// functions view catalog.LoadFunctionFromFS or catalog.LoadFunction. I
// honestly don't like the current solution for table valued functions.
