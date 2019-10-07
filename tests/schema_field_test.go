package tests

import (
	"context"
	"encoding/json"
	"github.com/rigglo/gql"
	"github.com/rigglo/gql/schema"
	"log"
	"testing"
)

type (
	Movie struct {
		ID       string
		Title    string
		Category int
	}
)

var (
	CategoryEnum = &schema.Enum{
		Name: "Categories",
		Values: schema.EnumValues{
			&schema.EnumValue{
				Value: 3,
				Name:  "SCIFI",
			},
			&schema.EnumValue{
				Value: 4,
				Name:  "ACTION",
			},
		},
	}

	MovieType = &schema.Object{
		Name:        "Movie",
		Description: "This is a record of a Movie",
		Fields: []*schema.Field{
			&schema.Field{
				Name:        "id",
				Description: "this is the id of a move",
				Type:        schema.ID,
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return parent.(*Movie).ID, nil
				},
			},
			&schema.Field{
				Name:        "title",
				Description: "title of a move",
				Type:        schema.String,
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return parent.(*Movie).Title, nil
				},
			},
			&schema.Field{
				Name:        "category",
				Description: "category of a move",
				Type:        CategoryEnum,
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return parent.(*Movie).Category, nil
				},
			},
		},
	}

	MoviesQuery = &schema.Field{
		Name:        "movies",
		Description: "This is a record of a Movie",
		Type:        schema.NewList(MovieType),
		Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
			return []*Movie{
				&Movie{
					ID:       "91902d98-f1a0-473a-9ff1-74b7b59b0ffa",
					Title:    "Interstellar",
					Category: 3,
				},
				&Movie{
					ID:       "6d122f45-77a4-41ca-a7a7-fcb20853684d",
					Title:    "Avatar",
					Category: 4,
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
		title
	}
	movies {
		category
	}
}

query GetMoviesWithFragments {
	movies {
		...IdAndTitle
	}
}
	`

	schema := schema.Schema{
		Query: &schema.Object{
			Name: "Query",
			Fields: schema.Fields{
				MoviesQuery,
			},
		},
	}

	res := gql.Execute(&gql.Params{
		Ctx:           context.Background(),
		Schema:        schema,
		OperationName: "GetMovies",
		Query:         query,
		Variables:     map[string]interface{}{},
	})
	if len(res.Errors) != 0 {
		t.Fatal(res.Errors)
	} else if res.Data == nil {
		t.Fatal("res is nil")
	}
	bs, err := json.MarshalIndent(res, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	log.Println("'", string(bs), "'")
}
