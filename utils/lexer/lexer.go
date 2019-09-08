package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/rigglo/gql/utils/lexer/tokens"
)

type Lexer struct {
	Name   string
	Input  string
	Tokens chan tokens.Token
	State  LexFn

	Start     int
	Pos       int
	Width     int
	LastToken *tokens.Token
}

// Emit puts a token onto the token channel.
func (l *Lexer) Emit(ttype tokens.TokenType) {
	l.Tokens <- tokens.Token{
		Type:  ttype,
		Value: l.Input[l.Start:l.Pos],
	}
	l.Start = l.Pos
}

// Inc increments the current position
func (l *Lexer) Inc() {
	l.Pos++
	if l.Pos >= utf8.RuneCountInString(l.Input) {
		l.Emit(tokens.TokenEOF)
	}
}

func (l *Lexer) Dec() {
	l.Pos--
}

func (l *Lexer) IsEOF() bool {
	return l.Pos >= len(l.Input)
}

// InputToEnd returns a slice of the input from the current pos to the end
func (l *Lexer) InputToEnd() string {
	return l.Input[l.Pos:]
}

func (l *Lexer) Next() rune {
	if l.Pos >= utf8.RuneCountInString(l.Input) {
		l.Width = 0
		return tokens.EOF
	}
	res, width := utf8.DecodeRuneInString(l.Input[l.Pos:])
	l.Width = width
	l.Pos += width
	return res
}

// NextToken ...
func (l *Lexer) NextToken() tokens.Token {
	for {
		select {
		case token := <-l.Tokens:
			l.LastToken = &token
			//log.Printf("NextToken: '%v', '%v'", token.Type, token.Value)
			return token
		default:
			l.State = l.State(l)
		}
	}
}

// SkipIgnoredTokens skips the ignorable tokens
// UnicodeBOM, WhiteSpace, LineTerminator, Comment, Comma
func (l *Lexer) SkipIgnoredTokens() {
	for {
		ch := l.Next()

		if !unicode.IsSpace(ch) {
			l.Dec()
			l.Start = l.Pos
			break
		}

		if ch == tokens.EOF {
			l.Emit(tokens.TokenEOF)
			break
		}
	}
}

func (l *Lexer) Errorf(format string, args ...interface{}) LexFn {
	l.Tokens <- tokens.Token{
		Type:  tokens.TokenError,
		Value: fmt.Sprintf(format, args...),
	}

	return nil
}
