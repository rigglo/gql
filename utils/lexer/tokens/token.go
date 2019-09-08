package tokens

import (
	"fmt"
)

type Token struct {
	Type  TokenType
	Value string
}

func (t Token) String() string {
	switch t.Type {
	case TokenEOF:
		return "EOF"

	case TokenError:
		return t.Value
	}

	return fmt.Sprintf("%q", t.Value)
}
