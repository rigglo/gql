package parser_test

import (
	"testing"

	"github.com/rigglo/gql/pkg/language/parser"
)

func TestParse(t *testing.T) {
	query := `
query {
	a
	b
	c
	d {
		da
		db
		...asd
	}
}
`
	token, _, err := parser.Parse([]byte(query))
	if err != nil {
		t.Errorf("error: %v, at Line: %v, Col: %v", err, token.Line, token.Col)
		return
	}
	// spew.Dump(doc)

}
