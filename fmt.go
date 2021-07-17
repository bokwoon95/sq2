package sq

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func BufferPrintf(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string, format string, values []interface{}) error {
	if i := strings.IndexByte(format, '{'); i < 0 {
		buf.WriteString(format)
		return nil
	}
	buf.Grow(len(format))
	runningValuesIndex := 0
	// valuesLookup is a map of the named parameters that are available for
	// reference in the args slice
	valuesLookup := make(map[string]int)
	for i, value := range values {
		switch value := value.(type) {
		case NamedParam:
			valuesLookup[value.Name] = i
		case sql.NamedArg:
			valuesLookup[value.Name] = i
		}
	}
	// ordinalNames track which ordinals are in use in the format string e.g.
	// {1}, {2}. The reason is because we are *temporarily* adding those into
	// the params map in order to track ordinal param status accross
	// BufferPrintValue calls. The reason we are tracking ordinal param status
	// across BufferPrintValue calls is because if the value for {1} has
	// already been appended into args, BufferPrintValue should not append the
	// value into args. But because ordinal param state is only tracked across
	// *BufferPrintValue* calls, not *BufferPrintf* calls, once BufferPrintf
	// exits all the ordinalNames added to the params map must be cleaned up.
	var ordinalNames []string
	// instead of looping over each rune in the format string, we jump straight
	// to each occurrence of '{'.
	for i := strings.IndexByte(format, '{'); i >= 0; i = strings.IndexByte(format, '{') {
		if i+2 <= len(format) && format[i:i+2] == "{{" {
			// To use a literal '{' in the format string, escape it by adding a
			// second curly brace after it i.e. '{{'. We treat all '{{' as '{'.
			buf.WriteString(format[:i])
			buf.WriteByte('{')
			format = format[i+2:]
			continue
		}
		buf.WriteString(format[:i])
		format = format[i:]
		j := strings.IndexByte(format, '}')
		if j < 0 {
			break
		}
		paramName := format[1:j] // if {1}, paramName=1. if {foobar}, paramName=foobar
		format = format[j+1:]
		var err error
		var value interface{}
		var modifierIndex map[string]int
		if i := strings.IndexByte(paramName, ':'); i >= 0 {
			var paramModifiers string
			paramName, paramModifiers = paramName[:i], paramName[i+1:]
			_, modifierIndex, err = lexModifiers(paramModifiers)
			if err != nil {
				return fmt.Errorf("lex %s: %w", paramModifiers, err)
			}
		}
		if paramName == "" {
			if runningValuesIndex >= len(values) {
				return fmt.Errorf("too few values passed in to BufferPrintf, expected more than %d", runningValuesIndex)
			}
			value = values[runningValuesIndex]
			runningValuesIndex++
		} else {
			num, err := strconv.Atoi(paramName)
			if err == nil {
				if num-1 < 0 || num-1 >= len(values) {
					return fmt.Errorf("ordinal parameter {%d} is out of bounds", num)
				}
				ordinalNames = append(ordinalNames, paramName)
				value = values[num-1]
			} else {
				num, ok := valuesLookup[paramName]
				if !ok {
					var availableParams []string
					for name := range valuesLookup {
						availableParams = append(availableParams, name)
					}
					sort.Strings(availableParams)
					return fmt.Errorf("named parameter {%s} not provided (available params: %s)", paramName, strings.Join(availableParams, ", "))
				}
				value = values[num]
			}
		}
		if _, ok := modifierIndex["nameonly"]; ok {
			switch v := value.(type) {
			case Field:
				value = Literal(v.GetName())
			case Table:
				value = Literal(v.GetName())
			}
		}
		err = BufferPrintValue(dialect, buf, args, params, excludedTableQualifiers, value, paramName)
		if err != nil {
			return err
		}
	}
	for _, ordinalName := range ordinalNames {
		delete(params, ordinalName)
	}
	buf.WriteString(format)
	return nil
}

