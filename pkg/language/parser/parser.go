package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/rigglo/gql/pkg/language/ast"
	"github.com/rigglo/gql/pkg/language/lexer"
)

type ParserError struct {
	Message string
	Line    int
	Column  int
	Token   lexer.Token
}

func (e *ParserError) Error() string {
	return e.Message
}

// Parse parses a gql document
func Parse(document []byte) (*ast.Document, error) {
	tokens := make(chan lexer.Token)
	src := bytes.NewReader(document)
	readr := bufio.NewReader(src)
	go lexer.Lex(readr, tokens)
	t, doc, err := parseDocument(tokens)
	if err != nil && t.Value == "" {
		return nil, &ParserError{
			Message: err.Error(),
			Line:    t.Line,
			Column:  t.Col,
			Token:   t,
		}
	}
	return doc, nil
}

// ParseDefinition parses a single schema, type, directive definition
func ParseDefinition(definition []byte) (ast.Definition, error) {
	tokens := make(chan lexer.Token)
	src := bytes.NewReader(definition)
	readr := bufio.NewReader(src)
	go lexer.Lex(readr, tokens)
	token := <-tokens

	desc := ""
	if token.Kind == lexer.StringValueToken {
		desc = strings.Trim(token.Value, `"`)
		token = <-tokens
		if token.Kind == lexer.NameToken && token.Value == "schema" {
			return nil, &ParserError{
				Message: "expected everything but 'schema'.. a Schema does NOT have description",
				Line:    token.Line,
				Column:  token.Col,
				Token:   token,
			}
		}
	}
	if token.Kind == lexer.NameToken && token.Value == "schema" {
		t, def, err := parseSchema(<-tokens, tokens)
		if err != nil {
			return nil, err
		} else if t.Kind != lexer.BadToken && err == nil {
			return nil, &ParserError{
				Message: fmt.Sprintf("invalid token after schema definition: '%s'", t.Value),
				Line:    token.Line,
				Column:  token.Col,
				Token:   token,
			}
		}
		return def, nil
	}
	t, def, err := parseDefinition(token, tokens, desc)
	if err != nil {
		return nil, &ParserError{
			Message: err.Error(),
			Line:    token.Line,
			Column:  token.Col,
			Token:   token,
		}
	} else if t.Kind != lexer.BadToken && err == nil {
		return nil, &ParserError{
			Message: fmt.Sprintf("invalid token after definition: '%s'", t.Value),
			Line:    token.Line,
			Column:  token.Col,
			Token:   token,
		}
	}
	return def, nil
}

func parseDocument(tokens chan lexer.Token) (lexer.Token, *ast.Document, error) {
	doc := ast.NewDocument()
	var err error
	token := <-tokens
	for {
		switch {
		case token.Kind == lexer.NameToken:
			switch token.Value {
			case "fragment":
				{
					f := new(ast.Fragment)
					token, f, err = parseFragment(tokens)
					if err != nil {
						return token, nil, err
					}
					doc.Fragments = append(doc.Fragments, f)
				}
			case "query", "mutation", "subscription":
				{
					op := new(ast.Operation)
					token, op, err = parseOperation(token, tokens)
					if err != nil {
						return token, nil, err
					}
					doc.Operations = append(doc.Operations, op)
				}
			case "scalar", "type", "interface", "union", "enum", "input", "directive":
				var def ast.Definition
				token, def, err = parseDefinition(token, tokens, "")
				if err != nil {
					return token, nil, err
				}
				doc.Definitions = append(doc.Definitions, def)
			case "schema":
				var def ast.Definition
				token, def, err = parseSchema(token, tokens)
				if err != nil {
					return token, nil, err
				}
				doc.Definitions = append(doc.Definitions, def)
			default:
				return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
			}
			break
		case token.Kind == lexer.StringValueToken:
			desc := strings.Trim(token.Value, `"`)
			token = <-tokens
			switch token.Value {
			case "scalar", "type", "interface", "union", "enum", "input", "directive":
				var def ast.Definition
				token, def, err = parseDefinition(token, tokens, desc)
				if err != nil {
					return token, nil, err
				}
				doc.Definitions = append(doc.Definitions, def)
			default:
				return token, nil, fmt.Errorf("unexpected token: '%s', kind: %v", token.Value, token.Kind)
			}
			break
		case token.Kind == lexer.PunctuatorToken && token.Value == "{":
			set := []ast.Selection{}
			token, set, err = parseSelectionSet(tokens)
			if err != nil {
				return token, nil, err
			}
			doc.Operations = append(doc.Operations, &ast.Operation{
				OperationType: ast.Query,
				SelectionSet:  set,
			})
			break
		case token.Kind == lexer.BadToken && token.Err == nil:
			return token, doc, nil
		default:
			return token, nil, fmt.Errorf("unexpected token: %s, kind: %v", token.Value, token.Kind)
		}
	}
}

