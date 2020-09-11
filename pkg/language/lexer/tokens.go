package lexer

type Token struct {
	Kind TokenKind

	Value string

	Err        error
	Start, End int
	Line, Col  int
}

// TokenKind tells what kind of value is in the Token
type TokenKind int

const (
	// Undefined token
	Undefined TokenKind = iota
	// EOFToken -> EOF
	EOFToken
	// PunctuatorToken has special characters
	PunctuatorToken
	// NameToken has names
	NameToken // /[_A-Za-z][_0-9A-Za-z]*/
	// IntValueToken has iteger numbers
	IntValueToken // NegativeSign(opt) | NonZeroDigit | Digit (list, opt)
	// FloatValueToken has float numbers
	FloatValueToken // Sign (opt) | IntegerPart | FractionalPart (ExponentPart)
	// StringValueToken has string values
	StringValueToken // "something which is a string" | """this is also valid"""
)

var tokenNames = []string{
	"Undefined",
	"EOF",
	"Punctuator",
	"Name",
	"IntValue",
	"FloatValue",
	"StringValue",
}

func (tk TokenKind) String() string {
	return tokenNames[tk]
}
