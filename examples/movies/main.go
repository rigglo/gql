package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rigglo/gql"
	"io/ioutil"
	"log"
	"net/http"
)

type (
	Movie struct {
		ID       string
		Title    string
		Category int
	}
)

var (
	movies = []*Movie{
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
	}

	CategoryEnum = &gql.Enum{
		Name: "Categories",
		Values: gql.EnumValues{
			&gql.EnumValue{
				Value: 3,
				Name:  "SCIFI",
			},
			&gql.EnumValue{
				Value: 4,
				Name:  "ACTION",
			},
			&gql.EnumValue{
				Value: 6,
				Name:  "ROMANTIC",
			},
		},
	}

	MovieType = &gql.Object{
		Name:        "Movie",
		Description: "This is a record of a Movie",
		Fields: gql.NewFields(
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
				Type:        CategoryEnum,
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return parent.(*Movie).Category, nil
				},
			},
		),
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

	AddMovieMutation = &gql.Field{
		Name:        "addMovie",
		Description: "This is a record of a Movie",
		Type:        gql.NewList(MovieType),
		Arguments: gql.NewArguments(
			&gql.Argument{
				Name: "title",
				Type: gql.NewNonNull(gql.String),
			},
			&gql.Argument{
				Name: "category",
				Type: CategoryEnum,
			},
		),
		Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
			movies = append(movies, &Movie{
				ID:       uuid.New().String(),
				Title:    args["title"].(string),
				Category: args["category"].(int),
			})
			log.Println(args)
			return movies, nil
		},
	}
)

func main() {
	schema := &gql.Schema{
		Query: &gql.Object{
			Name: "Query",
			Fields: gql.NewFields(
				MoviesQuery,
			),
		},
		Mutation: &gql.Object{
			Name: "Mutation",
			Fields: gql.NewFields(
				AddMovieMutation,
			),
		},
	}

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "couldn't read body", http.StatusInternalServerError)
		}
		params := gql.Params{}
		err = json.Unmarshal(bs, &params)
		if err != nil {
			http.Error(w, "couldn't read body", http.StatusInternalServerError)
		}
		log.Printf("params: %+v", params)
		result := schema.Execute(&params)
		log.Printf("result: %+v", result)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
	fmt.Println("Now server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