func parseDefinition(token lexer.Token, tokens chan lexer.Token, desc string) (lexer.Token, ast.Definition, error) {
	switch token.Value {
	case "scalar":
		return parseScalar(<-tokens, tokens, desc)
	case "type":
		return parseObject(<-tokens, tokens, desc)
	case "interface":
		return parseInterface(<-tokens, tokens, desc)
	case "union":
		return parseUnion(<-tokens, tokens, desc)
	case "enum":
		return parseEnum(<-tokens, tokens, desc)
	case "input":
		return parseInputObject(<-tokens, tokens, desc)
	case "directive":
		return parseDirectiveDefinition(<-tokens, tokens, desc)
	}
	return token, nil, fmt.Errorf("expected a schema, type or directive definition, got: '%s', err: '%v'", token.Value, token.Err)
}

func parseSchema(token lexer.Token, tokens chan lexer.Token) (lexer.Token, ast.Definition, error) {
	def := new(ast.SchemaDefinition)

	// parse Name
	if token.Kind != lexer.NameToken {
		log.Println(token.Kind)
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = <-tokens

	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		var (
			ds  []*ast.Directive
			err error
		)
		token, ds, err = parseDirectives(tokens)
		if err != nil {
			return token, nil, err
		}
		def.Directives = ds
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		def.RootOperations = map[ast.OperationType]*ast.NamedType{}

		token = <-tokens
		for {
			// quit if it's the end of the field definition list
			if token.Kind == lexer.PunctuatorToken && token.Value == "}" {
				return <-tokens, def, nil
			}

			var ot ast.OperationType

			// parse operation type
			if token.Kind == lexer.NameToken {
				switch token.Value {
				case "query":
					ot = ast.Query
				case "mutation":
					ot = ast.Mutation
				case "subscription":
					ot = ast.Subscription
				default:
					return token, nil, fmt.Errorf("expected root operation type, one of 'query', 'mutation' or 'subscription', got '%s'", token.Value)
				}
				if _, ok := def.RootOperations[ot]; ok {
					return token, nil, fmt.Errorf("the given operation type '%s' is already defined in the schema definition", token.Value)
				}
				token = <-tokens
			} else {
				return token, nil, fmt.Errorf("expected NameToken, got '%s'", token.Value)
			}

			if token.Kind == lexer.PunctuatorToken && token.Value == ":" {
				token = <-tokens
			} else {
				return token, nil, fmt.Errorf("expected token ':', got '%s'", token.Value)
			}

			if token.Kind == lexer.NameToken {
				def.RootOperations[ot] = &ast.NamedType{
					Name: token.Value,
					Location: ast.Location{
						Column: token.Col,
						Line:   token.Line,
					},
				}
				token = <-tokens
			} else {
				return token, nil, fmt.Errorf("expected NameToken, got '%s'", token.Value)
			}
		}
	}
	return token, nil, fmt.Errorf("expected '{', and a list of root operation types, got '%s'", token.Value)
}

func parseScalar(token lexer.Token, tokens chan lexer.Token, desc string) (lexer.Token, ast.Definition, error) {
	def := &ast.ScalarDefinition{
		Description: desc,
	}

	// parse Name
	if token.Kind != lexer.NameToken {
		log.Println(token.Kind)
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = <-tokens

	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		token, ds, err := parseDirectives(tokens)
		if err != nil {
			return token, nil, err
		}
		def.Directives = ds
		return token, def, nil
	}

	return token, def, nil
}

