package parse

import (
	"bufio"
	"fmt"
	"log"
	"strings"

	"github.com/rigglo/gql/utils/ast"
	"github.com/rigglo/gql/utils/lex"
)

// Parse parses a gql query
func Parse(query string) (lex.Token, *ast.Document, error) {
	tokens := make(chan lex.Token)
	src := strings.NewReader(query)
	readr := bufio.NewReader(src)
	go lex.Lex(readr, tokens)
	t, doc, err := parseDocument(tokens)
	if err != nil && t.Value == "" {
		return t, nil, fmt.Errorf("unexpected EOF")
	}
	return t, doc, err
}

func parseDocument(tokens chan lex.Token) (lex.Token, *ast.Document, error) {
	doc := ast.NewDocument()
	token := <-tokens
	switch {
	case token.Kind == lex.NameToken && token.Value == "fragment":
		// TODO: parse fragment definitions
		break
	case token.Kind == lex.NameToken:
		op := new(ast.Operation)
		token, op, err := parseOperation(token, tokens)
		if err != nil {
			return token, nil, err
		}
		doc.Operation = op
		break
	case token.Kind == lex.PunctuatorToken && token.Value == "{":
		token, set, err := parseSelectionSet(tokens)
		if err != nil {
			return token, nil, err
		}
		doc.Operation = &ast.Operation{
			OperationType: ast.Query,
			SelectionSet:  set,
		}
		token = <-tokens
		break
	default:
		return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
	}
	return token, doc, nil
}

func parseOperation(token lex.Token, tokens chan lex.Token) (lex.Token, *ast.Operation, error) {
	var err error
	if token.Kind != lex.NameToken {
		return token, nil, fmt.Errorf("unexpected token")
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
		log.Println("ops")
		switch {
		case token.Kind == lex.NameToken:
			op.Name = token.Value
			token = <-tokens
			break
		case token.Kind == lex.PunctuatorToken && token.Value == "@":
			ds := []*ast.Directive{}
			var err error
			token, ds, err = parseDirectives(tokens)
			if err != nil {
				return token, nil, err
			}
			op.Directives = ds
			break
		case token.Kind == lex.PunctuatorToken && token.Value == "{":
			sSet := []*ast.Selection{}
			token, sSet, err = parseSelectionSet(tokens)
			if err != nil {
				return token, nil, err
			}
			op.SelectionSet = sSet
			return token, op, nil
		}
		// TODO: Variables
	}
}

