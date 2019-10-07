package lexer_test

import (
	"bufio"
	"log"
	"strings"
	"testing"

	"github.com/rigglo/gql/language/lexer"
)

func TestLexer(t *testing.T) {
	query := `
	query  {
		foo {
			bar
		}
	  }
`

	tokens := make(chan lexer.Token)
	src := strings.NewReader(query)
	readr := bufio.NewReader(src)
	go lexer.Lex(readr, tokens)
	for token := range tokens {
		log.Printf("token: %#v", token)
		//log.Print(token.Value)
		if token.Err != nil {
			t.Error(token.Err.Error())
		}
	}
}