func BufferPrintValue(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string, value interface{}, paramName string) error {
	if v, ok := value.(sql.NamedArg); ok {
		if dialect == DialectPostgres || dialect == DialectMySQL {
			return fmt.Errorf("%s does not support named parameters, please do not use sql.NamedArg", dialect)
		}
		if v.Name == "" {
			return fmt.Errorf("sql.NamedArg name cannot be empty")
		}
		if len(params[v.Name]) > 0 {
			(*args)[params[v.Name][0]] = value
		} else {
			params[v.Name] = []int{len(*args)}
			*args = append(*args, value)
		}
		switch dialect {
		case DialectSQLServer:
			buf.WriteString("@" + v.Name)
		case DialectSQLite:
			buf.WriteString("$" + v.Name)
		default:
			buf.WriteString(":" + v.Name)
		}
		return nil
	}
	if v, ok := value.(SQLExcludeAppender); ok && v != nil {
		return v.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
	}
	if v, ok := value.(SQLAppender); ok && v != nil {
		return v.AppendSQL(dialect, buf, args, params)
	}
	if isExplodableSlice(value) {
		return explodeSlice(dialect, buf, args, params, excludedTableQualifiers, value)
	}
	var paramIndices []int
	if paramName != "" {
		paramIndices = params[paramName]
	}
	paramIndex := -1
	if len(paramIndices) > 0 {
		paramIndex = paramIndices[0]
	}
	switch dialect {
	case DialectPostgres, DialectSQLite:
		if paramIndex >= 0 {
			buf.WriteString("$" + strconv.Itoa(paramIndex+1))
		} else {
			buf.WriteString("$" + strconv.Itoa(len(*args)+1))
			*args = append(*args, value)
			if paramName != "" {
				params[paramName] = []int{len(*args) - 1}
			}
		}
	case DialectSQLServer:
		if paramIndex >= 0 {
			buf.WriteString("@p" + strconv.Itoa(paramIndex+1))
		} else {
			buf.WriteString("@p" + strconv.Itoa(len(*args)+1))
			*args = append(*args, value)
			if paramName != "" {
				params[paramName] = []int{len(*args) - 1}
			}
		}
	default:
		buf.WriteString("?")
		*args = append(*args, value)
	}
	return nil
}

func Sprintf(dialect string, query string, args []interface{}) (string, error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	buf.Grow(len(query))
	namedArgsLookup := make(map[string]int)
	for i, arg := range args {
		if arg, ok := arg.(sql.NamedArg); ok {
			namedArgsLookup[arg.Name] = i
		}
	}
	runningArgsIndex := 0
	var insideStringOrIdentifier bool
	var openingQuote rune
	var mustWriteCharAt int
	var paramName []rune
	nameTerminatingChars := map[rune]bool{
		',': true, '(': true, ')': true, ';': true,
		'=': true, '>': true, '<': true,
		'+': true, '-': true, '*': true, '/': true,
		'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true, 0x85: true, 0xA0: true,
	}
	for i, char := range query {
		if mustWriteCharAt == i {
			buf.WriteRune(char)
			continue
		}
		if insideStringOrIdentifier {
			buf.WriteRune(char)
			switch openingQuote {
			case '\'', '"', '`':
				if char == openingQuote {
					if i+1 < len(query) && rune(query[i+1]) == openingQuote {
						mustWriteCharAt = i + 1
					} else {
						insideStringOrIdentifier = false
					}
				}
			case '[':
				if char == ']' {
					if i+1 < len(query) && query[i+1] == ']' {
						mustWriteCharAt = i + 1
					} else {
						insideStringOrIdentifier = false
					}
				}
			}
			continue
		}
		if char == '\'' || char == '"' {
			insideStringOrIdentifier = true
			openingQuote = char
			buf.WriteRune(char)
			continue
		}
		if char == '`' && dialect == DialectMySQL {
			insideStringOrIdentifier = true
			openingQuote = char
			buf.WriteRune(char)
			continue
		}
		if char == '[' && dialect == DialectSQLServer {
			insideStringOrIdentifier = true
			openingQuote = char
			buf.WriteRune(char)
			continue
		}
		// paramName will be non-empty only if the previous iteration inserted
		// a parameter-prefixing character (i.e. '?', '$', ':' or '@') into it
		if len(paramName) > 0 {
			if nameTerminatingChars[char] {
				paramValue, err := lookupParam(dialect, args, paramName, namedArgsLookup, runningArgsIndex)
				if err != nil {
					return buf.String(), err
				}
				buf.WriteString(paramValue + string(char))
				if len(paramName) == 1 && paramName[0] == '?' {
					runningArgsIndex++
				}
				paramName = paramName[:0]
			} else {
				paramName = append(paramName, char)
			}
			continue
		}
		if (char == '$' && (dialect == DialectSQLite || dialect == DialectPostgres)) ||
			(char == ':' && dialect == DialectSQLite) ||
			(char == '@' && (dialect == DialectSQLite || dialect == DialectSQLServer)) {
			paramName = append(paramName, char)
			continue
		}
		if char == '?' && dialect != DialectPostgres {
			if dialect == DialectSQLite {
				// for sqlite, just because we encounter a '?' doesn't mean it
				// is an anonymous param. sqlite also supports using '?' for
				// ordinal params (e.g. ?1, ?2, ?3) or named params (?foo,
				// ?bar, ?baz). Hence we treat it as an ordinal/named param
				// first, and handle the edge case later when it isn't.
				paramName = append(paramName, char)
				continue
			}
			if runningArgsIndex >= len(args) {
				return buf.String(), fmt.Errorf("too few args provided, expected more than %d", runningArgsIndex+1)
			}
			paramValue, err := Sprint(dialect, args[runningArgsIndex])
			if err != nil {
				return buf.String(), err
			}
			buf.WriteString(paramValue)
			runningArgsIndex++
			continue
		}
		buf.WriteRune(char)
	}
	if len(paramName) > 0 {
		paramValue, err := lookupParam(dialect, args, paramName, namedArgsLookup, runningArgsIndex)
		if err != nil {
			return buf.String(), err
		}
		buf.WriteString(paramValue)
	}
	if insideStringOrIdentifier {
		return buf.String(), fmt.Errorf("unclosed string or identifier")
	}
	return buf.String(), nil
}

