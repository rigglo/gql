package parser_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/rigglo/gql/utils/parser"
)

func TestParse(t *testing.T) {
	query := `
query {
	users {
		... basicInfo
	} }
`
	token, doc, err := parser.Parse(query)
	if err != nil {
		t.Errorf("error: %v, at Line: %v, Col: %v", err, token.Line, token.Col)
		return
	}
	spew.Dump(doc)

}
