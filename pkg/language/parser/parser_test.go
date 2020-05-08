package parser_test

import (
	"log"
	"testing"

	"github.com/rigglo/gql/pkg/language/ast"

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

func TestParseScalar(t *testing.T) {
	query := `
	"""some scalar"""
	scalar foo @depricated(reason: "just")

	scalar bar`
	token, doc, err := parser.Parse([]byte(query))
	if err != nil {
		t.Errorf("error: %v, at Line: %v, Col: %v", err, token.Line, token.Col)
		return
	}
	log.Printf("%#v", doc.Definitions[0])
	log.Printf("%#v", doc.Definitions[1])
	// spew.Dump(doc)

}

func TestParseObject(t *testing.T) {
	query := `
	type Person {
		name(
			"some example arg"
			bar: String

			"some other arg"
			foo: Int
		): String
		age: Int
		picture: Url
	}
	`
	token, doc, err := parser.Parse([]byte(query))
	if err != nil {
		t.Errorf("error: %v, at Line: %v, Col: %v", err, token.Line, token.Col)
		return
	}
	log.Printf("%#v", doc.Definitions[0].(*ast.ObjectDefinition))
	//log.Printf("%#v", doc.Definitions[1])
	// spew.Dump(doc)

}

func TestParseInterface(t *testing.T) {
	query := `
	interface Person {
		name(
			"some example arg"
			bar: String

			"some other arg"
			foo: Int
		): String
		age: Int
		picture: Url
	}
	`
	def, err := parser.ParseDefinition([]byte(query))
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}
	log.Printf("%#v", def.(*ast.InterfaceDefinition))
	//log.Printf("%#v", doc.Definitions[1])
	// spew.Dump(doc)

}

func TestParseUnion(t *testing.T) {
	query := `
	union Entity @asd = Person | Company | Animal
	`
	def, err := parser.ParseDefinition([]byte(query))
	if err != nil {
		t.Errorf("error: %v", err)
		return
	}
	log.Printf("%#v", def.(*ast.UnionDefinition))
	//log.Printf("%#v", doc.Definitions[1])
	// spew.Dump(doc)

}
