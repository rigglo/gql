package gql_test

import (
	"context"
	"testing"

	"github.com/rigglo/gql"
)

type Movie struct {
	Title    string   `json:"title"`
	Category Category `json:"category"`
	Rate     int      `json:"rate"`
}

type Category struct {
	Name string `json:"name"`
	Rank int    `json:"rank"`
}

var (
	MovieType = &gql.Object{
		Name: "Movie",
		Fields: gql.Fields{
			"title": &gql.Field{
				Name: "title",
				Type: gql.String,
			},
			"rate": &gql.Field{
				Name: "rate",
				Type: gql.Int,
			},
			"category": &gql.Field{
				Name: "category",
				Type: CategoryType,
				Resolver: func(ctx context.Context, args gql.Arguments, info gql.ResolverInfo) (interface{}, error) {
					return Category{
						Name: "Sci-fi",
						Rank: 2,
					}, nil
				},
			},
		},
	}

	CategoryType = &gql.Object{
		Name: "Category",
		Fields: gql.Fields{
			"name": &gql.Field{
				Name: "name",
				Type: gql.String,
			},
			"rank": &gql.Field{
				Name: "rank",
				Type: gql.Int,
			},
		},
	}
)

func TestSchema(t *testing.T) {
	schema := new(gql.Schema)
	schema.RootQuery = &gql.Field{
		Fields: gql.Fields{
			"movies": &gql.Field{
				Name: "movies",
				Type: MovieType,
				Resolver: func(ctx context.Context, args gql.Arguments, info gql.ResolverInfo) (interface{}, error) {
					return Movie{
						Title: "Interstellar",
						Category: Category{
							Name: "Sci-fi",
							Rank: 2,
						},
						Rate: 4,
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
			category {
				name
				rank
			}
		}
	}`)
	t.Logf("%+v", res)
}
