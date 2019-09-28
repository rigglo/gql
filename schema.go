package gql

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/rigglo/gql/utils/lexer"
	"github.com/rigglo/gql/utils/lexer/tokens"
	"github.com/rigglo/gql/utils/parser"
)

// Schema ...
type Schema struct {
	RootQuery    *Field
	RootMutation *Field
}

const fieldTagName = "json"

type Map map[string]interface{}

func (s *Schema) Do(ctx context.Context, query string) *Result {
	op := parser.Parse(&lexer.Lexer{
		Input:  query,
		State:  lexer.LexSearch,
		Tokens: make(chan tokens.Token, 3),
	})
	res := map[string]interface{}{}
	switch op.RootField.Name {
	case "query":
		{
			for _, lexField := range op.RootField.Fields {
				log.Println("'"+lexField.Name+"'", s.RootQuery.Fields)
				field, ok := s.RootQuery.Fields[lexField.Name]
				if !ok {
					return &Result{
						Errors: []error{
							fmt.Errorf("field '%s' is not defined", lexField.Name),
						},
					}
				}
				log.Println("type", field.Type.Type())
				_ = s.ResolveField(ctx, []interface{}{lexField.Name}, nil, lexField, field, res)
				log.Printf("res: %+v", res)
			}
		}
	case "mutation":
		{

		}
	}
	return &Result{
		Data: res,
	}
}

func (s *Schema) ResolveField(ctx context.Context, path []interface{}, parent interface{}, lexField *parser.Field, field *Field, out map[string]interface{}) error {
	log.Println("---------")
	log.Printf("field: %+v", field)
	log.Printf("field.Type.Type(): %+v", field.Type.Type())
	if field.Type.Type() == ObjectType {
		resolverFunc := defaultObjectResolver
		if field.Resolver != nil {
			resolverFunc = field.Resolver
		}
		args, err := fromParsedArguments(field.Args, lexField.Args)
		if err != nil {
			return err
		}
		data, _ := resolverFunc(ctx, args, ResolverInfo{
			Path:   path,
			Parent: parent,
			Field:  *field,
		})
		o := map[string]interface{}{}

		for _, lf := range lexField.Fields {
			tpath := append(path, lf.Name)
			s.ResolveField(ctx, tpath, data, lf, field.Type.Value().(*Object).Fields[lf.Name], o)
		}
		out[lexField.Alias] = o
	} else if field.Type.Type() == ScalarType {
		if field.Resolver != nil {
			data, _ := field.Resolver(ctx, nil, ResolverInfo{
				Path:   path,
				Parent: parent,
				Field:  *field,
			})
			res, _ := field.Type.Value().(*Scalar).Encode(data)
			out[func() string {
				if lexField.Alias == "" {
					return lexField.Name
				}
				return lexField.Alias
			}()] = res
		} else {
			dv := reflect.ValueOf(parent)
			for i := 0; i < dv.NumField(); i++ {
				if dv.Type().Field(i).Tag.Get(fieldTagName) == field.Name {
					res, _ := field.Type.Value().(*Scalar).Encode(dv.Field(i).Interface())
					out[func() string {
						if lexField.Alias == "" {
							return lexField.Name
						}
						return lexField.Alias
					}()] = res
				}
			}
		}
		return nil
	}
	return nil
}
