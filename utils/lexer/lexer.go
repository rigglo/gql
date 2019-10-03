package lexer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// A Token is a single lexical element.
type Token struct {
	Kind  TokenKind
	Value string
	Err   error

	Line, Col int
}

// TokenKind tells what kind of value is in the Token
type TokenKind int

const (
	// BadToken is bad.. too bad..
	BadToken TokenKind = iota
	// PunctuatorToken has special characters
	PunctuatorToken // !	$	(	)	...	:	=	@	[	]	{	|	}
	// NameToken has names
	NameToken // /[_A-Za-z][_0-9A-Za-z]*/
	// IntValueToken has iteger numbers
	IntValueToken // NegativeSign(opt) | NonZeroDigit | Digit (list, opt)
	// FloatValueToken has float numbers
	FloatValueToken // Sign (opt) | IntegerPart | FractionalPart (ExponentPart)
	// StringValueToken has string values
	StringValueToken // "something which is a string"
	// UnicodeBOMToken is just the \ufeff
	UnicodeBOMToken // \ufeff
	// WhitespaceToken is \t and 'space'
	WhitespaceToken // \t and 'space'
	// LineTerminatorToken is \n and \r
	LineTerminatorToken // \n
	// CommentToken is # something like this
	CommentToken // # Just a comment....
	// CommaToken is just a ','
	CommaToken // ,
)

// lexFn is a lexer state function. Each lexFn lexes a token, sends it on the
// supplied channel, and returns the next lexFn to use.
type lexFn func(src *bufio.Reader, tokens chan<- Token, line, col int) (lexFn, int, int)

// Lex converts a source into a stream of tokens.
func Lex(src *bufio.Reader, tokens chan<- Token) {
	state := eatSpace
	line, col := 1, 1
	for state != nil {
		state, line, col = state(src, tokens, line, col)
	}
	close(tokens)
}

// accept appends the next run of characters in src which satisfy the predicate
// to b. Returns b after appending, the first rune which did not satisfy the
// predicate, and any error that occurred. If there was no such error, the
// last rune is unread.
func accept(src *bufio.Reader, predicate func(rune) bool, b []byte) ([]byte, rune, error) {
	r, _, err := src.ReadRune()
	for {
		if err != nil {
			return b, r, err
		}
		if !predicate(r) {
			break
		}
		b = append(b, string(r)...)
		r, _, err = src.ReadRune()
	}
	src.UnreadRune()
	return b, r, nil
}

// lexsend is a shortcut for sending a token with error checking. It returns
// eatSpace as the default lexing function.
func lexsend(err error, tokens chan<- Token, good Token) lexFn {
	if err != nil && err != io.EOF {
		good.Kind = BadToken
		good.Err = err
	}
	tokens <- good
	if err != nil {
		return nil
	}
	return eatSpace
}

// eatSpace consumes space and decides the next lexFn to use.
func eatSpace(src *bufio.Reader, tokens chan<- Token, line, col int) (lexFn, int, int) {
	eaten, r, err := accept(src, func(r rune) bool { return strings.ContainsRune(" \t", r) }, nil)
	col += len(eaten)
	if err != nil {
		if err != io.EOF {
			tokens <- Token{
				Kind:  BadToken,
				Value: string(r),
				Err:   err,
			}
		}
		return nil, line, col
	}
	switch {
	// Check for UnicodeBOM
	case '\ufeff' == r:
		src.ReadRune()
		tokens <- Token{
			Kind:  UnicodeBOMToken,
			Value: string(r),
			Line:  line,
			Col:   col,
		}
		line++
		col = 1
		return eatSpace, line, col
	// Check for LineTerminator
	case strings.ContainsRune("\n\r", r):
		src.ReadRune()
		/*tokens <- Token{
			Kind:  LineTerminatorToken,
			Value: string(r),
			Line:  line,
			Col:   col,
		}*/
		line++
		col = 1
		return eatSpace, line, col
	// Check for a Comment
	case '#' == r:
		return lexComment, line, col
	// Check for Comma
	case ',' == r:
		src.ReadRune()
		/*tokens <- Token{
			Kind:  CommaToken,
			Value: string(r),
			Line:  line,
			Col:   col,
		}*/
		col = 1
		return eatSpace, line, col
	// Checking for a Name
	case 'a' <= r && r <= 'z', 'A' <= r && r <= 'Z', r == '_':
		return lexName, line, col
	// Checking for Punctation
	case strings.ContainsRune("!$().:=@[]{|}", r):
		if r == '.' {
			return lexThreeDot(src, tokens, line, col)
		}
		src.ReadRune()
		tokens <- Token{
			Kind:  PunctuatorToken,
			Value: string(r),
			Line:  line,
			Col:   col,
		}
		line++
		col = 1
		return eatSpace, line, col
	case '0' <= r && r <= '9', r == '-':
		return lexNumber, line, col
	case r == '"':
		return lexString, line, col
	}
	tokens <- Token{
		Kind:  BadToken,
		Value: string(r),
		Err:   fmt.Errorf("lexer encountered invalid character %q", r),
		Line:  line,
		Col:   col,
	}
	return nil, line, col
}