func parseObject(token lexer.Token, tokens chan lexer.Token, desc string) (lexer.Token, ast.Definition, error) {
	def := &ast.ObjectDefinition{
		Description: desc,
		Fields:      []*ast.FieldDefinition{},
	}

	// parse Name
	if token.Kind != lexer.NameToken {
		log.Println(token.Kind)
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = <-tokens

	if token.Kind == lexer.NameToken && token.Value == "implements" {
		token = <-tokens
		ints := []*ast.NamedType{}
		for {
			if token.Kind == lexer.NameToken {
				ints = append(ints, &ast.NamedType{
					Name: token.Value,
					Location: ast.Location{
						Column: token.Col,
						Line:   token.Line,
					},
				})
				token = <-tokens
			} else if token.Kind == lexer.PunctuatorToken && token.Value == "&" {
				token = <-tokens
			} else {
				break
			}
		}
		if len(ints) == 0 {
			return token, nil, fmt.Errorf("language error..")
		}
		def.Implements = ints
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		var (
			ds  []*ast.Directive
			err error
		)
		token, ds, err = parseDirectives(tokens)
		if err != nil {
			return token, nil, fmt.Errorf("couldn't parse directives: %v", err)
		}
		def.Directives = ds
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		token = <-tokens
		for {
			// quit if it's the end of the field definition list
			if token.Kind == lexer.PunctuatorToken && token.Value == "}" {
				return <-tokens, def, nil
			}
			field := &ast.FieldDefinition{}

			// parse optional description
			if token.Kind == lexer.StringValueToken {
				field.Description = strings.Trim(token.Value, `"`)
				token = <-tokens
			}
			if token.Kind == lexer.NameToken {
				field.Name = token.Value
				token = <-tokens
			} else {
				return token, nil, fmt.Errorf("expected NameToken, got token kind '%v', with value '%s'", token.Kind, token.Value)
			}

			// parse arguments definition
			if token.Kind == lexer.PunctuatorToken && token.Value == "(" {
				token = <-tokens
				field.Arguments = []*ast.InputValueDefinition{}
				for {
					if token.Kind == lexer.PunctuatorToken && token.Value == ")" {
						token = <-tokens
						break
					}
					var (
						inputDef *ast.InputValueDefinition
						err      error
					)
					token, inputDef, err = parseInputValueDefinition(token, tokens)
					if err != nil {
						return token, nil, err
					}
					field.Arguments = append(field.Arguments, inputDef)
				}
			}

			if token.Kind == lexer.PunctuatorToken && token.Value == ":" {
				token = <-tokens
			} else {
				return token, nil, fmt.Errorf("expected ':', got token kind '%v', with value '%s'", token.Kind, token.Value)
			}

			// parse the type of the field
			var (
				t   ast.Type
				err error
			)
			token, t, err = parseType(token, tokens)
			if err != nil {
				return token, nil, err
			}
			field.Type = t

			// parse directives for field
			if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
				var (
					ds  []*ast.Directive
					err error
				)
				token, ds, err = parseDirectives(tokens)
				if err != nil {
					return token, nil, err
				}
				field.Directives = ds
			}

			def.Fields = append(def.Fields, field)
		}
	}

	return token, def, nil
}

func parseInterface(token lexer.Token, tokens chan lexer.Token, desc string) (lexer.Token, ast.Definition, error) {
	def := &ast.InterfaceDefinition{
		Description: desc,
		Fields:      []*ast.FieldDefinition{},
	}

	// parse Name
	if token.Kind != lexer.NameToken {
		log.Println(token.Kind)
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = <-tokens

	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		var (
			ds  []*ast.Directive
			err error
		)
		token, ds, err = parseDirectives(tokens)
		if err != nil {
			return token, nil, fmt.Errorf("couldn't parse directives: %v", err)
		}
		def.Directives = ds
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		token = <-tokens
		for {
			// quit if it's the end of the field definition list
			if token.Kind == lexer.PunctuatorToken && token.Value == "}" {
				return <-tokens, def, nil
			}
			field := &ast.FieldDefinition{}

			// parse optional description
			if token.Kind == lexer.StringValueToken {
				field.Description = strings.Trim(token.Value, `"`)
				token = <-tokens
			}
			if token.Kind == lexer.NameToken {
				field.Name = token.Value
				token = <-tokens
			} else {
				return token, nil, fmt.Errorf("expected NameToken, got token kind '%v', with value '%s'", token.Kind, token.Value)
			}

			// parse arguments definition
			if token.Kind == lexer.PunctuatorToken && token.Value == "(" {
				token = <-tokens
				field.Arguments = []*ast.InputValueDefinition{}
				for {
					if token.Kind == lexer.PunctuatorToken && token.Value == ")" {
						token = <-tokens
						break
					}
					var (
						inputDef *ast.InputValueDefinition
						err      error
					)
					token, inputDef, err = parseInputValueDefinition(token, tokens)
					if err != nil {
						return token, nil, err
					}
					field.Arguments = append(field.Arguments, inputDef)
				}
			}

			if token.Kind == lexer.PunctuatorToken && token.Value == ":" {
				token = <-tokens
			} else {
				return token, nil, fmt.Errorf("expected ':', got token kind '%v', with value '%s'", token.Kind, token.Value)
			}

			// parse the type of the field
			var (
				t   ast.Type
				err error
			)
			token, t, err = parseType(token, tokens)
			if err != nil {
				return token, nil, err
			}
			field.Type = t

			// parse directives for field
			if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
				var (
					ds  []*ast.Directive
					err error
				)
				token, ds, err = parseDirectives(tokens)
				if err != nil {
					return token, nil, err
				}
				field.Directives = ds
			}

			def.Fields = append(def.Fields, field)
		}
	}

	return token, def, nil
}