func lookupParam(dialect string, args []interface{}, paramName []rune, namedArgsLookup map[string]int, runningArgsIndex int) (paramValue string, err error) {
	var maybeNum string
	if paramName[0] == '@' && dialect == DialectSQLServer && len(paramName) > 2 && (paramName[1] == 'p' || paramName[1] == 'P') {
		maybeNum = string(paramName[2:])
	} else {
		maybeNum = string(paramName[1:])
	}
	if maybeNum == "" {
		if paramName[0] != '?' {
			return "", fmt.Errorf("parameter name missing")
		}
		paramValue, err = Sprint(dialect, args[runningArgsIndex])
		if err != nil {
			return "", err
		}
		return paramValue, nil
	}
	// attempt to parse paramName as an ordinal parameter
	num, err := strconv.Atoi(maybeNum)
	if err == nil {
		num-- // decrement because ordinal parameters always lead the index by 1 (e.g. $1 corresponds to index 0)
		if num < 0 || num >= len(args) {
			return "", fmt.Errorf("args index %d out of bounds", num)
		}
		paramValue, err = Sprint(dialect, args[num])
		if err != nil {
			return "", err
		}
		return paramValue, nil
	}
	// if we reach here, we know that paramName is not an ordinal parameter
	if dialect == DialectPostgres || dialect == DialectMySQL {
		return "", fmt.Errorf("%s does not support %s named parameter", dialect, string(paramName))
	}
	num, ok := namedArgsLookup[string(paramName[1:])]
	if !ok {
		return "", fmt.Errorf("named parameter %s not provided", string(paramName))
	}
	if num < 0 || num >= len(args) {
		return "", fmt.Errorf("args index %d out of bounds", num)
	}
	paramValue, err = Sprint(dialect, args[num])
	if err != nil {
		return "", err
	}
	return paramValue, nil
}

