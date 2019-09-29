package lexer

import (
	"fmt"
	"strings"
)

type numLexer struct {
	negSign       string
	firstDigit    string
	integerDigits []string
	dot           string
	fractionals   []string
	expIndicator  string
	expSign       string
	expDigits     []string
	Err           error
}

func (nl *numLexer) Predicate(r rune) bool {
	if nl.firstDigit == "" && nl.negSign == "" && r == '-' {
		nl.negSign = string(r)
		return true
	} else if nl.firstDigit == "" && '0' <= r && r <= '9' {
		nl.firstDigit = string(r)
		return true
	} else if nl.firstDigit != "" && nl.firstDigit != "0" && '0' <= r && r <= '9' && nl.dot == "" {
		nl.integerDigits = append(nl.integerDigits, string(r))
		return true
	} else if nl.firstDigit != "" && nl.dot == "" && r == '.' {
		nl.dot = string(r)
		return true
	} else if nl.dot != "" && '0' <= r && r <= '9' {
		nl.fractionals = append(nl.fractionals, string(r))
		return true
	} else if nl.firstDigit != "" && strings.ContainsRune("eE", r) {
		nl.expIndicator = string(r)
		return true
	} else if nl.expIndicator != "" && len(nl.expDigits) == 0 && strings.ContainsRune("+-", r) {
		nl.expSign = string(r)
		return true
	} else if nl.expIndicator != "" && '0' <= r && r <= '9' {
		nl.expDigits = append(nl.expDigits, string(r))
		return true
	} else if !strings.ContainsRune(",)]} \n\r\t", r) {
		nl.Err = fmt.Errorf("invalid form of number: %s", nl.String()+string(r))
		return false
	}
	return false
}

func (nl *numLexer) Kind() TokenKind {
	if nl.dot != "" || nl.expIndicator != "" {
		return FloatValueToken
	}
	return IntValueToken
}

func (nl *numLexer) String() string {
	return nl.negSign + nl.firstDigit + strings.Join(nl.integerDigits, "") + nl.dot + strings.Join(nl.fractionals, "") + nl.expIndicator + nl.expSign + strings.Join(nl.expDigits, "")
}