func parseUnion(token lexer.Token, tokens chan lexer.Token, desc string) (lexer.Token, ast.Definition, error) {
	def := &ast.UnionDefinition{
		Description: desc,
	}

	// parse Name
	if token.Kind != lexer.NameToken {
		log.Println(token.Kind)
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = <-tokens

	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		var (
			ds  []*ast.Directive
			err error
		)
		token, ds, err = parseDirectives(tokens)
		if err != nil {
			return token, nil, fmt.Errorf("couldn't parse directives: %v", err)
		}
		def.Directives = ds
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "=" {
		def.Members = []*ast.NamedType{}
		token = <-tokens
		if token.Kind == lexer.PunctuatorToken && token.Value == "|" {
			token = <-tokens
		}
		if token.Kind == lexer.NameToken {
			def.Members = append(def.Members, &ast.NamedType{
				Location: ast.Location{
					Column: token.Col,
					Line:   token.Line,
				},
				Name: token.Value,
			})
			token = <-tokens
		} else {
			return token, nil, fmt.Errorf("expected Name token, got '%s'", token.Value)
		}
		for {
			if token.Kind == lexer.PunctuatorToken && token.Value == "|" {
				token = <-tokens
			} else {
				return token, def, nil
			}
			if token.Kind == lexer.NameToken {
				def.Members = append(def.Members, &ast.NamedType{
					Location: ast.Location{
						Column: token.Col,
						Line:   token.Line,
					},
					Name: token.Value,
				})
				token = <-tokens
			} else {
				return token, nil, fmt.Errorf("expected Name token, got '%s'", token.Value)
			}
		}
	}
	return token, def, nil
}

func parseEnum(token lexer.Token, tokens chan lexer.Token, desc string) (lexer.Token, ast.Definition, error) {
	def := &ast.EnumDefinition{
		Description: desc,
	}

	// parse Name
	if token.Kind != lexer.NameToken {
		log.Println(token.Kind)
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = <-tokens

	// parse directives for the enum type
	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		var (
			ds  []*ast.Directive
			err error
		)
		token, ds, err = parseDirectives(tokens)
		if err != nil {
			return token, nil, fmt.Errorf("couldn't parse directives: %v", err)
		}
		def.Directives = ds
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		def.Values = []*ast.EnumValueDefinition{}
		token = <-tokens

		// parse all the enum value definitions
		for {
			if token.Kind == lexer.PunctuatorToken && token.Value == "}" {
				return <-tokens, def, nil
			}

			enumV := &ast.EnumValueDefinition{}
			if token.Kind == lexer.StringValueToken {
				enumV.Description = strings.Trim(token.Value, `"`)
				token = <-tokens
			}
			if token.Kind == lexer.NameToken {
				enumV.Value = &ast.EnumValue{
					Location: ast.Location{
						Column: token.Col,
						Line:   token.Line,
					},
					Value: token.Value,
				}
				token = <-tokens
			} else {
				return token, nil, fmt.Errorf("expected Name token, got '%v'", token.Value)
			}
			if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
				var (
					ds  []*ast.Directive
					err error
				)
				token, ds, err = parseDirectives(tokens)
				if err != nil {
					return token, nil, fmt.Errorf("couldn't parse directives on enum value: %v", err)
				}
				enumV.Directives = ds
			}
			def.Values = append(def.Values, enumV)
		}
	}
	return token, def, nil
}

