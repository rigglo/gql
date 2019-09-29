package lex_test

import (
	"bufio"
	"log"
	"strings"
	"testing"

	"github.com/rigglo/gql/utils/lex"
)

func TestLexer(t *testing.T) {
	query := `
	query withNestedFragments {
		user(id: 4) asd ") {
		  friends(first: 10) {
			...friendFields
		  }
		  mutualFriends(first: 10) {
			...friendFields
		  }
		}
	  }
	  
	  fragment friendFields on User {
		id
		name
		...standardProfilePic
	  }
	  
	  fragment standardProfilePic on User {
		profilePic(size: 50)
	  }	
`

	tokens := make(chan lex.Token)
	src := strings.NewReader(query)
	readr := bufio.NewReader(src)
	go lex.Lex(readr, tokens)
	for token := range tokens {
		log.Printf("token: %#v", token)
		//log.Print(token.Value)
		if token.Err != nil {
			t.Error(token.Err.Error())
		}
	}
}
