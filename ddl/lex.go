package ddl

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

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

func lexModifiers(config string) (modifiers [][2]string, nameIndex, colsIndex int, err error) {
	nameIndex, colsIndex = -1, -1
	i := 0
	value, rest := "", config
	for rest != "" {
		value, rest, err = cutValue(rest)
		if err != nil {
			return nil, nameIndex, colsIndex, err
		}
		subname, subvalue := value, ""
		if i := strings.Index(value, "="); i >= 0 {
			subname, subvalue = value[:i], value[i+1:]
			if subvalue[0] == '{' {
				subvalue = subvalue[1 : len(subvalue)-1]
			}
		}
		switch subname {
		case "name":
			nameIndex = i
		case "cols":
			colsIndex = i
		}
		modifiers = append(modifiers, [2]string{subname, subvalue})
		i++
	}
	return modifiers, nameIndex, colsIndex, nil
}

func lexValue(config string) (value string, modifiers [][2]string, nameIndex, colsIndex int, err error) {
	value, rest, err := cutValue(config)
	if err != nil {
		return "", nil, -1, -1, err
	}
	modifiers, nameIndex, colsIndex, err = lexModifiers(rest)
	if err != nil {
		return "", nil, nameIndex, colsIndex, err
	}
	return value, modifiers, nameIndex, colsIndex, err
}
