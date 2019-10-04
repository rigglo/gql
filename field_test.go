package gql_test

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/rigglo/gql"
)

type (
	Movie struct {
		ID       string
		Title    string
		Category string
	}
)

var (
	MovieType = &gql.Object{
		Name:        "Movie",
		Description: "This is a record of a Movie",
		Fields: []*gql.Field{
			&gql.Field{
				Name:        "id",
				Description: "this is the id of a move",
				Type:        gql.ID,
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return parent.(*Movie).ID, nil
				},
			},
			&gql.Field{
				Name:        "title",
				Description: "title of a move",
				Type:        gql.String,
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return parent.(*Movie).Title, nil
				},
			},
			&gql.Field{
				Name:        "category",
				Description: "category of a move",
				Type:        gql.String,
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return parent.(*Movie).Category, nil
				},
			},
		},
	}

	MoviesQuery = &gql.Field{
		Name:        "movies",
		Description: "This is a record of a Movie",
		Type:        gql.NewList(MovieType),
		Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
			return []*Movie{
				&Movie{
					ID:       "91902d98-f1a0-473a-9ff1-74b7b59b0ffa",
					Title:    "Interstellar",
					Category: "Sci-fi",
				},
				&Movie{
					ID:       "6d122f45-77a4-41ca-a7a7-fcb20853684d",
					Title:    "Avatar",
					Category: "Sci-fi",
				},
			}, nil
		},
	}
)

func TestFieldResolver(t *testing.T) {
	query := `
fragment IdAndTitle on Movie {
	id
	title
}

query GetMovies {
	movies {
		category
		title
		id
	}
}

query GetMoviesWithFragments {
	movies {
		...IdAndTitle
	}
}
	`

	schema := &gql.Schema{
		Query: &gql.Object{
			Name: "Query",
			Fields: gql.Fields{
				MoviesQuery,
			},
		},
	}
	res := schema.Execute(context.Background(), query, "GetMoviesWithFragments", map[string]interface{}{})
	//spew.Dump(res)
	bs, _ := json.MarshalIndent(res, "", "\t")
	log.Println(string(bs))
}
