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

func (param NamedParam) GetAlias() string {
	if v, ok := param.Value.(interface{ GetAlias() string }); ok {
		return v.GetAlias()
	}
	return ""
}

func (param NamedParam) GetName() string {
	if v, ok := param.Value.(interface{ GetName() string }); ok {
		return v.GetName()
	}
	return ""
}

func (param NamedParam) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	if param.Name == "" {
		return fmt.Errorf("Param name cannot be empty")
	}
	// TODO: what happens if you next Params? Param("a", Param("b", Param("c", 11))). Need to add a test for it.
	if v, ok := param.Value.(SQLExcludeAppender); ok && v != nil {
		return v.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
	}
	if v, ok := param.Value.(SQLAppender); ok && v != nil {
		return v.AppendSQL(dialect, buf, args, params, nil)
	}
	if isExplodableSlice(param.Value) {
		return explodeSlice(dialect, buf, args, params, excludedTableQualifiers, param.Value)
	}
	indices := params[param.Name]
	switch dialect {
	case DialectPostgres, DialectSQLite:
		if len(indices) > 0 {
			(*args)[indices[0]] = param.Value
			buf.WriteString("$" + strconv.Itoa(params[param.Name][0]+1))
		} else {
			params[param.Name] = append(indices, len(*args))
			buf.WriteString("$" + strconv.Itoa(len(*args)+1))
			*args = append(*args, param.Value)
		}
	case DialectSQLServer:
		if len(indices) > 0 {
			(*args)[indices[0]] = param.Value
			buf.WriteString("@p" + strconv.Itoa(indices[0]+1))
		} else {
			params[param.Name] = append(indices, len(*args))
			buf.WriteString("@p" + strconv.Itoa(len(*args)+1))
			*args = append(*args, param.Value)
		}
	default:
		params[param.Name] = append(indices, len(*args))
		buf.WriteString("?")
		*args = append(*args, param.Value)
	}
	return nil
}
