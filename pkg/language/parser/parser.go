package parser

import (
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
	lex := lexer.NewLexer(lexer.NewInput(document))
	t, doc, err := parseDocument(lex)
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
	lex := lexer.NewLexer(lexer.NewInput(definition))
	token := lex.Read()

	desc := ""
	if token.Kind == lexer.StringValueToken {
		desc = strings.Trim(token.Value, `"`)
		token = lex.Read()
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
		t, def, err := parseSchema(lex.Read(), lex)
		if err != nil {
			return nil, err
		} else if t.Kind != lexer.EOFToken && err == nil {
			return nil, &ParserError{
				Message: fmt.Sprintf("invalid token after schema definition: '%s'", t.Value),
				Line:    token.Line,
				Column:  token.Col,
				Token:   token,
			}
		}
		return def, nil
	}
	t, def, err := parseDefinition(token, lex, desc)
	if err != nil {
		return nil, &ParserError{
			Message: err.Error(),
			Line:    token.Line,
			Column:  token.Col,
			Token:   token,
		}
	} else if t.Kind != lexer.EOFToken && err == nil {
		return nil, &ParserError{
			Message: fmt.Sprintf("invalid token after definition: '%s'", t.Value),
			Line:    token.Line,
			Column:  token.Col,
			Token:   token,
		}
	}
	return def, nil
}

