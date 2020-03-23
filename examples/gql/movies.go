package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/rigglo/gql/pkg/gql"
)

func main() {
	res := BlockBusters.Exec(context.Background(), gql.Params{
		OperationName: "",
		Query: `
		query {
			top_movies {
				id
				title
			}
			foo(asd: {asd: "bar"}, bar: "foo")
		}`,
		Variables: map[string]interface{}{
			"fromvar": map[string]interface{}{
				"asd": "barvar",
			},
		},
	})
	bs, _ := json.MarshalIndent(res, "", "  ")
	fmt.Printf("%s\n", string(bs))
	//fmt.Printf("%+v", res)
}

type Movie struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

var (
	BlockBusters = &gql.Schema{
		Query: Query,
	}

	FooInput = &gql.InputObject{
		Name: "FooInput",
		Fields: gql.InputFields{
			"asd": &gql.InputField{
				Type: gql.String,
			},
		},
	}

	Query = &gql.Object{
		Name:        "Query",
		Description: "Just the blockbusters root query",
		Fields: gql.Fields{
			&gql.Field{
				Name:        "top_movies",
				Type:        gql.NewList(MovieType),
				Description: "",
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return []Movie{
						Movie{
							ID:    "22424234",
							Title: "Interstellar",
						},
						Movie{
							ID:    "34363453",
							Title: "Titanic",
						},
					}, nil
				},
			},
			&gql.Field{
				Name:        "foo",
				Type:        gql.String,
				Description: "",
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "asd",
						Type: FooInput,
					},
					&gql.Argument{
						Name: "bar",
						Type: gql.String,
					},
				},
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					log.Println(args)
					return args["asd"].(map[string]interface{})["asd"], nil
				},
			},
		},
	}

	MovieType = &gql.Object{
		Name:        "Movie",
		Description: "This is a movie from the BlockBusters",
		Fields: gql.Fields{
			&gql.Field{
				Name:        "id",
				Type:        gql.String,
				Description: "id of the movie",
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "foo",
						Type: gql.String,
					},
				},
			},
			&gql.Field{
				Name:        "title",
				Type:        gql.String,
				Description: "title of the movie",
			},
			&gql.Field{
				Name: "name",
				Type: gql.String,
				Directives: gql.Directives{
					gql.DepricatiedDirective,
				},
				Description: "name of the movie",
			},
		},
	}
)
