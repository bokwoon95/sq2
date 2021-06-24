package sq

import (
	"bytes"
	"fmt"
	"strconv"
)

type Params map[string]interface{}

type NamedParam struct {
	Name  string
	Value interface{}
	_     struct{}
}

var _ Field = NamedParam{}

func Param(name string, value interface{}) NamedParam {
	return NamedParam{Name: name, Value: value}
}

func (param NamedParam) GetAlias() string { return "" }

func (param NamedParam) GetName() string { return "" }

func (param NamedParam) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	if param.Name == "" {
		return fmt.Errorf("sq: Param name cannot be empty")
	}
	if v, ok := param.Value.(SQLExcludeAppender); ok && v != nil {
		return v.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
	}
	if v, ok := param.Value.(SQLAppender); ok && v != nil {
		return v.AppendSQL(dialect, buf, args, params)
	}
	if isExplodableSlice(param.Value) {
		return explodeSlice(dialect, buf, args, params, excludedTableQualifiers, param.Value)
	}
	switch dialect {
	case DialectPostgres, DialectSQLite:
		if len(params[param.Name]) > 0 {
			(*args)[params[param.Name][0]] = param.Value
			buf.WriteString("$" + strconv.Itoa(params[param.Name][0]+1))
		} else {
			params[param.Name] = []int{len(*args)}
			buf.WriteString("$" + strconv.Itoa(len(*args)+1))
			*args = append(*args, param.Value)
		}
	case DialectSQLServer:
		if len(params[param.Name]) > 0 {
			(*args)[params[param.Name][0]] = param.Value
			buf.WriteString("@p" + strconv.Itoa(params[param.Name][0]+1))
		} else {
			params[param.Name] = []int{len(*args)}
			buf.WriteString("@p" + strconv.Itoa(len(*args)+1))
			*args = append(*args, param.Value)
		}
	default:
		params[param.Name] = append(params[param.Name], len(*args))
		buf.WriteString("?")
		*args = append(*args, param.Value)
	}
	return nil
}
