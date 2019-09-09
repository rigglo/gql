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
			for _, f := range op.RootField.Fields {
				log.Println("'"+f.Name+"'", s.RootQuery.Fields)
				field, ok := s.RootQuery.Fields[f.Name]
				if !ok {
					return &Result{
						Errors: []error{
							fmt.Errorf("field '%s' is not defined", f.Name),
						},
					}
				}
				data, _ := field.Resolver(ctx, nil, ResolverInfo{
					//Path: []interface{}{f.Name},
				})
				log.Println(data)
				dataRes := map[string]interface{}{}
				for _, pf := range f.Fields {
					dv := reflect.ValueOf(data)
					for i := 0; i < dv.NumField(); i++ {
						if dv.Type().Field(i).Tag.Get("json") == pf.Name {
							log.Println(pf.Alias)
							dataRes[pf.Alias] = dv.Field(i).Interface()
						}
					}
				}
				res[f.Alias] = dataRes
			}
		}
	case "mutation":
		{

		}
	}

	/*
		bs, err := json.MarshalIndent(op.RootField, "", " ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(bs))
	*/
	return &Result{
		Data: res,
	}
}

func (s *Schema) resolveField(ctx context.Context, parent interface{}, lexField *parser.Field, field *Field) (interface{}, error) {
	if field.Type.FieldType() == ObjectType {
		data, _ := field.Resolver(ctx, nil, ResolverInfo{
			Path: []interface{}{},
		})

		for _, lf := range lexField.Fields {
			s.resolveField(ctx, data, lf, field.Fields[lf.Name])
		}
	}
	return nil, nil
}