func EscapeQuote(str string, quote byte) string {
	i := strings.IndexByte(str, quote)
	if i < 0 {
		return str
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	buf.Grow(len(str))
	escapedQuote := string([]byte{quote, quote})
	for i >= 0 {
		buf.WriteString(str[:i] + escapedQuote)
		if len(str[i:]) > 2 && str[i:i+2] == escapedQuote {
			str = str[i+2:]
		} else {
			str = str[i+1:]
		}
		i = strings.IndexByte(str, quote)
	}
	buf.WriteString(str)
	return buf.String()
}

func Sprint(dialect string, v interface{}) (string, error) {
	switch v := v.(type) {
	case nil:
		return "NULL", nil
	case bool:
		if v {
			return "TRUE", nil
		} else {
			return "FALSE", nil
		}
	case []byte:
		// TODO: Sprint needs to take in the dialect, because postgres BYTEA
		// literals requires a different format 🙄.
		return `x'` + hex.EncodeToString(v) + `'`, nil
	case string:
		return `'` + EscapeQuote(v, '\'') + `'`, nil
	case time.Time:
		return `'` + v.Format(time.RFC3339Nano) + `'`, nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 64), nil
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	case sql.NamedArg:
		return Sprint(dialect, v.Value)
	case sql.NullBool:
		if !v.Valid {
			return "NULL", nil
		} else {
			if v.Bool {
				return "TRUE", nil
			} else {
				return "FALSE", nil
			}
		}
	case sql.NullFloat64:
		if !v.Valid {
			return "NULL", nil
		} else {
			return strconv.FormatFloat(v.Float64, 'g', -1, 64), nil
		}
	case sql.NullInt64:
		if !v.Valid {
			return "NULL", nil
		} else {
			return strconv.FormatInt(v.Int64, 10), nil
		}
	case sql.NullInt32:
		if !v.Valid {
			return "NULL", nil
		} else {
			return strconv.FormatInt(int64(v.Int32), 10), nil
		}
	case sql.NullString:
		if !v.Valid {
			return "NULL", nil
		} else {
			return `'` + EscapeQuote(v.String, '\'') + `'`, nil
		}
	case sql.NullTime:
		if !v.Valid {
			return "NULL", nil
		} else {
			return `'` + v.Time.Format(time.RFC3339Nano) + `'`, nil
		}
	case driver.Valuer:
		vv, err := v.Value()
		if err != nil {
			return "", fmt.Errorf("error when calling Value(): %w", err)
		}
		switch vv := vv.(type) {
		case int64:
			return strconv.FormatInt(vv, 10), nil
		case float64:
			return strconv.FormatFloat(vv, 'g', -1, 64), nil
		case bool:
			if vv {
				return "TRUE", nil
			} else {
				return "FALSE", nil
			}
		case []byte:
			return `x'` + hex.EncodeToString(vv) + `'`, nil
		case string:
			return `'` + EscapeQuote(vv, '\'') + `'`, nil
		case time.Time:
			return `'` + vv.Format(time.RFC3339Nano) + `'`, nil
		default:
			return "", fmt.Errorf("unrecognized driver.Valuer type (must be one of int64, float64, bool, []byte, string, time.Time)")
		}
	}
	var deref int
	rv := reflect.ValueOf(v)
	// keep dereferencing until we are no longer at a pointer or interface type (i.e a concrete type)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
		deref++
	}
	if !rv.IsValid() {
		return "", fmt.Errorf("value is not valid (whatever that means??? Tell me how you got here)")
	}
	if rv.Kind() == reflect.Chan {
		return "", fmt.Errorf("Go channels cannot be represented in SQL")
	}
	if rv.Kind() == reflect.Func {
		return "", fmt.Errorf("Go functions cannot be represented in SQL")
	}
	if deref > 0 {
		return Sprint(dialect, rv.Interface())
	}
	return "", fmt.Errorf("could not convert %#v into its SQL representation", v)
}

type customTable struct {
	format string
	values []interface{}
}

var _ Table = customTable{}

func Tablef(format string, values ...interface{}) Table {
	return customTable{format: format, values: values}
}

func (tbl customTable) GetAlias() string { return "" }

func (tbl customTable) GetName() string { return "" }

func (tbl customTable) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	return BufferPrintf(dialect, buf, args, params, nil, tbl.format, tbl.values)
}

type customQuery struct {
	dialect string
	format  string
	values  []interface{}
}

var _ Query = customQuery{}

func Queryf(format string, values ...interface{}) Query {
	return customQuery{format: format, values: values}
}

func (d SQLiteDialect) Queryf(format string, values ...interface{}) Query {
	return customQuery{dialect: DialectSQLite, format: format, values: values}
}

func (d PostgresDialect) Queryf(format string, values ...interface{}) Query {
	return customQuery{dialect: DialectPostgres, format: format, values: values}
}

func (d MySQLDialect) Queryf(format string, values ...interface{}) Query {
	return customQuery{dialect: DialectMySQL, format: format, values: values}
}

func (q customQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	return BufferPrintf(dialect, buf, args, params, nil, q.format, q.values)
}

func (q customQuery) SetFetchableFields([]Field) (Query, error) {
	return nil, fmt.Errorf("custom %w", ErrNonFetchableQuery)
}

func (q customQuery) GetFetchableFields() ([]Field, error) {
	return nil, fmt.Errorf("custom %w", ErrNonFetchableQuery)
}

func (q customQuery) GetDialect() string { return q.dialect }
