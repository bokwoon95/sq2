package ddl

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bokwoon95/sq"
)

func popBraceToken(s string) (value, rest string, err error) {
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

func tokenizeValue(s string) (value string, modifiers [][2]string, modifierIndex map[string]int, err error) {
	value, rest, err := popBraceToken(s)
	if err != nil {
		return "", nil, modifierIndex, err
	}
	modifiers, modifierIndex, err = tokenizeModifiers(rest)
	if err != nil {
		return "", nil, modifierIndex, err
	}
	return value, modifiers, modifierIndex, nil
}

func tokenizeModifiers(s string) (modifiers [][2]string, modifierIndex map[string]int, err error) {
	modifierIndex = make(map[string]int)
	var currentIndex int
	value, rest := "", s
	for rest != "" {
		value, rest, err = popBraceToken(rest)
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

func popIdentifierToken(dialect, s string) (word, rest string) {
	s = strings.TrimLeft(s, " \t\n\v\f\r\u0085\u00A0")
	if s == "" {
		return "", ""
	}
	var openingQuote rune
	var insideIdentifier bool
	var splitAt int
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		splitAt = i
		if insideIdentifier {
			switch openingQuote {
			case '\'', '"', '`':
				if r == openingQuote {
					if i < len(s) && rune(s[i]) == openingQuote {
						i += 1
					} else {
						insideIdentifier = false
					}
				}
			case '[':
				if r == ']' {
					if i < len(s) && s[i] == ']' {
						i += 1
					} else {
						insideIdentifier = false
					}
				}
			}
			continue
		}
		if r == '"' || (r == '`' && dialect == sq.DialectMySQL) || (r == '[' && dialect == sq.DialectSQLServer) {
			insideIdentifier = true
			openingQuote = r
			continue
		}
		if unicode.IsSpace(r) {
			splitAt -= size
			break
		}
	}
	return s[:splitAt], s[splitAt:]
}

func popIdentifierTokens(dialect, s string, num int) (words []string, rest string) {
	word, rest := "", s
	for i := 0; i < num && rest != ""; i++ {
		word, rest = popIdentifierToken(dialect, rest)
		words = append(words, word)
	}
	return words, rest
}