func parseSelectionSet(tokens chan lex.Token) (token lex.Token, set []*ast.Selection, err error) {
	log.Println("start selectionset")
	defer func() {
		log.Println("exit selectionset")
	}()
	end := false
	token = <-tokens
	for {

		log.Println("selectionset", token.Value)
		switch {
		case token.Kind == lex.PunctuatorToken && token.Value == "...":
			sel := new(ast.Selection)
			token, sel, err = parseFragments(tokens)
			if err != nil {
				return token, nil, err
			}
			set = append(set, sel)
			break
		case token.Kind == lex.NameToken:
			f := new(ast.Field)
			token, f, err = parseField(token, tokens)
			if err != nil {
				return token, nil, err
			}
			set = append(set, &ast.Selection{
				Kind:  ast.FieldSelectionKind,
				Field: f,
			})
			break
		case token.Kind == lex.PunctuatorToken && token.Value == "}":
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

func parseField(token lex.Token, tokens chan lex.Token) (lex.Token, *ast.Field, error) {
	var err error
	f := new(ast.Field)
	f.Alias = token.Value
	defer func() {
		if f.Name == "" {
			f.Name = f.Alias
		}
	}()

	end := false
	for {
		log.Println("fields")
		token = <-tokens
		switch {
		case token.Kind == lex.PunctuatorToken && token.Value == ":" && f.Name == "":
			token = <-tokens
			if token.Kind == lex.NameToken {
				f.Name = token.Value
			} else {
				return token, nil, fmt.Errorf("unexpected token, expected name token, got: %s", token.Value)
			}
			break
		case token.Kind == lex.PunctuatorToken && token.Value == "(":
			args, err := parseArguments(tokens)
			if err != nil {
				return <-tokens, nil, err
			}
			f.Arguments = args
			break
		case token.Kind == lex.PunctuatorToken && token.Value == "@":
			ds := []*ast.Directive{}
			var err error
			token, ds, err = parseDirectives(tokens)
			if err != nil {
				return token, nil, err
			}
			f.Directives = ds
			break
		case token.Kind == lex.PunctuatorToken && token.Value == "{":
			sSet := []*ast.Selection{}
			token, sSet, err = parseSelectionSet(tokens)
			if err != nil {
				return token, nil, err
			}
			f.SelectionSet = sSet
			return token, f, nil
		case token.Kind == lex.NameToken:
			return token, f, nil
		case token.Kind == lex.PunctuatorToken && token.Value == "}":
			return token, f, nil
		default:
			return token, nil, fmt.Errorf("invalid token")
		}

		if end {
			break
		}
	}
	return token, f, nil
}

func parseArguments(tokens chan lex.Token) (args []*ast.Argument, err error) {
	token := <-tokens
	for {
		log.Println("args")
		arg := new(ast.Argument)
		if token.Kind == lex.NameToken {
			arg.Name = token.Value
		} else {
			return nil, fmt.Errorf("unexpected token")
		}

		token = <-tokens
		if token.Kind != lex.PunctuatorToken && token.Value != ":" {
			return nil, fmt.Errorf("unexpected token, expected ':'")
		}

		token = <-tokens
		val, err := parseValue(token, tokens)
		if err != nil {
			return nil, err
		}
		arg.Value = val
		args = append(args, arg)

		token = <-tokens
		if token.Kind == lex.PunctuatorToken && token.Value == ")" {
			return args, nil
		}
	}
}

func parseValue(token lex.Token, tokens chan lex.Token) (ast.Value, error) {
	switch {
	case token.Kind == lex.PunctuatorToken && token.Value == "$":
		v := new(ast.VariableValue)
		token = <-tokens
		if token.Kind != lex.NameToken {
			return nil, fmt.Errorf("invalid token")
		}
		v.Name = token.Value
		return v, nil
	case token.Kind == lex.IntValueToken:
		v := new(ast.IntValue)
		v.Value = token.Value
		return v, nil
	case token.Kind == lex.FloatValueToken:
		v := new(ast.FloatValue)
		v.Value = token.Value
		return v, nil
	case token.Kind == lex.StringValueToken:
		v := new(ast.StringValue)
		v.Value = token.Value
		return v, nil
	case token.Kind == lex.NameToken && (token.Value == "false" || token.Value == "true"):
		v := new(ast.BooleanValue)
		v.Value = token.Value
		return v, nil
	case token.Kind == lex.NameToken && token.Value == "null":
		v := new(ast.NullValue)
		v.Value = token.Value
		return v, nil
	case token.Kind == lex.NameToken:
		v := new(ast.EnumValue)
		v.Value = token.Value
		return v, nil
	case token.Kind == lex.PunctuatorToken && token.Value == "[":
		return parseListValue(tokens)
	case token.Kind == lex.PunctuatorToken && token.Value == "{":
		return parseObjectValue(tokens)
	}
	return nil, fmt.Errorf("unexpected token: %s", token.Value)
}

func parseListValue(tokens chan lex.Token) (*ast.ListValue, error) {
	list := new(ast.ListValue)
	for {
		log.Println("lists")
		token := <-tokens
		if token.Kind == lex.PunctuatorToken && token.Value == "]" {
			return list, nil
		}

		v, err := parseValue(token, tokens)
		if err != nil {
			return nil, err
		}
		list.Values = append(list.Values, v)
	}
}

func parseObjectValue(tokens chan lex.Token) (*ast.ObjectValue, error) {
	token := <-tokens
	o := new(ast.ObjectValue)
	for {
		log.Println("objs")
		field := new(ast.ObjectFieldValue)
		if token.Kind == lex.NameToken {
			field.Name = token.Value
		} else {
			return nil, fmt.Errorf("unexpected token: %s", token.Value)
		}

		token = <-tokens
		if token.Kind != lex.PunctuatorToken && token.Value != ":" {
			return nil, fmt.Errorf("unexpected token: %s", token.Value)
		}

		token = <-tokens
		val, err := parseValue(token, tokens)
		if err != nil {
			return nil, err
		}
		field.Value = val
		o.Fields = append(o.Fields, field)

		token = <-tokens
		if token.Kind == lex.PunctuatorToken && token.Value == "}" {
			return o, nil
		}
	}
}

func parseFragments(tokens chan lex.Token) (token lex.Token, sel *ast.Selection, err error) {
	token = <-tokens
	sel = new(ast.Selection)
	if token.Kind == lex.NameToken && token.Value == "on" {
		inf := new(ast.InlineFragment)

		token = <-tokens
		if token.Kind == lex.NameToken {
			inf.TypeCondition = token.Value
			token = <-tokens
		}

		if token.Kind == lex.PunctuatorToken && token.Value == "@" {
			ds := []*ast.Directive{}
			token, ds, err = parseDirectives(tokens)
			if err != nil {
				return token, nil, err
			}
			inf.Directives = ds
		}

		if token.Kind == lex.PunctuatorToken && token.Value == "{" {
			sSet := []*ast.Selection{}
			token, sSet, err = parseSelectionSet(tokens)
			if err != nil {
				return token, nil, err
			}
			inf.SelectionSet = sSet
		} else {
			return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
		}
		sel.Kind = ast.InlineFragmentSelectionKind
		sel.InlineFragment = inf
		return token, sel, nil
	} else if token.Kind == lex.PunctuatorToken && token.Value == "{" {
		sSet := []*ast.Selection{}
		token, sSet, err = parseSelectionSet(tokens)
		if err != nil {
			return token, nil, err
		}
		inf := &ast.InlineFragment{
			SelectionSet: sSet,
		}

		sel.Kind = ast.InlineFragmentSelectionKind
		sel.InlineFragment = inf
		return token, sel, nil
	} else if token.Kind == lex.NameToken && token.Value != "on" {
		fs := new(ast.FragmentSpread)
		fs.Name = token.Value
		token = <-tokens

		if token.Kind == lex.PunctuatorToken && token.Value == "@" {
			ds := []*ast.Directive{}
			token, ds, err = parseDirectives(tokens)
			if err != nil {
				return token, nil, err
			}
			fs.Directives = ds
		}

		sel.Kind = ast.FragmentSpreadSelectionKind
		sel.FragmentSpread = fs
		return token, sel, nil
	}
	return token, nil, fmt.Errorf("unexpected token: %s", token.Value)
}

func parseDirectives(tokens chan lex.Token) (lex.Token, []*ast.Directive, error) {
	token := <-tokens
	ds := []*ast.Directive{}
	for {
		if token.Kind == lex.NameToken {
			d := new(ast.Directive)
			d.Name = token.Value
			token = <-tokens

			if token.Kind == lex.PunctuatorToken && token.Value == "(" {
				args, err := parseArguments(tokens)
				if err != nil {
					return token, nil, err
				}
				d.Arguments = args
				token = <-tokens
			}
			ds = append(ds, d)
		} else {
			return token, ds, nil
		}
	}
}