func parseDirectiveDefinition(token lexer.Token, tokens chan lexer.Token, desc string) (lexer.Token, ast.Definition, error) {
	def := &ast.DirectiveDefinition{
		Description: desc,
	}

	// parse Name
	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		token = <-tokens
	} else {
		return token, nil, fmt.Errorf("expected token '@', got: '%s'", token.Value)
	}
	if token.Kind != lexer.NameToken {
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = <-tokens

	if token.Kind == lexer.NameToken && token.Value == "on" {
		def.Locations = []string{}
		token = <-tokens

		if token.Kind == lexer.PunctuatorToken && token.Value == "|" {
			token = <-tokens
		}
		if token.Kind == lexer.NameToken {
			def.Locations = append(def.Locations, token.Value)
			token = <-tokens
		} else {
			return token, nil, fmt.Errorf("expected Name token, got '%s'", token.Value)
		}
		for {
			if token.Kind == lexer.PunctuatorToken && token.Value == "|" {
				token = <-tokens
			} else {
				return token, def, nil
			}
			if token.Kind == lexer.NameToken && ast.IsValidDirective(token.Value) {
				def.Locations = append(def.Locations, token.Value)
				token = <-tokens
			} else {
				return token, nil, fmt.Errorf("expected Name token with a valid directive locaton, got '%s'", token.Value)
			}
		}
	}
	return token, def, nil
}

func parseInputObject(token lexer.Token, tokens chan lexer.Token, desc string) (lexer.Token, ast.Definition, error) {
	def := &ast.InputObjectDefinition{
		Description: desc,
	}

	// parse Name
	if token.Kind != lexer.NameToken {
		log.Println(token.Kind)
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = <-tokens

	// parse directives for the enum type
	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		var (
			ds  []*ast.Directive
			err error
		)
		token, ds, err = parseDirectives(tokens)
		if err != nil {
			return token, nil, fmt.Errorf("couldn't parse directives: %v", err)
		}
		def.Directives = ds
	}

	// parse input object field definitions
	if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		token = <-tokens
		def.Fields = []*ast.InputValueDefinition{}
		for {
			if token.Kind == lexer.PunctuatorToken && token.Value == "}" {
				token = <-tokens
				break
			}
			var (
				inputDef *ast.InputValueDefinition
				err      error
			)
			token, inputDef, err = parseInputValueDefinition(token, tokens)
			if err != nil {
				return token, nil, err
			}
			def.Fields = append(def.Fields, inputDef)
		}
	}
	return token, def, nil
}

func parseInputValueDefinition(token lexer.Token, tokens chan lexer.Token) (lexer.Token, *ast.InputValueDefinition, error) {
	val := &ast.InputValueDefinition{}

	// parse description for input
	if token.Kind == lexer.StringValueToken {
		val.Description = strings.Trim(token.Value, `"`)
		token = <-tokens
	}

	// parse name of the input
	if token.Kind == lexer.NameToken {
		val.Name = token.Value
		token = <-tokens
	} else {
		return token, nil, fmt.Errorf("expected NameToken, got token kind '%v', with value '%s'", token.Kind, token.Value)
	}

	// parse type for the input
	if token.Kind == lexer.PunctuatorToken && token.Value == ":" {
		token = <-tokens
		var (
			t   ast.Type
			err error
		)
		token, t, err = parseType(token, tokens)
		if err != nil {
			return token, nil, err
		}
		val.Type = t
	} else {
		// raise error since ':' and type are required in SDL
		return token, nil, fmt.Errorf("expected ':', got token kind '%v', with value '%s'", token.Kind, token.Value)
	}

	// parse default value ()
	if token.Kind == lexer.PunctuatorToken && token.Value == "=" {
		token = <-tokens
		var (
			v   ast.Value
			err error
		)
		token, v, err = parseValue(token, tokens)
		if err != nil {
			return token, nil, err
		}
		val.DefaultValue = v
	}

	// parse directives for input
	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		var (
			ds  []*ast.Directive
			err error
		)
		token, ds, err = parseDirectives(tokens)
		if err != nil {
			return token, nil, err
		}
		val.Directives = ds
	}

	return token, val, nil
}

func parseFragment(tokens chan lexer.Token) (lexer.Token, *ast.Fragment, error) {
	f := new(ast.Fragment)
	var err error

	token := <-tokens
	if token.Kind == lexer.NameToken && token.Value != "on" {
		f.Name = token.Value
	} else {
		return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
	}

	token = <-tokens
	if token.Kind == lexer.NameToken && token.Value == "on" {
		token = <-tokens
		if token.Kind == lexer.NameToken {
			f.TypeCondition = token.Value
		} else {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}
	} else {
		return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
	}

	token = <-tokens
	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		ds := []*ast.Directive{}
		token, ds, err = parseDirectives(tokens)
		if err != nil {
			return token, nil, err
		}
		f.Directives = ds
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		sSet := []ast.Selection{}
		token, sSet, err = parseSelectionSet(tokens)
		if err != nil {
			return token, nil, err
		}
		f.SelectionSet = sSet
	} else {
		return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
	}

	return token, f, nil
}

