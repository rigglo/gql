package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/rigglo/gql"
)

type Movie struct {
	ID       int      `json:"id"`
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
			"id": &gql.Field{
				Name: "id",
				Type: gql.Int,
			},
			"title": &gql.Field{
				Name: "title",
				Type: gql.String,
			},
			"rate": &gql.Field{
				Name: "rate",
				Type: gql.Int,
			},
			"now": &gql.Field{
				Name: "now",
				Type: gql.UnixTime,
				Resolver: func(ctx context.Context, args gql.ResolverArgs, info gql.ResolverInfo) (interface{}, error) {
					return time.Now(), nil
				},
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
				Resolver: func(ctx context.Context, args gql.ResolverArgs, info gql.ResolverInfo) (interface{}, error) {
					return rand.Intn(1000), nil
				},
			},
			"managed_by": &gql.Field{
				Name: "managed_by",
				Type: UserType,
				Resolver: func(ctx context.Context, args gql.ResolverArgs, info gql.ResolverInfo) (interface{}, error) {
					return User{
						Name:  "John Doe",
						Email: "john.doe@rigglo.io",
					}, nil
				},
			},
		},
	}
)
var movies = []*Movie{
	31: &Movie{},
}

func main() {
	schema := new(gql.Schema)
	schema.RootQuery = &gql.Field{
		Fields: gql.Fields{
			"movie": &gql.Field{
				Type: MovieType,
				Args: gql.Arguments{
					"id": &gql.Argument{
						Name: "id",
						Type: gql.NewNonNull(gql.Int),
					},
				},
				Resolver: func(ctx context.Context, args gql.ResolverArgs, info gql.ResolverInfo) (interface{}, error) {
					id, _ := args.Get("id")
					return Movie{
						ID:    id.(int),
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
		movie {
			title
			rate
			now
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
