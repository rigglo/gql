package parser_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/rigglo/gql/utils/parser"
)

func TestParse(t *testing.T) {
	query := `
fragment basicInfo on User {
	id
	name
}

query myNamed($device: [Int]! [4, 3]) {
	users(foo:bar) {
		... basicInfo
	} 
}
`
	token, doc, err := parser.Parse(query)
	if err != nil {
		t.Errorf("error: %v, at Line: %v, Col: %v", err, token.Line, token.Col)
		return
	}
	spew.Dump(doc)

}