func parseOperation(token lexer.Token, tokens chan lexer.Token) (lexer.Token, *ast.Operation, error) {
	var err error
	if token.Kind != lexer.NameToken {
		return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
	}
	var ot ast.OperationType
	if token.Value == "query" {
		ot = ast.Query
	} else if token.Value == "mutation" {
		ot = ast.Mutation
	} else if token.Value == "subscription" {
		ot = ast.Subscription
	} else {
		return token, nil, fmt.Errorf("invalid token value: %s, wanted one of 'query', 'mutation' or 'subscription'", token.Value)
	}

	op := ast.NewOperation(ot)

	token = <-tokens
	for {
		switch {
		case token.Kind == lexer.NameToken:
			op.Name = token.Value
			token = <-tokens
			break
		case token.Kind == lexer.PunctuatorToken && token.Value == "(":
			vs := []*ast.Variable{}
			token, vs, err = parseVariables(tokens)
			if err != nil {
				return token, nil, err
			}
			op.Variables = vs
			break
		case token.Kind == lexer.PunctuatorToken && token.Value == "@":
			ds := []*ast.Directive{}
			var err error
			token, ds, err = parseDirectives(tokens)
			if err != nil {
				return token, nil, err
			}
			op.Directives = ds
			break
		case token.Kind == lexer.PunctuatorToken && token.Value == "{":
			sSet := []ast.Selection{}
			token, sSet, err = parseSelectionSet(tokens)
			if err != nil {
				return token, nil, err
			}
			op.SelectionSet = sSet
			return token, op, nil
		default:
			return token, nil, fmt.Errorf("invalid token value: %s", token.Value)
		}
	}
}

func parseVariables(tokens chan lexer.Token) (lexer.Token, []*ast.Variable, error) {
	token := <-tokens
	vs := []*ast.Variable{}
	var err error

	for {
		if token.Kind == lexer.PunctuatorToken && token.Value == ")" {
			return <-tokens, vs, nil
		}

		v := new(ast.Variable)
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		if token.Kind == lexer.PunctuatorToken && token.Value == "$" {
			token = <-tokens
			if token.Kind == lexer.NameToken {
				v.Name = token.Value
			} else {
				return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
			}
		} else {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}

		token = <-tokens
		if token.Kind == lexer.PunctuatorToken && token.Value == ":" {
			token = <-tokens
		} else {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}

		var t ast.Type
		token, t, err = parseType(token, tokens)
		if err != nil {
			return token, nil, err
		}
		v.Type = t

		if token.Kind == lexer.PunctuatorToken && token.Value == "=" {
			var dv ast.Value
			token, dv, err = parseValue(<-tokens, tokens)
			if err != nil {
				return token, nil, err
			}
			v.DefaultValue = dv
		}

		if token.Kind == lexer.PunctuatorToken && token.Value == "$" {
			vs = append(vs, v)
			continue
		} else if token.Kind == lexer.PunctuatorToken && token.Value == ")" {
			vs = append(vs, v)
			return <-tokens, vs, nil
		}
	}
}

func parseType(token lexer.Token, tokens chan lexer.Token) (lexer.Token, ast.Type, error) {
	switch {
	case token.Kind == lexer.NameToken:
		nt := new(ast.NamedType)
		nt.Name = token.Value
		nt.Location.Column = token.Col
		nt.Location.Line = token.Line - 1

		token = <-tokens
		if token.Kind == lexer.PunctuatorToken && token.Value == "!" {
			nnt := new(ast.NonNullType)
			nnt.Type = nt
			nnt.Location.Column = token.Col
			nnt.Location.Line = token.Line - 1
			return <-tokens, nnt, nil
		}
		return token, nt, nil
	case token.Kind == lexer.PunctuatorToken && token.Value == "[":
		token = <-tokens
		var (
			t   ast.Type
			err error
		)
		loc := ast.Location{
			Column: token.Col,
			Line:   token.Line,
		}
		token, t, err = parseType(token, tokens)
		if err != nil {
			return token, nil, err
		}

		if token.Kind == lexer.PunctuatorToken && token.Value == "]" {
			token = <-tokens
		} else {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}

		lt := &ast.ListType{
			Type:     t,
			Location: loc,
		}

		if token.Kind == lexer.PunctuatorToken && token.Value == "!" {
			return <-tokens, &ast.NonNullType{
				Type: lt,
			}, nil
		}
		return token, lt, nil
	default:
		return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
	}
}

