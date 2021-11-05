package sq

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

func ReflectTable(table Table, alias string) error {
	ptrvalue := reflect.ValueOf(table)
	typ := ptrvalue.Type()
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("not a pointer")
	}
	value := reflect.Indirect(ptrvalue)
	typ = value.Type()
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct pointer")
	}
	if value.NumField() == 0 {
		return fmt.Errorf("empty struct")
	}
	v := value.Field(0)
	vtype := typ.Field(0)
	tableInfo, ok := v.Interface().(TableInfo)
	if !ok {
		return fmt.Errorf("first field not a embedded TableInfo")
	}
	if !vtype.Anonymous {
		return fmt.Errorf("first field not an embedded TableInfo")
	}
	if !v.CanSet() {
		return nil
	}
	if tableInfo.TableSchema == "" || tableInfo.TableName == "" {
		modifiers, modifierIndex, err := lexModifiers(vtype.Tag.Get("sq"))
		if err != nil {
			return err
		}
		if tableInfo.TableSchema == "" {
			if i, ok := modifierIndex["schema"]; ok {
				tableInfo.TableSchema = modifiers[i][1]
			}
		}
		if tableInfo.TableName == "" {
			if i, ok := modifierIndex["name"]; ok {
				tableInfo.TableName = modifiers[i][1]
			}
		}
	}
	if tableInfo.TableName == "" {
		tableInfo.TableName = strings.ToLower(typ.Name())
	}
	tableInfo.TableAlias = alias
	value.Field(0).Set(reflect.ValueOf(tableInfo))
	for i := 1; i < value.NumField(); i++ {
		v := value.Field(i)
		if !v.CanInterface() {
			continue
		}
		fieldValue, ok := v.Interface().(Field)
		if !ok {
			continue
		}
		if fieldValue.GetName() != "" {
			continue
		}
		if !v.CanSet() {
			continue
		}
		fieldType := typ.Field(i)
		var fieldName string
		modifiers, modifierIndex, err := lexModifiers(fieldType.Tag.Get("sq"))
		if err != nil {
			return err
		}
		if i, ok := modifierIndex["name"]; ok {
			fieldName = modifiers[i][1]
		}
		if fieldName == "" {
			fieldName = strings.ToLower(fieldType.Name)
		}
		switch fieldValue.(type) {
		case BinaryField:
			v.Set(reflect.ValueOf(NewBlobField(fieldName, tableInfo)))
		case BooleanField:
			v.Set(reflect.ValueOf(NewBooleanField(fieldName, tableInfo)))
		case CustomField:
			v.Set(reflect.ValueOf(NewCustomField(fieldName, tableInfo)))
		case JSONField:
			v.Set(reflect.ValueOf(NewJSONField(fieldName, tableInfo)))
		case NumberField:
			v.Set(reflect.ValueOf(NewNumberField(fieldName, tableInfo)))
		case StringField:
			v.Set(reflect.ValueOf(NewStringField(fieldName, tableInfo)))
		case TimeField:
			v.Set(reflect.ValueOf(NewTimeField(fieldName, tableInfo)))
		case UUIDField:
			v.Set(reflect.ValueOf(NewUUIDField(fieldName, tableInfo)))
		}
	}
	return nil
}

func cutValue(s string) (value, rest string, err error) {
	s = strings.TrimLeft(s, " \t\n\v\f\r\u0085\u00A0")
	if s == "" {
		return "", "", nil
	}
	var bracelevel, splitAt int
	isBraceQuoted := s[0] == '{'
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		splitAt = i
		switch r {
		case '{':
			bracelevel++
		case '}':
			bracelevel--
		}
		if bracelevel < 0 {
			return "", "", fmt.Errorf("too many closing braces")
		}
		if bracelevel == 0 && isBraceQuoted {
			break
		}
		if bracelevel == 0 && unicode.IsSpace(r) {
			splitAt -= size
			break
		}
	}
	if bracelevel > 0 {
		return "", "", fmt.Errorf("unclosed brace")
	}
	value = s[:splitAt]
	rest = s[splitAt:]
	if isBraceQuoted {
		value = value[1 : len(value)-1]
	}
	return value, rest, nil
}

func lexValue(s string) (value string, modifiers [][2]string, modifierIndex map[string]int, err error) {
	value, rest, err := cutValue(s)
	if err != nil {
		return "", nil, modifierIndex, err
	}
	modifiers, modifierIndex, err = lexModifiers(rest)
	if err != nil {
		return "", nil, modifierIndex, err
	}
	return value, modifiers, modifierIndex, err
}

func lexModifiers(s string) (modifiers [][2]string, modifierIndex map[string]int, err error) {
	modifierIndex = make(map[string]int)
	var currentIndex int
	value, rest := "", s
	for rest != "" {
		value, rest, err = cutValue(rest)
		if err != nil {
			return nil, modifierIndex, err
		}
		subname, subvalue := value, ""
		if j := strings.Index(value, "="); j >= 0 {
			subname, subvalue = value[:j], value[j+1:]
			if subvalue[0] == '{' {
				subvalue = subvalue[1 : len(subvalue)-1]
			}
		}
		modifierIndex[subname] = currentIndex
		modifiers = append(modifiers, [2]string{subname, subvalue})
		currentIndex++
	}
	return modifiers, modifierIndex, nil
}
