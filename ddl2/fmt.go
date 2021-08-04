package ddl2

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

func tokenizeValue(s string) (value string, modifiers [][2]string, modifierPositions map[string]int, err error) {
	value, remainder, err := popBraceToken(s)
	if err != nil {
		return "", nil, modifierPositions, err
	}
	modifiers, modifierPositions, err = tokenizeModifiers(remainder)
	if err != nil {
		return "", nil, modifierPositions, err
	}
	return value, modifiers, modifierPositions, nil
}

func tokenizeModifiers(s string) (modifiers [][2]string, modifierPositions map[string]int, err error) {
	modifierPositions = make(map[string]int)
	var n int
	token, remainder := "", s
	for remainder != "" {
		token, remainder, err = popBraceToken(remainder)
		if err != nil {
			return nil, modifierPositions, err
		}
		key, value := token, ""
		if j := strings.Index(token, "="); j >= 0 {
			key, value = token[:j], token[j+1:]
			if value[0] == '{' {
				value = value[1 : len(value)-1]
			}
		}
		modifiers = append(modifiers, [2]string{key, value})
		modifierPositions[key] = n
		n++
	}
	return modifiers, modifierPositions, nil
}

func popIdentifierToken(dialect, s string) (word, rest string, splitAt int) {
	s = strings.TrimLeft(s, " \t\n\v\f\r\u0085\u00A0")
	if s == "" {
		return "", "", -1
	}
	var openingQuote rune
	var insideIdentifier bool
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
	return s[:splitAt], s[splitAt:], splitAt
}

func popIdentifierTokens(dialect, s string, count int) (tokens []string, remainder string, splitAt int) {
	splitAt = -1
	token, remainder := "", s
	for i := 0; (count < 0 || i < count) && remainder != ""; i++ {
		token, remainder, splitAt = popIdentifierToken(dialect, remainder)
		if token == "" && remainder == "" {
			break
		}
		tokens = append(tokens, token)
	}
	return tokens, remainder, splitAt
}

func splitArgs(s string) []string {
	var args []string
	var splitAt, skipCharAt, arrayLevel, bracketLevel int
	var insideString bool
	for {
		splitAt, skipCharAt, arrayLevel, bracketLevel = -1, -1, 0, 0
		insideString = false
		for i, char := range s {
			// do we unconditionally skip the current char?
			if skipCharAt == i {
				continue
			}
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
				continue
			}
			// are we currently inside a bracket expression?
			if bracketLevel > 0 {
				switch char {
				// does the current char close a bracket expression?
				case ')':
					bracketLevel--
				// does the current char start a new bracket expression?
				case '(':
					bracketLevel++
				}
				continue
			}
			// are we currently inside a string?
			if insideString {
				nextIndex := i + 1
				// does the current char terminate the current string?
				if char == '\'' {
					// is the next char the same as the current char, which
					// escapes it and prevents it from terminating the current
					// string?
					if nextIndex < len(s) && s[nextIndex] == '\'' {
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
			// does the current char mark the start of a new bracket expression?
			if char == '(' {
				bracketLevel++
				continue
			}
			// does the current char mark the start of a new string?
			if char == '\'' {
				insideString = true
				continue
			}
			// is the current char an argument delimiter?
			if char == ',' {
				splitAt = i
				break
			}
		}
		// did we find an argument delimiter?
		if splitAt >= 0 {
			args, s = append(args, s[:splitAt]), s[splitAt+1:]
		} else {
			args = append(args, s)
			break
		}
	}
	return args
}