// lexComment lexes a comment which starts with a # and ends with a LineTerminator
func lexComment(src *bufio.Reader, tokens chan<- Token, line, col int) (lexFn, int, int) {
	b, _, _ := accept(src, func(r rune) bool {
		return !strings.ContainsRune("\n\r", r)
	}, nil)
	// TODO: check for error
	ncol := col + len(b)
	// return lexsend(err, tokens, Token{Kind: CommentToken, Value: string(b), Line: line, Col: col}), line, ncol
	return eatSpace, line, ncol
}

// lexName lexes a name which is follows the /[_A-Za-z][_0-9A-Za-z]*/ form
func lexName(src *bufio.Reader, tokens chan<- Token, line, col int) (lexFn, int, int) {
	b, _, err := accept(src, func(r rune) bool {
		return 'a' <= r && r <= 'z' ||
			'A' <= r && r <= 'Z' ||
			'0' <= r && r <= '9' ||
			r == '_'
	}, nil)
	ncol := col + len(b)
	return lexsend(err, tokens, Token{Kind: NameToken, Value: string(b), Line: line, Col: col}), line, ncol
}

// lexComment lexes a comment which starts with a # and ends with a LineTerminator
func lexThreeDot(src *bufio.Reader, tokens chan<- Token, line, col int) (lexFn, int, int) {
	b, _, err := accept(src, func(r rune) bool { return r == '.' }, nil)
	ncol := col + len(b)
	if len(b) != 3 {
		tokens <- Token{
			Kind:  BadToken,
			Value: string(b),
			Err:   fmt.Errorf("invalid character '.'"),
			Line:  line,
			Col:   col,
		}
		return nil, line, ncol
	}
	return lexsend(err, tokens, Token{Kind: PunctuatorToken, Value: "...", Line: line, Col: col}), line, ncol
}

// lexNumber lexes a number, an integer or a float
func lexNumber(src *bufio.Reader, tokens chan<- Token, line, col int) (lexFn, int, int) {
	nl := new(numLexer)
	b, _, err := accept(src, nl.Predicate, nil)
	ncol := col + len(b)
	if err != nil {
		return lexsend(err, tokens, Token{Kind: BadToken, Value: string(b), Line: line, Col: col}), line, ncol
	} else if nl.Err != nil {
		return lexsend(nl.Err, tokens, Token{Kind: BadToken, Value: string(b), Line: line, Col: col}), line, ncol
	}

	return lexsend(err, tokens, Token{Kind: nl.Kind(), Value: string(b), Line: line, Col: col}), line, ncol
}

// lexString lexes single quoted and triple quoted strings
func lexString(src *bufio.Reader, tokens chan<- Token, line, col int) (lexFn, int, int) {
	peek, _ := src.Peek(3)
	if bytes.Equal(peek, []byte{'"', '"', '"'}) {
		return lexTriplequotedString(src, tokens, line, col)
	}
	return lexSinglequotedString(src, tokens, line, col)
}

// lexSinglequotedString lexes single quoted strings
func lexSinglequotedString(src *bufio.Reader, tokens chan<- Token, line, col int) (lexFn, int, int) {
	b := make([]byte, 1, 2)
	src.Read(b)
	ncol := col + 1
	ps := false
	for {
		r, _, err := src.ReadRune()
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			tokens <- Token{
				Kind:  BadToken,
				Value: string(b),
				Err:   err,
				Line:  line,
				Col:   col,
			}
			return nil, line, ncol
		}
		ncol++
		b = append(b, string(r)...)
		if r == '\\' {
			ps = !ps
		} else if r == '"' && !ps {
			return lexsend(err, tokens, Token{Kind: StringValueToken, Value: string(b), Line: line, Col: col}), line, ncol
		} else {
			ps = false
		}
	}
}

func lexTriplequotedString(src *bufio.Reader, tokens chan<- Token, line, col int) (lexFn, int, int) {
	b := make([]byte, 3, 6)
	src.Read(b)
	nline := line
	ncol := col + 3
	for {
		r, _, err := src.ReadRune()
		ncol++
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			tokens <- Token{
				Kind:  BadToken,
				Value: string(b),
				Err:   err,
				Line:  line,
				Col:   col,
			}
			return nil, line, ncol
		}
		if r == '\n' {
			nline++
			ncol = 1
		} else if r == '"' {
			peek, err := src.Peek(2)
			if bytes.Equal(peek, []byte{'"', '"'}) {
				src.Read([]byte{1: 0})
				ncol += 2
				return lexsend(err, tokens, Token{Kind: StringValueToken, Value: string(b) + `"""`, Line: line, Col: col}), nline, ncol
			}
		}
		b = append(b, string(r)...)
	}
}

// nummLexer helps to lex a number and decide if it's an integer or a float
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

// Predicate helps decide what to read from the source
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

// Kind returns if the final number is Float or Int
func (nl *numLexer) Kind() TokenKind {
	if nl.dot != "" || nl.expIndicator != "" {
		return FloatValueToken
	}
	return IntValueToken
}

// String just returns the final number
func (nl *numLexer) String() string {
	return nl.negSign + nl.firstDigit + strings.Join(nl.integerDigits, "") + nl.dot + strings.Join(nl.fractionals, "") + nl.expIndicator + nl.expSign + strings.Join(nl.expDigits, "")
}
