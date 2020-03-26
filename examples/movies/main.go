package main

import (
	"context"
	"log"
	"net/http"

	"github.com/rigglo/gql/pkg/handler"

	"github.com/rigglo/gql/pkg/gql"
)

func main() {
	h := handler.New(handler.Config{
		Schema: BlockBusters,
	})
	/* res := BlockBusters.Exec(context.Background(), gql.Params{
		OperationName: "introspection",
		Query: `
		query introspection {
			__schema {
				queryType {
					name
				}
			}
		}
		query myOp($barVar: String = "foo") {
			top_movies {
				__typename
				... {
					title
					... {
						id
					}
				}
			}
			foo(asd: {asd: "bar"}, bar: $barVar)
		}`,
		Variables: "",
	})
	bs, _ := json.MarshalIndent(res, "", "  ")
	fmt.Printf("%s\n", string(bs)) */
	//fmt.Printf("%+v", res)
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
		Mutation: Query,
	}

	FooInput = &gql.InputObject{
		Name: "FooInput",
		Fields: gql.InputFields{
			"asd": &gql.InputField{
				Type:         gql.String,
				DefaultValue: "defoocska",
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
				Name:        "name",
				Type:        gql.String,
				Directives:  gql.Directives{},
				Description: "name of the movie",
			},
		},
	}
)
