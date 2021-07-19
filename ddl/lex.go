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
		if token == "" && remainder == "" {
			break
		}
		tokens = append(tokens, token)
	}
	return tokens, remainder
}

func qlevel(s string, quote byte, level int) int {
	for s != "" {
		i := strings.IndexByte(s, quote)
		if i < 0 {
			break
		}
		if level == 0 {
			level = 1
			s = s[i+1:]
			continue
		}
		j := i + 1
		if j < len(s) && s[j] == quote {
			s = s[j+1:]
			continue
		}
		s = s[i+1:]
		level = 0
	}
	return level
}

func splitArgs(s string) []string {
	var args []string
	tmp := s
	for tmp != "" {
		skipCharAt := -1
		splitAt := -1
		insideString := false
		arrayLevel := 0
		for i, char := range tmp {
			// do we unconditionally skip the current char?
			if skipCharAt == i {
				continue
			}
			nextIndex := i + 1
			// are we currently inside an array literal?
			if arrayLevel > 0 {
				switch char {
				// does the current char close an array literal?
				case ']':
					arrayLevel--
				// does the current char start a new array literal?
				case '[':
					arrayLevel++
				}
				// are we still inside an array literal?
				if arrayLevel > 0 {
					continue
				}
			}
			// are we currently inside a string?
			if insideString {
				// does the current char terminate the current string?
				if char == '\'' {
					// is the next char the same as the current char, which
					// escapes it and prevents it from terminating the current
					// string?
					if nextIndex < len(tmp) && tmp[nextIndex] == '\'' {
						skipCharAt = nextIndex
					} else {
						insideString = false
					}
				}
				continue
			}
			// does the current char mark the start of a new array literal?
			if char == '[' {
				arrayLevel++
				continue
			}
			// does the current char mark the start of a new string?
			if char == '\'' {
				insideString = true
				continue
			}
			// does the current char delimit arguments?
			if char == ',' {
				// are we currently inside an array literal or string? if yes,
				// the delimiter is part of the array literal or string and is
				// not used to delimit arguments
				if arrayLevel > 0 || insideString {
					continue
				}
				splitAt = i
				break
			}
		}
		// did we find an argument delimiter?
		if splitAt >= 0 {
			args, tmp = append(args, tmp[:splitAt]), tmp[splitAt+1:]
		} else {
			args = append(args, tmp)
			break
		}
	}
	return args
}
