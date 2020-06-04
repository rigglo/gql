package main

import (
	"net/http"

	"github.com/rigglo/gql"
	"github.com/rigglo/gql/pkg/handler"
)

type Pizza struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Size    int    `json:"size"`
}

var PizzaType = &gql.Object{
	Name: "Pizza",
	Fields: gql.Fields{
		"id": &gql.Field{
			Description: "id of the pizza",
			Type:        gql.ID,
		},
		"name": &gql.Field{
			Description: "name of the pizza",
			Type:        gql.String,
		},
		"size": &gql.Field{
			Description: "size of the pizza (in cm)",
			Type:        gql.Int,
		},
	},
}

var RootQuery = &gql.Object{
	Name: "RootQuery",
	Fields: gql.Fields{
		"pizzas": &gql.Field{
			Description: "lists all the pizzas",
			Type:        gql.NewList(PizzaType),
			Resolver: func(ctx gql.Context) (interface{}, error) {
				return []Pizza{
					Pizza{
						ID:   1,
						Name: "Veggie",
						Size: 32,
					},
					Pizza{
						ID:   2,
						Name: "Salumi",
						Size: 45,
					},
				}, nil
			},
		},
	},
}

var PizzeriaSchema = &gql.Schema{
	Query: RootQuery,
}

func main() {
	http.Handle("/graphql", handler.New(handler.Config{
		Executor:   gql.DefaultExecutor(PizzeriaSchema),
		Playground: true,
	}))
	if err := http.ListenAndServe(":9999", nil); err != nil {
		panic(err)
	}
}
