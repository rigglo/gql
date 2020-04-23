package main

import (
	"context"
	"log"
	"net/http"

	"github.com/rigglo/gql"
	"github.com/rigglo/gql/pkg/handler"
)

func main() {
	h := handler.New(handler.Config{
		Executor: gql.DefaultExecutor(Schema),
		GraphiQL: true,
	})
	http.Handle("/graphql", h)
	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Println(err)
	}
}

var (
	Schema = &gql.Schema{
		Query: Query,
	}

	Query = &gql.Object{
		Name: "Query",
		Fields: gql.Fields{
			"someErr": &gql.Field{
				Type:       gql.NewList(gql.String),
				Directives: gql.Directives{&myDirective{}},
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return nil, gql.NewError("asdasdasd", map[string]interface{}{"code": "SOME_ERR_CODE"})
				},
			},
		},
	}
)

type myDirective struct {
}

func (d *myDirective) GetName() string {
	return "myDirective"
}

func (d *myDirective) GetDescription() string {
	return "it's my directive"
}

func (d *myDirective) GetArguments() gql.Arguments {
	return gql.Arguments{}
}

func (d *myDirective) GetLocations() []gql.DirectiveLocation {
	return []gql.DirectiveLocation{gql.FieldDefinitionLoc}
}

func (d *myDirective) Variables() map[string]interface{} {
	return nil
}

func (d *myDirective) VisitFieldDefinition(ctx context.Context, f gql.Field, r gql.Resolver) gql.Resolver {
	return func(ctx gql.Context) (interface{}, error) {
		log.Println("my directive was called")
		v, err := r(ctx)
		if err != nil {
			return nil, nil
		}
		return v, nil
	}
}
