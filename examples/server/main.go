package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"

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

type User struct {
	Email string `json:"email"`
	Name  string `json:"name"`
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
				Type: CategoryType, /*
					Resolver: func(ctx context.Context, args gql.Arguments, info gql.ResolverInfo) (interface{}, error) {
						return Category{
							Name: "Sci-fi",
							Rank: 2,
						}, nil
					}, */
			},
		},
	}

	UserType = &gql.Object{
		Name: "User",
		Fields: gql.Fields{
			"name": &gql.Field{
				Name: "name",
				Type: gql.String,
			},
			"email": &gql.Field{
				Name: "email",
				Type: gql.String,
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
				Resolver: func(ctx context.Context, args gql.Arguments, info gql.ResolverInfo) (interface{}, error) {
					return rand.Intn(1000), nil
				},
			},
			"managed_by": &gql.Field{
				Name: "managed_by",
				Type: UserType,
				Resolver: func(ctx context.Context, args gql.Arguments, info gql.ResolverInfo) (interface{}, error) {
					return User{
						Name:  "John Doe",
						Email: "john.doe@rigglo.io",
					}, nil
				},
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
				managed_by {
					name
					email
				}
			}
		}
	}`)
	log.Printf("%+v", res)
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		req := struct {
			Query string `json:"query"`
		}{}
		bs, err := ioutil.ReadAll(r.Body)
		log.Println(err)

		err = json.Unmarshal(bs, &req)
		log.Println(err)

		res := schema.Do(context.Background(), req.Query)

		bs, err = json.MarshalIndent(res, "", " ")
		log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.Write(bs)
	})
	http.ListenAndServe(":8080", nil)
}
