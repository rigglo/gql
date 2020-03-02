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
		Query: `query {
			top_movie {
				id
			}
			foo
		}`,
	})
	bs, _ := json.Marshal(res)
	fmt.Printf("%s\n", string(bs))
	//fmt.Printf("%+v", res)
}

type Movie struct {
	ID    string `gql:"id"`
	Title string `gql:"title"`
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
				Resolver: func(ctx context.Context) (interface{}, error) {
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
				Resolver: func(ctx context.Context) (interface{}, error) {
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
				Resolver: func(ctx context.Context) (interface{}, error) {
					return "some id", nil
				},
			},
			&gql.Field{
				Name:        "title",
				Type:        gql.String,
				Description: "title of the movie",
				Resolver: func(ctx context.Context) (interface{}, error) {
					return "title......", nil
				},
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
