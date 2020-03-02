package main

import (
	"context"
	"encoding/json"
	"errors"
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
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return parent.(Movie).ID, errors.New("asd")
				},
			},
			&gql.Field{
				Name:        "title",
				Type:        gql.String,
				Description: "title of the movie",
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return parent.(Movie).Title, nil
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
