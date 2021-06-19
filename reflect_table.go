package sq

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

func ReflectTable(table Table) error {
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
	genericTable, ok := v.Interface().(GenericTable)
	if !ok {
		return fmt.Errorf("first field not a embedded GenericTable")
	}
	if !vtype.Anonymous {
		return fmt.Errorf("first field not an embedded GenericTable")
	}
	if !v.CanSet() {
		return nil
	}
	if genericTable.TableSchema == "" || genericTable.TableName == "" {
		modifiers, modifierIndex, err := lexModifiers(vtype.Tag.Get("sq"))
		if err != nil {
			return err
		}
		if genericTable.TableSchema == "" {
			if i, ok := modifierIndex["schema"]; ok {
				genericTable.TableSchema = modifiers[i][1]
			}
		}
		if genericTable.TableName == "" {
			if i, ok := modifierIndex["name"]; ok {
				genericTable.TableSchema = modifiers[i][1]
			}
		}
	}
	if genericTable.TableName == "" {
		genericTable.TableName = strings.ToLower(typ.Name())
	}
	value.Field(0).Set(reflect.ValueOf(genericTable))
	for i := 1; i < value.NumField(); i++ {
		v := value.Field(i)
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
		case BlobField:
			v.Set(reflect.ValueOf(NewBlobField(fieldName, genericTable)))
		case BooleanField:
			v.Set(reflect.ValueOf(NewBooleanField(fieldName, genericTable)))
		case JSONField:
			v.Set(reflect.ValueOf(NewJSONField(fieldName, genericTable)))
		case NumberField:
			v.Set(reflect.ValueOf(NewNumberField(fieldName, genericTable)))
		case StringField:
			v.Set(reflect.ValueOf(NewStringField(fieldName, genericTable)))
		case TimeField:
			v.Set(reflect.ValueOf(NewTimeField(fieldName, genericTable)))
		case GenericField:
			v.Set(reflect.ValueOf(NewGenericField(fieldName, genericTable)))
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
