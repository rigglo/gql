/*
Package gql is a GraphQL implementation

NOTE: THIS PACKAGE IS STILL UNDER DEVELOPMENT

About gql

The gql package aims to fulfill a few extra features that I missed in other packages,
be as efficient as possible and 100% close to the GraphQL specificaton. The idea behind
the package is to achive a really easy learning curve, but still leave many option to customize
everything to your benefit.

Just a few features that are already built in
  - Custom scalars
  - request validation (still in progress)
  - handler

Getting started

We'll build our own pizzeria, for that we need a `Pizza` type to store the properties of the pizzas.
At the beginning, we only store the name, size and if the pizza is spicy.

	var PizzaType = &gql.Object{
		Name: "Pizza",
		Fields: gql.Fields{
			&gql.Field{
				Name: "name",
				Description: "name of the pizza",
				Type: gql.String,
			},
			&gql.Field{
				Name: "size",
				Description: "size of the pizza (in cm)",
				Type: gql.Int,
			},
			&gql.Field{
				Name: "is_spicy",
				Description: "is the pizza spicy or not",
				Type: gql.Boolean,
			},
		},
	}

Since we have a type, we need to build a query to be able to access to it. The root query will be
the starting point of any query, fields in this type represents be all the available query fields.
Let's add the root query with a "pizzas" query, which will return a list of the available pizzas.
Since we need a little logic here to return the available pizzas, we'll use a resolver.

	var RootQuery = &gql.Object{
		Name: "Query",
		Fields: gql.Fields{
			&gql.Field{
				Name: "pizzas",
				Type: gql.NewList(PizzaType),
				Resolver: func(c gql.Context) (interface{}, error) {
					return []Pizza{
						Pizza{
							Name: "songoku",
							Size: 32,
							IsSpicy: false,
						},
						Pizza{
							Name: "hell-o",
							Size: 40,
							IsSpicy: true,
						},
					}, nil
				},
			},
		},
	}

*/
package gql