func parseDocument(lex *lexer.Lexer) (lexer.Token, *ast.Document, error) {
	doc := ast.NewDocument()
	var err error
	token := lex.Read()
	for {
		switch {
		case token.Kind == lexer.NameToken:
			switch token.Value {
			case "fragment":
				{
					f := new(ast.Fragment)
					token, f, err = parseFragment(lex)
					if err != nil {
						return token, nil, err
					}
					doc.Fragments = append(doc.Fragments, f)
				}
			case "query", "mutation", "subscription":
				{
					op := new(ast.Operation)
					token, op, err = parseOperation(token, lex)
					if err != nil {
						return token, nil, err
					}
					doc.Operations = append(doc.Operations, op)
				}
			case "scalar", "type", "interface", "union", "enum", "input", "directive":
				var def ast.Definition
				token, def, err = parseDefinition(token, lex, "")
				if err != nil {
					return token, nil, err
				}
				doc.Definitions = append(doc.Definitions, def)
			case "schema":
				var def ast.Definition
				token, def, err = parseSchema(token, lex)
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
			token = lex.Read()
			switch token.Value {
			case "scalar", "type", "interface", "union", "enum", "input", "directive":
				var def ast.Definition
				token, def, err = parseDefinition(token, lex, desc)
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
			token, set, err = parseSelectionSet(lex)
			if err != nil {
				return token, nil, err
			}
			doc.Operations = append(doc.Operations, &ast.Operation{
				OperationType: ast.Query,
				SelectionSet:  set,
			})
			break
		case token.Kind == lexer.EOFToken && token.Err == nil:
			return token, doc, nil
		default:
			return token, nil, fmt.Errorf("unexpected token: %s, kind: %v", token.Value, token.Kind)
		}
	}
}

func parseDefinition(token lexer.Token, lex *lexer.Lexer, desc string) (lexer.Token, ast.Definition, error) {
	switch token.Value {
	case "scalar":
		return parseScalar(lex.Read(), lex, desc)
	case "type":
		return parseObject(lex.Read(), lex, desc)
	case "interface":
		return parseInterface(lex.Read(), lex, desc)
	case "union":
		return parseUnion(lex.Read(), lex, desc)
	case "enum":
		return parseEnum(lex.Read(), lex, desc)
	case "input":
		return parseInputObject(lex.Read(), lex, desc)
	case "directive":
		return parseDirectiveDefinition(lex.Read(), lex, desc)
	}
	return token, nil, fmt.Errorf("expected a schema, type or directive definition, got: '%s', err: '%v'", token.Value, token.Err)
}

func parseSchema(token lexer.Token, lex *lexer.Lexer) (lexer.Token, ast.Definition, error) {
	def := new(ast.SchemaDefinition)

	// parse Name
	if token.Kind != lexer.NameToken {
		log.Println(token.Kind)
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = lex.Read()

	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		var (
			ds  []*ast.Directive
			err error
		)
		token, ds, err = parseDirectives(lex)
		if err != nil {
			return token, nil, err
		}
		def.Directives = ds
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		def.RootOperations = map[ast.OperationType]*ast.NamedType{}

		token = lex.Read()
		for {
			// quit if it's the end of the field definition list
			if token.Kind == lexer.PunctuatorToken && token.Value == "}" {
				return lex.Read(), def, nil
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
				token = lex.Read()
			} else {
				return token, nil, fmt.Errorf("expected NameToken, got '%s'", token.Value)
			}

			if token.Kind == lexer.PunctuatorToken && token.Value == ":" {
				token = lex.Read()
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
				token = lex.Read()
			} else {
				return token, nil, fmt.Errorf("expected NameToken, got '%s'", token.Value)
			}
		}
	}
	return token, nil, fmt.Errorf("expected '{', and a list of root operation types, got '%s'", token.Value)
}

func parseScalar(token lexer.Token, lex *lexer.Lexer, desc string) (lexer.Token, ast.Definition, error) {
	def := &ast.ScalarDefinition{
		Description: desc,
	}

	// parse Name
	if token.Kind != lexer.NameToken {
		log.Println(token.Kind)
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = lex.Read()

	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		token, ds, err := parseDirectives(lex)
		if err != nil {
			return token, nil, err
		}
		def.Directives = ds
		return token, def, nil
	}

	return token, def, nil
}

func parseObject(token lexer.Token, lex *lexer.Lexer, desc string) (lexer.Token, ast.Definition, error) {
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
	token = lex.Read()

	if token.Kind == lexer.NameToken && token.Value == "implements" {
		token = lex.Read()
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
				token = lex.Read()
			} else if token.Kind == lexer.PunctuatorToken && token.Value == "&" {
				token = lex.Read()
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
		token, ds, err = parseDirectives(lex)
		if err != nil {
			return token, nil, fmt.Errorf("couldn't parse directives: %v", err)
		}
		def.Directives = ds
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		token = lex.Read()
		for {
			// quit if it's the end of the field definition list
			if token.Kind == lexer.PunctuatorToken && token.Value == "}" {
				return lex.Read(), def, nil
			}
			field := &ast.FieldDefinition{}

			// parse optional description
			if token.Kind == lexer.StringValueToken {
				field.Description = token.Value
				token = lex.Read()
			}
			if token.Kind == lexer.NameToken {
				field.Name = token.Value
				token = lex.Read()
			} else {
				return token, nil, fmt.Errorf("expected NameToken, got token kind '%v', with value '%s'", token.Kind, token.Value)
			}

			// parse arguments definition
			if token.Kind == lexer.PunctuatorToken && token.Value == "(" {
				token = lex.Read()
				field.Arguments = []*ast.InputValueDefinition{}
				for {
					if token.Kind == lexer.PunctuatorToken && token.Value == ")" {
						token = lex.Read()
						break
					}
					var (
						inputDef *ast.InputValueDefinition
						err      error
					)
					token, inputDef, err = parseInputValueDefinition(token, lex)
					if err != nil {
						return token, nil, err
					}
					field.Arguments = append(field.Arguments, inputDef)
				}
			}

			if token.Kind == lexer.PunctuatorToken && token.Value == ":" {
				token = lex.Read()
			} else {
				return token, nil, fmt.Errorf("expected ':', got token kind '%v', with value '%s'", token.Kind, token.Value)
			}

			// parse the type of the field
			var (
				t   ast.Type
				err error
			)
			token, t, err = parseType(token, lex)
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
				token, ds, err = parseDirectives(lex)
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

func parseInterface(token lexer.Token, lex *lexer.Lexer, desc string) (lexer.Token, ast.Definition, error) {
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
	token = lex.Read()

	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		var (
			ds  []*ast.Directive
			err error
		)
		token, ds, err = parseDirectives(lex)
		if err != nil {
			return token, nil, fmt.Errorf("couldn't parse directives: %v", err)
		}
		def.Directives = ds
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		token = lex.Read()
		for {
			// quit if it's the end of the field definition list
			if token.Kind == lexer.PunctuatorToken && token.Value == "}" {
				return lex.Read(), def, nil
			}
			field := &ast.FieldDefinition{}

			// parse optional description
			if token.Kind == lexer.StringValueToken {
				field.Description = strings.Trim(token.Value, `"`)
				token = lex.Read()
			}
			if token.Kind == lexer.NameToken {
				field.Name = token.Value
				token = lex.Read()
			} else {
				return token, nil, fmt.Errorf("expected NameToken, got token kind '%v', with value '%s'", token.Kind, token.Value)
			}

			// parse arguments definition
			if token.Kind == lexer.PunctuatorToken && token.Value == "(" {
				token = lex.Read()
				field.Arguments = []*ast.InputValueDefinition{}
				for {
					if token.Kind == lexer.PunctuatorToken && token.Value == ")" {
						token = lex.Read()
						break
					}
					var (
						inputDef *ast.InputValueDefinition
						err      error
					)
					token, inputDef, err = parseInputValueDefinition(token, lex)
					if err != nil {
						return token, nil, err
					}
					field.Arguments = append(field.Arguments, inputDef)
				}
			}

			if token.Kind == lexer.PunctuatorToken && token.Value == ":" {
				token = lex.Read()
			} else {
				return token, nil, fmt.Errorf("expected ':', got token kind '%v', with value '%s'", token.Kind, token.Value)
			}

			// parse the type of the field
			var (
				t   ast.Type
				err error
			)
			token, t, err = parseType(token, lex)
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
				token, ds, err = parseDirectives(lex)
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

func parseUnion(token lexer.Token, lex *lexer.Lexer, desc string) (lexer.Token, ast.Definition, error) {
	def := &ast.UnionDefinition{
		Description: desc,
	}

	// parse Name
	if token.Kind != lexer.NameToken {
		log.Println(token.Kind)
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = lex.Read()

	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		var (
			ds  []*ast.Directive
			err error
		)
		token, ds, err = parseDirectives(lex)
		if err != nil {
			return token, nil, fmt.Errorf("couldn't parse directives: %v", err)
		}
		def.Directives = ds
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "=" {
		def.Members = []*ast.NamedType{}
		token = lex.Read()
		if token.Kind == lexer.PunctuatorToken && token.Value == "|" {
			token = lex.Read()
		}
		if token.Kind == lexer.NameToken {
			def.Members = append(def.Members, &ast.NamedType{
				Location: ast.Location{
					Column: token.Col,
					Line:   token.Line,
				},
				Name: token.Value,
			})
			token = lex.Read()
		} else {
			return token, nil, fmt.Errorf("expected Name token, got '%s'", token.Value)
		}
		for {
			if token.Kind == lexer.PunctuatorToken && token.Value == "|" {
				token = lex.Read()
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
				token = lex.Read()
			} else {
				return token, nil, fmt.Errorf("expected Name token, got '%s'", token.Value)
			}
		}
	}
	return token, def, nil
}

func parseEnum(token lexer.Token, lex *lexer.Lexer, desc string) (lexer.Token, ast.Definition, error) {
	def := &ast.EnumDefinition{
		Description: desc,
	}

	// parse Name
	if token.Kind != lexer.NameToken {
		log.Println(token.Kind)
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = lex.Read()

	// parse directives for the enum type
	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		var (
			ds  []*ast.Directive
			err error
		)
		token, ds, err = parseDirectives(lex)
		if err != nil {
			return token, nil, fmt.Errorf("couldn't parse directives: %v", err)
		}
		def.Directives = ds
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		def.Values = []*ast.EnumValueDefinition{}
		token = lex.Read()

		// parse all the enum value definitions
		for {
			if token.Kind == lexer.PunctuatorToken && token.Value == "}" {
				return lex.Read(), def, nil
			}

			enumV := &ast.EnumValueDefinition{}
			if token.Kind == lexer.StringValueToken {
				enumV.Description = strings.Trim(token.Value, `"`)
				token = lex.Read()
			}
			if token.Kind == lexer.NameToken {
				enumV.Value = &ast.EnumValue{
					Location: ast.Location{
						Column: token.Col,
						Line:   token.Line,
					},
					Value: token.Value,
				}
				token = lex.Read()
			} else {
				return token, nil, fmt.Errorf("expected Name token, got '%v'", token.Value)
			}
			if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
				var (
					ds  []*ast.Directive
					err error
				)
				token, ds, err = parseDirectives(lex)
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

func parseDirectiveDefinition(token lexer.Token, lex *lexer.Lexer, desc string) (lexer.Token, ast.Definition, error) {
	def := &ast.DirectiveDefinition{
		Description: desc,
	}

	// parse Name
	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		token = lex.Read()
	} else {
		return token, nil, fmt.Errorf("expected token '@', got: '%s'", token.Value)
	}
	if token.Kind != lexer.NameToken {
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = lex.Read()

	if token.Kind == lexer.NameToken && token.Value == "on" {
		def.Locations = []string{}
		token = lex.Read()

		if token.Kind == lexer.PunctuatorToken && token.Value == "|" {
			token = lex.Read()
		}
		if token.Kind == lexer.NameToken {
			def.Locations = append(def.Locations, token.Value)
			token = lex.Read()
		} else {
			return token, nil, fmt.Errorf("expected Name token, got '%s'", token.Value)
		}
		for {
			if token.Kind == lexer.PunctuatorToken && token.Value == "|" {
				token = lex.Read()
			} else {
				return token, def, nil
			}
			if token.Kind == lexer.NameToken && ast.IsValidDirective(token.Value) {
				def.Locations = append(def.Locations, token.Value)
				token = lex.Read()
			} else {
				return token, nil, fmt.Errorf("expected Name token with a valid directive locaton, got '%s'", token.Value)
			}
		}
	}
	return token, def, nil
}

func parseInputObject(token lexer.Token, lex *lexer.Lexer, desc string) (lexer.Token, ast.Definition, error) {
	def := &ast.InputObjectDefinition{
		Description: desc,
	}

	// parse Name
	if token.Kind != lexer.NameToken {
		log.Println(token.Kind)
		return token, nil, fmt.Errorf("expected NameToken, got: '%s', err: '%v'", token.Value, token.Err)
	}
	def.Name = token.Value
	token = lex.Read()

	// parse directives for the enum type
	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		var (
			ds  []*ast.Directive
			err error
		)
		token, ds, err = parseDirectives(lex)
		if err != nil {
			return token, nil, fmt.Errorf("couldn't parse directives: %v", err)
		}
		def.Directives = ds
	}

	// parse input object field definitions
	if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		token = lex.Read()
		def.Fields = []*ast.InputValueDefinition{}
		for {
			if token.Kind == lexer.PunctuatorToken && token.Value == "}" {
				token = lex.Read()
				break
			}
			var (
				inputDef *ast.InputValueDefinition
				err      error
			)
			token, inputDef, err = parseInputValueDefinition(token, lex)
			if err != nil {
				return token, nil, err
			}
			def.Fields = append(def.Fields, inputDef)
		}
	}
	return token, def, nil
}

func parseInputValueDefinition(token lexer.Token, lex *lexer.Lexer) (lexer.Token, *ast.InputValueDefinition, error) {
	val := &ast.InputValueDefinition{}

	// parse description for input
	if token.Kind == lexer.StringValueToken {
		val.Description = token.Value
		token = lex.Read()
	}

	// parse name of the input
	if token.Kind == lexer.NameToken {
		val.Name = token.Value
		token = lex.Read()
	} else {
		return token, nil, fmt.Errorf("expected NameToken, got token kind '%v', with value '%s'", token.Kind, token.Value)
	}

	// parse type for the input
	if token.Kind == lexer.PunctuatorToken && token.Value == ":" {
		token = lex.Read()
		var (
			t   ast.Type
			err error
		)
		token, t, err = parseType(token, lex)
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
		token = lex.Read()
		var (
			v   ast.Value
			err error
		)
		token, v, err = parseValue(token, lex)
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
		token, ds, err = parseDirectives(lex)
		if err != nil {
			return token, nil, err
		}
		val.Directives = ds
	}

	return token, val, nil
}

func parseFragment(lex *lexer.Lexer) (lexer.Token, *ast.Fragment, error) {
	f := new(ast.Fragment)
	var err error

	token := lex.Read()
	if token.Kind == lexer.NameToken && token.Value != "on" {
		f.Name = token.Value
	} else {
		return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
	}

	token = lex.Read()
	if token.Kind == lexer.NameToken && token.Value == "on" {
		token = lex.Read()
		if token.Kind == lexer.NameToken {
			f.TypeCondition = token.Value
		} else {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}
	} else {
		return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
	}

	token = lex.Read()
	if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
		ds := []*ast.Directive{}
		token, ds, err = parseDirectives(lex)
		if err != nil {
			return token, nil, err
		}
		f.Directives = ds
	}

	if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
		sSet := []ast.Selection{}
		token, sSet, err = parseSelectionSet(lex)
		if err != nil {
			return token, nil, err
		}
		f.SelectionSet = sSet
	} else {
		return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
	}

	return token, f, nil
}

func parseOperation(token lexer.Token, lex *lexer.Lexer) (lexer.Token, *ast.Operation, error) {
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

	token = lex.Read()
	for {
		switch {
		case token.Kind == lexer.NameToken:
			op.Name = token.Value
			token = lex.Read()
			break
		case token.Kind == lexer.PunctuatorToken && token.Value == "(":
			vs := []*ast.Variable{}
			token, vs, err = parseVariables(lex)
			if err != nil {
				return token, nil, err
			}
			op.Variables = vs
			break
		case token.Kind == lexer.PunctuatorToken && token.Value == "@":
			ds := []*ast.Directive{}
			var err error
			token, ds, err = parseDirectives(lex)
			if err != nil {
				return token, nil, err
			}
			op.Directives = ds
			break
		case token.Kind == lexer.PunctuatorToken && token.Value == "{":
			sSet := []ast.Selection{}
			token, sSet, err = parseSelectionSet(lex)
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

func parseVariables(lex *lexer.Lexer) (lexer.Token, []*ast.Variable, error) {
	token := lex.Read()
	vs := []*ast.Variable{}
	var err error

	for {
		if token.Kind == lexer.PunctuatorToken && token.Value == ")" {
			return lex.Read(), vs, nil
		}

		v := new(ast.Variable)
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		if token.Kind == lexer.PunctuatorToken && token.Value == "$" {
			token = lex.Read()
			if token.Kind == lexer.NameToken {
				v.Name = token.Value
			} else {
				return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
			}
		} else {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}

		token = lex.Read()
		if token.Kind == lexer.PunctuatorToken && token.Value == ":" {
			token = lex.Read()
		} else {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}

		var t ast.Type
		token, t, err = parseType(token, lex)
		if err != nil {
			return token, nil, err
		}
		v.Type = t

		if token.Kind == lexer.PunctuatorToken && token.Value == "=" {
			var dv ast.Value
			token, dv, err = parseValue(lex.Read(), lex)
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
			return lex.Read(), vs, nil
		}
	}
}

func parseType(token lexer.Token, lex *lexer.Lexer) (lexer.Token, ast.Type, error) {
	switch {
	case token.Kind == lexer.NameToken:
		nt := new(ast.NamedType)
		nt.Name = token.Value
		nt.Location.Column = token.Col
		nt.Location.Line = token.Line - 1

		token = lex.Read()
		if token.Kind == lexer.PunctuatorToken && token.Value == "!" {
			nnt := new(ast.NonNullType)
			nnt.Type = nt
			nnt.Location.Column = token.Col
			nnt.Location.Line = token.Line - 1
			return lex.Read(), nnt, nil
		}
		return token, nt, nil
	case token.Kind == lexer.PunctuatorToken && token.Value == "[":
		token = lex.Read()
		var (
			t   ast.Type
			err error
		)
		loc := ast.Location{
			Column: token.Col,
			Line:   token.Line,
		}
		token, t, err = parseType(token, lex)
		if err != nil {
			return token, nil, err
		}

		if token.Kind == lexer.PunctuatorToken && token.Value == "]" {
			token = lex.Read()
		} else {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}

		lt := &ast.ListType{
			Type:     t,
			Location: loc,
		}

		if token.Kind == lexer.PunctuatorToken && token.Value == "!" {
			return lex.Read(), &ast.NonNullType{
				Type: lt,
			}, nil
		}
		return token, lt, nil
	default:
		return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
	}
}

func parseSelectionSet(lex *lexer.Lexer) (token lexer.Token, set []ast.Selection, err error) {
	end := false
	token = lex.Read()
	for {
		switch {
		case token.Kind == lexer.PunctuatorToken && token.Value == "...":
			var sel ast.Selection
			token, sel, err = parseFragments(lex)
			if err != nil {
				return token, nil, err
			}
			set = append(set, sel)
			break
		case token.Kind == lexer.NameToken:
			f := new(ast.Field)
			token, f, err = parseField(token, lex)
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
	return lex.Read(), set, nil
}

func parseField(token lexer.Token, lex *lexer.Lexer) (lexer.Token, *ast.Field, error) {
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
	token = lex.Read()
	for {
		switch {
		case token.Kind == lexer.PunctuatorToken && token.Value == ":" && f.Name == "":
			token = lex.Read()
			if token.Kind == lexer.NameToken {
				f.Name = token.Value
			} else {
				return token, nil, fmt.Errorf("unexpected token, expected name token, got: %s", token.Value)
			}
			token = lex.Read()
			break
		case token.Kind == lexer.PunctuatorToken && token.Value == "(":
			args, err := parseArguments(lex)
			if err != nil {
				return lex.Read(), nil, err
			}
			f.Arguments = args
			token = lex.Read()
			break
		case token.Kind == lexer.PunctuatorToken && token.Value == "@":
			ds := []*ast.Directive{}
			var err error
			token, ds, err = parseDirectives(lex)
			if err != nil {
				return token, nil, err
			}
			f.Directives = ds
		case token.Kind == lexer.PunctuatorToken && token.Value == "{":
			sSet := []ast.Selection{}
			token, sSet, err = parseSelectionSet(lex)
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

func parseArguments(lex *lexer.Lexer) (args []*ast.Argument, err error) {
	token := lex.Read()
	for {
		arg := new(ast.Argument)
		if token.Kind == lexer.NameToken {
			arg.Name = token.Value
		} else {
			return nil, fmt.Errorf("unexpected token args: %+v", token.Value)
		}

		token = lex.Read()
		if token.Kind != lexer.PunctuatorToken && token.Value != ":" {
			return nil, fmt.Errorf("unexpected token '%s' '%s', expected ':'", token.Value, token.Kind.String())
		}

		token = lex.Read()
		var val ast.Value
		token, val, err = parseValue(token, lex)
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

func parseValue(token lexer.Token, lex *lexer.Lexer) (lexer.Token, ast.Value, error) {
	switch {
	case token.Kind == lexer.PunctuatorToken && token.Value == "$":
		v := new(ast.VariableValue)
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		token = lex.Read()
		if token.Kind != lexer.NameToken {
			return token, nil, fmt.Errorf("invalid token")
		}
		v.Name = token.Value
		return lex.Read(), v, nil
	case token.Kind == lexer.IntValueToken:
		v := new(ast.IntValue)
		v.Value = token.Value
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		return lex.Read(), v, nil
	case token.Kind == lexer.FloatValueToken:
		v := new(ast.FloatValue)
		v.Value = token.Value
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		return lex.Read(), v, nil
	case token.Kind == lexer.StringValueToken:
		v := new(ast.StringValue)
		v.Value = token.Value
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		return lex.Read(), v, nil
	case token.Kind == lexer.NameToken && (token.Value == "false" || token.Value == "true"):
		v := new(ast.BooleanValue)
		v.Value = token.Value
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		return lex.Read(), v, nil
	case token.Kind == lexer.NameToken && token.Value == "null":
		v := new(ast.NullValue)
		v.Value = token.Value
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		return lex.Read(), v, nil
	case token.Kind == lexer.NameToken:
		v := new(ast.EnumValue)
		v.Value = token.Value
		v.Location.Column = token.Col
		v.Location.Line = token.Line - 1
		return lex.Read(), v, nil
	case token.Kind == lexer.PunctuatorToken && token.Value == "[":
		return parseListValue(lex)
	case token.Kind == lexer.PunctuatorToken && token.Value == "{":
		return parseObjectValue(lex)
	}
	return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
}

func parseListValue(lex *lexer.Lexer) (lexer.Token, *ast.ListValue, error) {
	list := new(ast.ListValue)
	token := lex.Read()
	for {
		if token.Kind == lexer.PunctuatorToken && token.Value == "]" {
			return lex.Read(), list, nil
		}

		var (
			err error
			v   ast.Value
		)
		token, v, err = parseValue(token, lex)
		if err != nil {
			return token, nil, err
		}
		list.Values = append(list.Values, v)
	}
}

func parseObjectValue(lex *lexer.Lexer) (lexer.Token, *ast.ObjectValue, error) {
	token := lex.Read()
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

		token = lex.Read()
		if token.Kind != lexer.PunctuatorToken && token.Value != ":" {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}

		token = lex.Read()
		var val ast.Value
		token, val, err = parseValue(token, lex)
		if err != nil {
			return token, nil, err
		}
		field.Value = val
		o.Fields = append(o.Fields, field)

		if token.Kind == lexer.PunctuatorToken && token.Value == "}" {
			return lex.Read(), o, nil
		}
	}
}

func parseFragments(lex *lexer.Lexer) (token lexer.Token, sel ast.Selection, err error) {
	token = lex.Read()
	if token.Kind == lexer.NameToken && token.Value == "on" {
		inf := new(ast.InlineFragment)

		token = lex.Read()
		if token.Kind == lexer.NameToken {
			inf.TypeCondition = token.Value
			token = lex.Read()
		}

		if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
			ds := []*ast.Directive{}
			token, ds, err = parseDirectives(lex)
			if err != nil {
				return token, nil, err
			}
			inf.Directives = ds
		}

		if token.Kind == lexer.PunctuatorToken && token.Value == "{" {
			sSet := []ast.Selection{}
			token, sSet, err = parseSelectionSet(lex)
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
		token, sSet, err = parseSelectionSet(lex)
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
		token = lex.Read()

		if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
			ds := []*ast.Directive{}
			token, ds, err = parseDirectives(lex)
			if err != nil {
				return token, nil, err
			}
			fs.Directives = ds
		}

		return token, fs, nil
	}
	return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
}

func parseDirectives(lex *lexer.Lexer) (lexer.Token, []*ast.Directive, error) {
	token := lex.Read()
	ds := []*ast.Directive{}
	for {
		if token.Kind == lexer.NameToken {
			d := new(ast.Directive)
			d.Name = token.Value
			d.Location.Column = token.Col
			d.Location.Line = token.Line - 1
			token = lex.Read()

			if token.Kind == lexer.PunctuatorToken && token.Value == "(" {
				args, err := parseArguments(lex)
				if err != nil {
					return token, nil, err
				}
				d.Arguments = args
				token = lex.Read()
			}
			ds = append(ds, d)

			if token.Kind == lexer.PunctuatorToken && token.Value == "@" {
				token = lex.Read()
			} else {
				return token, ds, nil
			}
		} else {
			return token, ds, nil
		}
	}
}
