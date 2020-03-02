package main

import (
	"context"
	"fmt"

	"github.com/rigglo/gql/schema"

	"github.com/rigglo/gql/gql"
)

func main() {
	res := BlockBusters.Exec(context.Background(), schema.ExecuteParams{
		OperationName: "",
		Query: `query {
			top_movie
		}`,
	})
	fmt.Printf("%+v", res)
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
				Type:        gql.String,
				Description: "",
				Resolver: func(ctx context.Context) (interface{}, error) {
					return "Interstellar", nil
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
