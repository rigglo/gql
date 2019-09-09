package parser

import (
	"strings"

	"github.com/rigglo/gql/utils/lexer"
	"github.com/rigglo/gql/utils/lexer/tokens"
)

func isEOF(token tokens.Token) bool {
	return token.Type == tokens.TokenEOF
}

type (
	// ResolveFn ...
	ResolveFn func(src interface{}, args []*Argument) interface{}
	// Operation ...
	Operation struct {
		Op        string
		RootField Field
	}
	// Field ...
	Field struct {
		Name   string      `json:"name"`
		Alias  string      `json:"alias"`
		Fields []*Field    `json:"fields"`
		Args   []*Argument `json:"args"`
	}
	// Argument ...
	Argument struct {
		Key   string
		Value string
	}
)

func Parse(l *lexer.Lexer) Operation {
	var token tokens.Token
	var op Operation
	for {
		if l.IsEOF() {
			break
		}

		token = l.NextToken()

		if isEOF(token) {
			break
		}
		switch token.Type {
		case tokens.TokenQuery:
			op.Op = tokens.Query
			op.RootField.Name = "query"
			break
		case tokens.TokenSelectionSetStart:
			ParseFields(&op.RootField, nil, l)
			break
		case tokens.TokenSelectionSetEnd:
			break
		}
	}
	return op
}

func ParseFields(sf *Field, prevToken *tokens.Token, l *lexer.Lexer) {
	var token tokens.Token

	if prevToken != nil && prevToken.Type == tokens.TokenField {
		token = *prevToken
	} else {
		token = l.NextToken()
	}
	var f *Field

	for {
		if l.IsEOF() {
			break
		}

		switch token.Type {
		case tokens.TokenAlias:
			f = new(Field)
			f.Alias = token.Value
		case tokens.TokenField:
			if f == nil {
				f = new(Field)
			}
			f.Name = strings.TrimSpace(token.Value)
			if f.Alias == "" {
				f.Alias = f.Name
			}
			ParseField(l, f, sf)
			sf.Fields = append(sf.Fields, f)
			f = nil
			break
		case tokens.TokenSelectionSetEnd:
			return
		}

		if l.IsEOF() {
			return
		}

		token = l.NextToken()
	}
}

func ParseField(l *lexer.Lexer, f *Field, sf *Field) {
	var arg *Argument
	var token tokens.Token
	for {
		if l.IsEOF() {
			break
		}

		token = l.NextToken()

		if isEOF(token) {
			break
		}

		switch token.Type {
		case tokens.TokenFieldEnd:
			return
		case tokens.TokenFieldArgument:
			arg = new(Argument)
			arg.Key = token.Value
			break
		case tokens.TokenFieldArgumentValue:
			arg.Value = token.Value
			f.Args = append(f.Args, arg)
			arg = nil
			break
		case tokens.TokenSelectionSetStart:
			ParseFields(f, nil, l)
			return
		case tokens.TokenSelectionSetEnd:
			return
		}
	}
}
