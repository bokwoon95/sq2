package sq

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func BufferPrintf(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string, format string, values []interface{}) error {
	buf.Grow(len(format))
	runningValuesIndex := 0
	valuesLookup := make(map[string]int)
	for i, value := range values {
		switch value := value.(type) {
		case NamedParam:
			valuesLookup[value.Name] = i
		case sql.NamedArg:
			valuesLookup[value.Name] = i
		}
	}
	var ordinalNames []string
	for i := strings.IndexByte(format, '{'); i >= 0; i = strings.IndexByte(format, '{') {
		if i > 0 && format[i-1] == '\\' {
			buf.WriteString(format[:i-1])
			buf.WriteByte('{')
			format = format[i+1:]
			continue
		}
		buf.WriteString(format[:i])
		format = format[i:]
		j := strings.IndexByte(format, '}')
		if j < 0 {
			break
		}
		paramName := format[1:j]
		format = format[j+1:]
		var value interface{}
		if paramName == "" {
			if runningValuesIndex >= len(values) {
				return fmt.Errorf("too few values passed in, expected more than %d", runningValuesIndex)
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
					return fmt.Errorf("named parameter {%s} not provided", paramName)
				}
				value = values[num]
			}
		}
		err := BufferPrintValue(dialect, buf, args, params, excludedTableQualifiers, value, paramName)
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

func BufferPrintValue(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string, value interface{}, name string) error {
	if v, ok := value.(sql.NamedArg); ok {
		if v.Name == "" {
			return fmt.Errorf("sql.NamedArg name cannot be empty")
		}
		if dialect == DialectPostgres || dialect == DialectMySQL {
			return fmt.Errorf("your database dialect does not support named parameters, please do not use sql.NamedArg")
		}
		if len(params[v.Name]) > 0 {
			(*args)[params[v.Name][0]] = value
		} else {
			params[v.Name] = []int{len(*args)}
			*args = append(*args, value)
		}
		switch dialect {
		case DialectMSSQL:
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
		explodeSlice(dialect, buf, args, params, excludedTableQualifiers, value)
		return nil
	}
	switch dialect {
	case DialectPostgres, DialectSQLite:
		if name != "" && len(params[name]) > 0 {
			buf.WriteString("$" + strconv.Itoa(params[name][0]+1))
			return nil
		} else {
			buf.WriteString("$" + strconv.Itoa(len(*args)+1))
		}
	case DialectMSSQL:
		if name != "" && len(params[name]) > 0 {
			buf.WriteString("@p" + strconv.Itoa(params[name][0]+1))
			return nil
		} else {
			buf.WriteString("@p" + strconv.Itoa(len(*args)+1))
		}
	default:
		buf.WriteString("?")
	}
	if name != "" {
		params[name] = []int{len(*args)}
	}
	*args = append(*args, value)
	return nil
}

func lookupParam(dialect string, args []interface{}, argsLookup map[string]int, namebuf *[]rune, runningArgsIndex *int) (paramValue string, err error) {
	defer func() { *namebuf = (*namebuf)[:0] }()
	if (*namebuf)[0] == '@' && dialect == DialectMSSQL {
		// TODO: implement MSSQL support
	}
	name := string((*namebuf)[1:])
	if name == "" {
		if (*namebuf)[0] != '?' {
			return "", fmt.Errorf("sq: parameter name missing")
		}
		paramValue, err = Sprint(args[*runningArgsIndex])
		if err != nil {
			return "", err
		}
		(*runningArgsIndex)++
		return paramValue, nil
	}
	num, err := strconv.Atoi(name)
	if err == nil {
		num-- // decrement because ordinal numbers always lead the index by 1 (e.g. $1 corresponds to index 0)
		if num < 0 || num >= len(args) {
			return "", fmt.Errorf("sq: args index %d out of bounds", num)
		}
		paramValue, err = Sprint(args[num])
		if err != nil {
			return "", err
		}
		return paramValue, nil
	}
	if dialect == DialectPostgres {
		return "", fmt.Errorf("sq: Postgres does not support $%s named parameter", name)
	}
	num, ok := argsLookup[name]
	if !ok {
		return "", fmt.Errorf("sq: named parameter $%s not provided", name)
	}
	if num < 0 || num >= len(args) {
		return "", fmt.Errorf("sq: args index %d out of bounds", num)
	}
	paramValue, err = Sprint(args[num])
	if err != nil {
		return "", err
	}
	return paramValue, nil
}

func Sprintf(dialect string, query string, args []interface{}) (string, error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	buf.Grow(len(query))
	argsLookup := make(map[string]int)
	for i, arg := range args {
		if arg, ok := arg.(sql.NamedArg); ok {
			argsLookup[arg.Name] = i
		}
	}
	runningArgsIndex := 0
	var insideString bool
	var insideIdentifier bool
	var namebuf []rune
	nameTerminatingChars := map[rune]bool{
		',': true, '(': true, ')': true, ';': true,
		'=': true, '>': true, '<': true,
		'+': true, '-': true, '*': true, '/': true,
		'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true, 0x85: true, 0xA0: true,
	}
	for _, char := range query {
		if char == '\'' && !insideIdentifier {
			insideString = !insideString
			buf.WriteRune(char)
			continue
		}
		if char == '"' && !insideString {
			insideIdentifier = !insideIdentifier
			buf.WriteRune(char)
			continue
		}
		if insideString || insideIdentifier {
			buf.WriteRune(char)
			continue
		}
		// If namebuf is non-empty, it means we are inside a parameter name.
		// This is because the first character will be inserted into namebuf
		// only if the previous iteration encounter a parameter-related
		// character (i.e. '?', '$', ':' or '@')
		if len(namebuf) > 0 {
			if !nameTerminatingChars[char] {
				namebuf = append(namebuf, char)
			} else {
				paramValue, err := lookupParam(dialect, args, argsLookup, &namebuf, &runningArgsIndex)
				if err != nil {
					return buf.String(), err
				}
				buf.WriteString(paramValue + string(char))
			}
			continue
		}
		switch {
		case char == '$' && (dialect == DialectSQLite || dialect == DialectPostgres),
			char == ':' && dialect == DialectSQLite,
			char == '@' && (dialect == DialectSQLite || dialect == DialectMSSQL),
			char == '?' && dialect == DialectSQLite:
			namebuf = append(namebuf, char)
			continue
		case char == '?' && dialect != DialectPostgres && dialect != DialectMSSQL:
			if runningArgsIndex < 0 || runningArgsIndex >= len(args) {
				return buf.String(), fmt.Errorf("sq: too few args provided, expected more than %d", runningArgsIndex+1)
			}
			paramValue, err := Sprint(args[runningArgsIndex])
			if err != nil {
				return buf.String(), err
			}
			buf.WriteString(paramValue)
			runningArgsIndex++
			continue
		}
		buf.WriteRune(char)
	}
	if len(namebuf) > 0 {
		paramValue, err := lookupParam(dialect, args, argsLookup, &namebuf, &runningArgsIndex)
		if err != nil {
			return buf.String(), err
		}
		buf.WriteString(paramValue)
	}
	if insideString || insideIdentifier {
		// means something went wrong, unclosed quote somewhere
	}
	return buf.String(), nil
}

func escapeSQLSingleQuote(s string) string {
	i := strings.IndexByte(s, '\'')
	if i < 0 {
		return s
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	buf.Grow(len(s))
	for i >= 0 {
		buf.WriteString(s[:i] + `''`)
		if len(s[i:]) > 2 && s[i:i+2] == `''` {
			s = s[i+2:]
		} else {
			s = s[i+1:]
		}
		i = strings.IndexByte(s, '\'')
	}
	buf.WriteString(s)
	return buf.String()
}

func Sprint(v interface{}) (string, error) {
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
		return `x'` + hex.EncodeToString(v) + `'`, nil
	case string:
		return `'` + escapeSQLSingleQuote(v) + `'`, nil
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
		return Sprint(v.Value)
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
			return `'` + escapeSQLSingleQuote(v.String) + `'`, nil
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
			return "", fmt.Errorf("sq: error when calling Value(): %w", err)
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
			return `'` + vv + `'`, nil
		case time.Time:
			return `'` + vv.Format(time.RFC3339Nano) + `'`, nil
		default:
			return "", fmt.Errorf("sq: unrecognized driver.Valuer type (must be one of int64, float64, bool, []byte, string, time.Time)")
		}
	}
	var deref int
	rv := reflect.ValueOf(v)
	// keep dereferencing until we are no longer at a pointer or interface type (i.e a concrete type)
	for rv.Kind() != reflect.Ptr && rv.Kind() != reflect.Interface {
		rv = rv.Elem()
		deref++
	}
	if !rv.IsValid() {
		return "", fmt.Errorf("sq: value is not valid (whatever that means??? Tell me how you got here)")
	}
	if rv.Kind() == reflect.Chan {
		return "", fmt.Errorf("sq: channels cannot be represented in SQL")
	}
	if rv.Kind() == reflect.Func {
		return "", fmt.Errorf("sq: functions cannot be represented in SQL")
	}
	if deref > 0 {
		return Sprint(rv.Interface())
	}
	return "", fmt.Errorf("sq: could not convert %#v into its SQL representation", v)
}
