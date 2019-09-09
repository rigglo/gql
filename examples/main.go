package main

import (
	"context"
	"log"

	"github.com/rigglo/gql"
)

type Movie struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Rate     int    `json:"rate"`
}

var (
	MovieType = &gql.Object{
		Fields: gql.Fields{
			"title": &gql.Field{
				Type: gql.StringType,
			},
		},
	}
)

func main() {
	schema := new(gql.Schema)
	schema.RootQuery = &gql.Field{
		Fields: gql.Fields{
			"movies": &gql.Field{
				Type: MovieType,
				Resolver: func(ctx context.Context, args gql.Arguments, info gql.ResolverInfo) (interface{}, error) {
					return Movie{
						Title:    "Interstellar",
						Category: "Sci-fi",
						Rate:     4,
					}, nil
				},
			},
		},
	}

	res := schema.Do(context.Background(), `
	query {
		movies	{
			title
			rate
		}
	}`)
	log.Printf("%+v", res)
}
