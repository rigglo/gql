package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/rigglo/gql"
	"github.com/rigglo/gql/pkg/handler"
)

func main() {
	h := handler.New(handler.Config{
		Executor:   gql.DefaultExecutor(BlockBusters),
		Playground: true,
	})
	http.Handle("/graphql", h)
	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Println(err)
	}
}

type Movie struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

var (
	BlockBusters = &gql.Schema{
		Query:    Query,
		Mutation: nil,
		AdditionalTypes: []gql.Type{
			FooInput,
		},
	}

	FooInput = &gql.InputObject{
		Name: "FooInput",
		Fields: gql.InputFields{
			"asd": &gql.InputField{
				Type:         gql.NewNonNull(gql.String),
				DefaultValue: "bar",
			},
		},
	}

	Query = &gql.Object{
		Name:        "Query",
		Description: "Just the blockbusters root query",
		Fields: gql.Fields{
			"top_movies": &gql.Field{
				Type:        gql.NewNonNull(gql.NewList(MovieType)),
				Description: "",
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return []Movie{
						{
							ID:    "22424234",
							Title: "Interstellar",
						},
						{
							ID:    "34363453",
							Title: "Titanic",
						},
					}, nil
				},
			},
			"foo": &gql.Field{
				Type:        gql.String,
				Description: "",
				Arguments: gql.Arguments{
					"asd": &gql.Argument{
						Type: gql.NewNonNull(FooInput),
					},
					"bar": &gql.Argument{
						Type: gql.String,
					},
				},
				Resolver: func(ctx gql.Context) (interface{}, error) {
					log.Printf("%v", ctx.Args())
					return ctx.Args()["asd"].(map[string]interface{})["asd"], nil
				},
			},
		},
	}

	MovieType = &gql.Object{
		Name:        "Movie",
		Description: "This is a movie from the BlockBusters",
		Fields: gql.Fields{
			"id": &gql.Field{
				Type:        gql.String,
				Description: "id of the movie",
				Arguments: gql.Arguments{
					"foo": &gql.Argument{
						Type: gql.String,
					},
				},
			},
			"title": &gql.Field{
				Type:        gql.String,
				Description: "title of the movie",
			},
			"name": &gql.Field{
				Type:        gql.NewNonNull(gql.String),
				Description: "name of the movie",
				Directives: gql.TypeSystemDirectives{
					gql.Deprecate("It's just not implemented and movies have titles, not names"),
				},
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return "asd", errors.New("something bad happened")
				},
			},
		},
	}
)
