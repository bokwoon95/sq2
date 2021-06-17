package sq

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
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
		parameterName := format[1:j]
		format = format[j+1:]
		var value interface{}
		if parameterName == "" {
			if runningValuesIndex >= len(values) {
				return fmt.Errorf("too few values passed in, expected more than %d", runningValuesIndex)
			}
			value = values[runningValuesIndex]
			runningValuesIndex++
		} else {
			num, err := strconv.Atoi(parameterName)
			if err == nil {
				if num-1 < 0 || num-1 >= len(values) {
					return fmt.Errorf("ordinal parameter {%d} is out of bounds", num)
				}
				ordinalNames = append(ordinalNames, parameterName)
				value = values[num-1]
			} else {
				num, ok := valuesLookup[parameterName]
				if !ok {
					return fmt.Errorf("named parameter {%s} not provided", parameterName)
				}
				value = values[num]
			}
		}
		err := BufferPrintValue(dialect, buf, args, params, excludedTableQualifiers, value, parameterName)
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
		default:
			buf.WriteString("$" + v.Name)
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

func Sprintf(dialect string, query string, args []interface{}) string {
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
	var namebuf []rune
	nameTerminatingChars := map[rune]bool{
		',': true, '(': true, ')': true, ';': true,
		'=': true, '>': true, '<': true,
		'+': true, '-': true, '*': true, '/': true,
		'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true, 0x85: true, 0xA0: true,
	}
	for _, char := range query {
		if char == '\'' {
			insideString = !insideString
			buf.WriteRune(char)
			continue
		}
		if insideString {
			buf.WriteRune(char)
			continue
		}
		if len(namebuf) > 0 && nameTerminatingChars[char] {
			if namebuf[0] == '@' && dialect == DialectMSSQL {
			}
			name := string(namebuf[1:])
			if name == "" {
				if namebuf[0] == '?' {
					buf.WriteString(Sprint(args[runningArgsIndex]))
					runningArgsIndex++
				} else {
					buf.WriteString("%!(MISSING)")
				}
			} else if num, err := strconv.Atoi(name); err == nil {
				if num-1 < 0 || num-1 >= len(args) {
					buf.WriteString("%!(BADINDEX)")
				} else {
					buf.WriteString(Sprint(args[num-1]))
				}
			} else {
				num, ok := argsLookup[name]
				if dialect == DialectPostgres {
					buf.WriteString("%!(INVALID)")
				} else if !ok {
					buf.WriteString("%!(MISSING)")
				} else if num < 0 || num >= len(args) {
					buf.WriteString("%!(BADINDEX)")
				} else {
					buf.WriteString(Sprint(args[num]))
				}
			}
			buf.WriteRune(char)
			namebuf = namebuf[:0]
			continue
		}
		if len(namebuf) > 0 {
			namebuf = append(namebuf, char)
			continue
		}
		switch {
		case char == '$' && (dialect == DialectSQLite || dialect == DialectPostgres),
			char == ':' && dialect == DialectSQLite,
			char == '@' && (dialect == DialectSQLite || dialect == DialectMSSQL),
			char == '?' && dialect == DialectSQLite:
			namebuf = append(namebuf, char)
			continue
		case char == '?' && dialect == DialectMySQL:
			if runningArgsIndex < 0 || runningArgsIndex >= len(args) {
				buf.WriteString("%!(BADINDEX)")
			} else {
				buf.WriteString(Sprint(args[runningArgsIndex]))
				runningArgsIndex++
			}
			continue
		}
		buf.WriteRune(char)
	}
	if len(namebuf) > 0 {
		name := string(namebuf[1:])
		if name == "" {
			if namebuf[0] == '?' {
				buf.WriteString(Sprint(args[runningArgsIndex]))
				runningArgsIndex++
			} else {
				buf.WriteString("%!(MISSING)")
			}
		} else if num, err := strconv.Atoi(name); err == nil {
			if num-1 < 0 || num-1 >= len(args) {
				buf.WriteString("%!(BADINDEX)")
			} else {
				buf.WriteString(Sprint(args[num-1]))
			}
		} else {
			num, ok := argsLookup[name]
			if dialect == DialectPostgres {
				buf.WriteString("%!(INVALID)")
			} else if !ok {
				buf.WriteString("%!(MISSING)")
			} else if num < 0 || num >= len(args) {
				buf.WriteString("%!(BADINDEX)")
			} else {
				buf.WriteString(Sprint(args[num]))
			}
		}
	}
	return buf.String()
}

func Sprint(v interface{}) string {
	switch v := v.(type) {
	case sql.NamedArg:
		return Sprint(v.Value)
	case nil:
		return "NULL"
	case bool:
		if v {
			return "TRUE"
		} else {
			return "FALSE"
		}
	case []byte:
		return `x'` + hex.EncodeToString(v) + `'`
	case string:
		return `'` + v + `'`
	case time.Time:
		return `'` + v.Format(time.RFC3339Nano) + `'`
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case driver.Valuer:
		v2, err := v.Value()
		if err != nil {
			return `%!(` + err.Error() + `)`
		}
		switch v2 := v2.(type) {
		case int64:
			return strconv.FormatInt(v2, 10)
		case float64:
			return strconv.FormatFloat(v2, 'g', -1, 64)
		case bool:
			if v2 {
				return "TRUE"
			} else {
				return "FALSE"
			}
		case []byte:
			return `x'` + hex.EncodeToString(v2) + `'`
		case string:
			return `'` + v2 + `'`
		case time.Time:
			return `'` + v2.Format(time.RFC3339Nano) + `'`
		default:
			return fmt.Sprintf("%#v", v2)
		}
	}
	return fmt.Sprintf("%#v", v)
}
