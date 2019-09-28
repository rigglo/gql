package gql_test

import (
	"context"
	"log"
	"testing"

	"github.com/rigglo/gql"
)

type Book struct {
	Title string `json:"title"`
	Rate  int    `json:"rate"`
}

var (
	BookType = &gql.Object{
		Name: "Book",
		Fields: gql.Fields{
			"title": &gql.Field{
				Name: "title",
				Type: gql.String,
			},
			"rate": &gql.Field{
				Name: "rate",
				Type: gql.Int,
			},
		},
	}
)

func TestArguments(t *testing.T) {
	schema := new(gql.Schema)
	schema.RootQuery = &gql.Field{
		Fields: gql.Fields{
			"book": &gql.Field{
				Name: "book",
				Type: BookType,
				Args: gql.Arguments{
					"category": &gql.Argument{
						Name:         "category",
						Type:         gql.String,
						DefaultValue: "Sci-fi",
					},
					"foo": &gql.Argument{
						Name:         "foo",
						Type:         gql.String,
						DefaultValue: "bar",
					},
				},
				Resolver: func(ctx context.Context, args gql.ResolverArgs, info gql.ResolverInfo) (interface{}, error) {
					log.Println(args.Get("category"))
					log.Println(args.Get("foo"))
					return Book{
						Title: "Interstellar",
						Rate:  4,
					}, nil
				},
			},
		},
	}

	res := schema.Do(context.Background(), `
	query {
		book(category: null)	{
			title
			rate
		}
	}`)
	t.Logf("%+v", res)
}