func parseSelectionSet(tokens chan lexer.Token) (token lexer.Token, set []ast.Selection, err error) {
	end := false
	token = <-tokens
	for {
		switch {
		case token.Kind == lexer.PunctuatorToken && token.Value == "...":
			var sel ast.Selection
			token, sel, err = parseFragments(tokens)
			if err != nil {
				return token, nil, err
			}
			set = append(set, sel)
			break
		case token.Kind == lexer.NameToken:
			f := new(ast.Field)
			token, f, err = parseField(token, tokens)
			if err != nil {
				return token, nil, err
			}
			set = append(set, f)
			break
		case token.Kind == lexer.PunctuatorToken && token.Value == "}":
			end = true
			break
		default:
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}
		if end {
			break
		}
	}
	return <-tokens, set, nil
}

func parseField(token lexer.Token, tokens chan lexer.Token) (lexer.Token, *ast.Field, error) {
	var err error
	f := new(ast.Field)
	f.Alias = token.Value
	f.Location.Column = token.Col
	f.Location.Line = token.Line - 1
	defer func() {
		if f.Name == "" {
			f.Name = f.Alias
		}
	}()

	end := false
	token = <-tokens
	for {
		switch {
		case token.Kind == lexer.PunctuatorToken && token.Value == ":" && f.Name == "":
			token = <-tokens
			if token.Kind == lexer.NameToken {
				f.Name = token.Value
			} else {
				return token, nil, fmt.Errorf("unexpected token, expected name token, got: %s", token.Value)
			}
			token = <-tokens
			break
		case token.Kind == lexer.PunctuatorToken && token.Value == "(":
			args, err := parseArguments(tokens)
			if err != nil {
				return <-tokens, nil, err
			}
			f.Arguments = args
			token = <-tokens
			break
		case token.Kind == lexer.PunctuatorToken && token.Value == "@":
			ds := []*ast.Directive{}
			var err error
			token, ds, err = parseDirectives(tokens)
			if err != nil {
				return token, nil, err
			}
			f.Directives = ds
		case token.Kind == lexer.PunctuatorToken && token.Value == "{":
			sSet := []ast.Selection{}
			token, sSet, err = parseSelectionSet(tokens)
			if err != nil {
				return token, nil, err
			}
			f.SelectionSet = sSet
			return token, f, nil
		case token.Kind == lexer.NameToken:
			return token, f, nil
		case token.Kind == lexer.PunctuatorToken && token.Value == "}":
			return token, f, nil
		default:
			if f.Alias != "" {
				return token, f, nil
			}
			return token, nil, fmt.Errorf("invalid token: '%s'", token.Value)
		}

		if end {
			break
		}
	}
	return token, f, nil
}

func parseArguments(tokens chan lexer.Token) (args []*ast.Argument, err error) {
	token := <-tokens
	for {
		arg := new(ast.Argument)
		if token.Kind == lexer.NameToken {
			arg.Name = token.Value
		} else {
			return nil, fmt.Errorf("unexpected token args: %+v", token.Value)
		}

		token = <-tokens
		if token.Kind != lexer.PunctuatorToken && token.Value != ":" {
			return nil, fmt.Errorf("unexpected token, expected ':'")
		}

		token = <-tokens
		var val ast.Value
		token, val, err = parseValue(token, tokens)
		if err != nil {
			return nil, err
		}
		arg.Value = val
		args = append(args, arg)

		if token.Kind == lexer.PunctuatorToken && token.Value == ")" {
			return args, nil
		}
	}
}

func parseValue(token lexer.Token, tokens chan lexer.Token) (lexer.Token, ast.Value, error) {
	switch {
	case token.Kind == lexer.PunctuatorToken && token.Value == "$":
		v := new(ast.VariableValue)
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		token = <-tokens
		if token.Kind != lexer.NameToken {
			return token, nil, fmt.Errorf("invalid token")
		}
		v.Name = token.Value
		return <-tokens, v, nil
	case token.Kind == lexer.IntValueToken:
		v := new(ast.IntValue)
		v.Value = token.Value
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		return <-tokens, v, nil
	case token.Kind == lexer.FloatValueToken:
		v := new(ast.FloatValue)
		v.Value = token.Value
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		return <-tokens, v, nil
	case token.Kind == lexer.StringValueToken:
		v := new(ast.StringValue)
		v.Value = token.Value
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		return <-tokens, v, nil
	case token.Kind == lexer.NameToken && (token.Value == "false" || token.Value == "true"):
		v := new(ast.BooleanValue)
		v.Value = token.Value
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		return <-tokens, v, nil
	case token.Kind == lexer.NameToken && token.Value == "null":
		v := new(ast.NullValue)
		v.Value = token.Value
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		return <-tokens, v, nil
	case token.Kind == lexer.NameToken:
		v := new(ast.EnumValue)
		v.Value = token.Value
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		return <-tokens, v, nil
	case token.Kind == lexer.PunctuatorToken && token.Value == "[":
		return parseListValue(tokens)
	case token.Kind == lexer.PunctuatorToken && token.Value == "{":
		return parseObjectValue(tokens)
	}
	return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
}

