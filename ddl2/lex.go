package ddl2

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
