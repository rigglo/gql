package federation

import (
	"github.com/rigglo/gql"
)

var (
	service = &gql.Object{
		Name: "_Service",
		Fields: gql.Fields{
			"sdl": &gql.Field{
				Type: gql.NewNonNull(gql.String),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return `
					type Query {
						pizzas: [Pizza]	
					}
					
					type Pizza {
						id: ID
						name: String
						size: Int
					}
					`, nil
				},
			},
		},
	}

	any = &gql.Scalar{
		Name:        "_Any",
		Description: "The `_Any` scalar is used to pass representations of entities from external services into the root _entities field for execution",
	}

	entity = &gql.Union{
		Name: "_Entity",
	}

	fieldset = &gql.Scalar{
		Name:        "_FieldSet",
		Description: "This is a custom scalar type that is used to represent a set of fields",
	}
)

func Federate(schema *gql.Schema) {
	schema.Query.Fields["_service"] = &gql.Field{
		Type: gql.NewNonNull(service),
		Resolver: func(ctx gql.Context) (interface{}, error) {
			return true, nil
		},
	}

	schema.Query.Fields["_entities"] = &gql.Field{
		Type: gql.NewNonNull(gql.NewList(entity)),
		Arguments: gql.Arguments{
			"representations": &gql.Argument{
				Type: gql.NewNonNull(gql.NewList(gql.NewNonNull(any))),
			},
		},
		Resolver: func(ctx gql.Context) (interface{}, error) {
			return nil, nil
		},
	}
}
