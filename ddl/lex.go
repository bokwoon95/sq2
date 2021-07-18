package ddl

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bokwoon95/sq"
)

func popBraceToken(s string) (token, remainder string, err error) {
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
			// TODO: do we actually ever end up here? can we just break from loop instead?
			return "", "", fmt.Errorf("popBraceToken: too many closing braces")
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
		// TODO: do we actually ever end up here? can we just return instead?
		return "", "", fmt.Errorf("popBraceToken: unclosed brace")
	}
	token, remainder = s[:splitAt], s[splitAt:]
	if isBraceQuoted {
		token = token[1 : len(token)-1]
	}
	return token, remainder, nil
}

func tokenizeValue(s string) (value string, modifiers [][2]string, modifierIndex map[string]int, err error) {
	value, remainder, err := popBraceToken(s)
	if err != nil {
		return "", nil, modifierIndex, err
	}
	modifiers, modifierIndex, err = tokenizeModifiers(remainder)
	if err != nil {
		return "", nil, modifierIndex, err
	}
	return value, modifiers, modifierIndex, nil
}

func tokenizeModifiers(s string) (modifiers [][2]string, modifierIndex map[string]int, err error) {
	modifierIndex = make(map[string]int)
	var i int
	token, remainder := "", s
	for remainder != "" {
		token, remainder, err = popBraceToken(remainder)
		if err != nil {
			return nil, modifierIndex, err
		}
		key, value := token, ""
		if j := strings.Index(token, "="); j >= 0 {
			key, value = token[:j], token[j+1:]
			if value[0] == '{' {
				value = value[1 : len(value)-1]
			}
		}
		modifiers = append(modifiers, [2]string{key, value})
		modifierIndex[key] = i
		i++
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

func popIdentifierTokens(dialect, s string, num int) (tokens []string, remainder string) {
	token, remainder := "", s
	for i := 0; i < num && remainder != ""; i++ {
		token, remainder = popIdentifierToken(dialect, remainder)
		tokens = append(tokens, token)
	}
	return tokens, remainder
}
