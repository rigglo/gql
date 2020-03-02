package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rigglo/gql/schema"

	"github.com/rigglo/gql/gql"
)

func main() {
	res := BlockBusters.Exec(context.Background(), schema.ExecuteParams{
		OperationName: "",
		Query: `
		query {
			top_movie {
				id
				title
			}
			foo
		}`,
	})
	bs, _ := json.Marshal(res)
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

	Query = &gql.Object{
		Name:        "Query",
		Description: "Just the blockbusters root query",
		Fields: gql.Fields{
			&gql.Field{
				Name:        "top_movie",
				Type:        MovieType,
				Description: "",
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return Movie{
						ID:    "22424234",
						Title: "Interstellar",
					}, nil
				},
			},
			&gql.Field{
				Name:        "foo",
				Type:        gql.String,
				Description: "",
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return "bar", nil
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
