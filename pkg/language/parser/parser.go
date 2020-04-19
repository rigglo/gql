package parser

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/rigglo/gql/pkg/language/ast"
	"github.com/rigglo/gql/pkg/language/lexer"
)

// Parse parses a gql query
func Parse(query []byte) (lexer.Token, *ast.Document, error) {
	tokens := make(chan lexer.Token)
	src := bytes.NewReader(query)
	readr := bufio.NewReader(src)
	go lexer.Lex(readr, tokens)
	t, doc, err := parseDocument(tokens)
	if err != nil && t.Value == "" {
		return t, nil, fmt.Errorf("unexpected EOF")
	}
	return t, doc, err
}

func parseDocument(tokens chan lexer.Token) (lexer.Token, *ast.Document, error) {
	doc := ast.NewDocument()
	var err error
	token := <-tokens
	for {
		switch {
		case token.Kind == lexer.NameToken && token.Value == "fragment":
			f := new(ast.Fragment)
			token, f, err = parseFragment(tokens)
			if err != nil {
				return token, nil, err
			}
			doc.Fragments = append(doc.Fragments, f)
			break
		case token.Kind == lexer.NameToken:
			op := new(ast.Operation)
			token, op, err = parseOperation(token, tokens)
			if err != nil {
				return token, nil, err
			}
			doc.Operations = append(doc.Operations, op)
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
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}
	}
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