func parseListValue(tokens chan lexer.Token) (lexer.Token, *ast.ListValue, error) {
	list := new(ast.ListValue)
	token := <-tokens
	for {
		if token.Kind == lexer.PunctuatorToken && token.Value == "]" {
			return <-tokens, list, nil
		}

		var (
			err error
			v   ast.Value
		)
		token, v, err = parseValue(token, tokens)
		if err != nil {
			return token, nil, err
		}
		list.Values = append(list.Values, v)
	}
}

func parseObjectValue(tokens chan lexer.Token) (lexer.Token, *ast.ObjectValue, error) {
	token := <-tokens
	var err error
	o := new(ast.ObjectValue)
	o.Fields = []*ast.ObjectFieldValue{}
	for {
		field := new(ast.ObjectFieldValue)
		field.Location.Column = token.Col
		field.Location.Line = token.Line - 1

		if token.Kind == lexer.NameToken {
			field.Name = token.Value
		} else {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}

		token = <-tokens
		if token.Kind != lexer.PunctuatorToken && token.Value != ":" {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}

		token = <-tokens
		var val ast.Value
		token, val, err = parseValue(token, tokens)
		if err != nil {
			return token, nil, err
		}
		field.Value = val
		o.Fields = append(o.Fields, field)

		if token.Kind == lexer.PunctuatorToken && token.Value == "}" {
			return <-tokens, o, nil
		}
	}
}

func parseFragments(tokens chan lexer.Token) (token lexer.Token, sel ast.Selection, err error) {
	token = <-tokens
	if token.Kind == lexer.NameToken && token.Value == "on" {
		inf := new(ast.InlineFragment)

		token = <-tokens
		if token.Kind == lexer.NameToken {
			inf.TypeCondition = token.Value
			token = <-tokens
		}

		if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
			ds := []*ast.Directive{}
			token, ds, err = parseDirectives(tokens)
			if err != nil {
				return token, nil, err
			}
			inf.Directives = ds
		}

		if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
			sSet := []ast.Selection{}
			token, sSet, err = parseSelectionSet(tokens)
			if err != nil {
				return token, nil, err
			}
			inf.SelectionSet = sSet
		} else {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}

		return token, inf, nil
	} else if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		sSet := []ast.Selection{}
		loc := ast.Location{
			Line:   token.Line,
			Column: token.Col,
		}
		token, sSet, err = parseSelectionSet(tokens)
		if err != nil {
			return token, nil, err
		}
		inf := &ast.InlineFragment{
			SelectionSet: sSet,
			Location:     loc,
		}

		return token, inf, nil
	} else if token.Kind == lexer.NameToken && token.Value != "on" {
		fs := new(ast.FragmentSpread)
		fs.Name = token.Value
		fs.Location.Column = token.Col
		fs.Location.Line = token.Line - 1
		token = <-tokens

		if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
			ds := []*ast.Directive{}
			token, ds, err = parseDirectives(tokens)
			if err != nil {
				return token, nil, err
			}
			fs.Directives = ds
		}

		return token, fs, nil
	}
	return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
}

func parseDirectives(tokens chan lexer.Token) (lexer.Token, []*ast.Directive, error) {
	token := <-tokens
	ds := []*ast.Directive{}
	for {
		if token.Kind == lexer.NameToken {
			d := new(ast.Directive)
			d.Name = token.Value
			d.Location.Column = token.Col
			d.Location.Line = token.Line - 1
			token = <-tokens

			if token.Kind == lexer.PunctuatorToken && token.Value == "(" {
				args, err := parseArguments(tokens)
				if err != nil {
					return token, nil, err
				}
				d.Arguments = args
				token = <-tokens
			}
			ds = append(ds, d)

			if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
				token = <-tokens
			} else {
				return token, ds, nil
			}
		} else {
			return token, ds, nil
		}
	}
}
