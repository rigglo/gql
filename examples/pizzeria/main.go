package main

import (
	"net/http"

	"github.com/rigglo/gql"
	"github.com/rigglo/gql/pkg/handler"
)

func main() {
	http.Handle("/graphql", handler.New(handler.Config{
		Executor: gql.DefaultExecutor(Schema),
		GraphiQL: true,
	}))
	if err := http.ListenAndServe(":9999", nil); err != nil {
		panic(err)
	}
}

type Pizza struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Size    int    `json:"size"`
	IsSpicy bool   `json:"is_spicy"`
}

var pizzas = []Pizza{
	Pizza{
		ID:      1,
		Name:    "songoku",
		Size:    32,
		IsSpicy: false,
	},
	Pizza{
		ID:      2,
		Name:    "hell-o",
		Size:    40,
		IsSpicy: true,
	},
}

var (
	PizzaType = &gql.Object{
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
			"is_spicy": &gql.Field{
				Description: "is the pizza spicy or not",
				Type:        gql.Boolean,
			},
		},
	}

	RootQuery = &gql.Object{
		Name: "Query",
		Fields: gql.Fields{
			"pizzas": &gql.Field{
				Type: gql.NewList(PizzaType),
				Resolver: func(c gql.Context) (interface{}, error) {
					return pizzas, nil
				},
			},
		},
	}

	PizzaInput = &gql.InputObject{
		Name: "PizzaInput",
		Fields: gql.InputFields{
			"name": &gql.InputField{
				Type: gql.NewNonNull(gql.String),
			},
			"size": &gql.InputField{
				Type: gql.NewNonNull(gql.Int),
			},
			"is_spicy": &gql.InputField{
				Type: gql.NewNonNull(gql.Boolean),
			},
		},
	}

	RootMutation = &gql.Object{
		Name: "Mutation",
		Fields: gql.Fields{
			"addPizza": &gql.Field{
				Type: gql.NewNonNull(PizzaType),
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "input",
						Type: gql.NewNonNull(PizzaInput),
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					pizza := c.Args()["input"].(map[string]interface{})
					p := Pizza{
						ID:      len(pizzas) + 1,
						Name:    pizza["name"].(string),
						Size:    pizza["size"].(int),
						IsSpicy: pizza["is_spicy"].(bool),
					}
					pizzas = append(pizzas, p)
					return p, nil
				},
			},
		},
	}

	Schema = &gql.Schema{
		Query:    RootQuery,
		Mutation: RootMutation,
	}
)
