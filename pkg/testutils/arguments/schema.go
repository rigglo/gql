package arguments

import (
	"github.com/rigglo/gql/pkg/gql"
)

var (
	Arguments = &gql.Object{
		Name: "Arguments",
		Fields: gql.Fields{
			&gql.Field{
				Name: "multipleReqs",
				Type: gql.NewNonNull(gql.Int),
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "x",
						Type: gql.NewNonNull(gql.Int),
					},
					&gql.Argument{
						Name: "y",
						Type: gql.NewNonNull(gql.Int),
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return 42, nil
				},
			},
			&gql.Field{
				Name: "booleanArgField",
				Type: gql.Boolean,
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "booleanArg",
						Type: gql.Boolean,
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return true, nil
				},
			},
			&gql.Field{
				Name: "floatArgField",
				Type: gql.Float,
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "floatArg",
						Type: gql.Float,
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return 0.42, nil
				},
			},
			&gql.Field{
				Name: "intArgField",
				Type: gql.Int,
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "intArg",
						Type: gql.Int,
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return 42, nil
				},
			},
			&gql.Field{
				Name: "nonNullBooleanArgField",
				Type: gql.NewNonNull(gql.Boolean),
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "nonNullBooleanArg",
						Type: gql.NewNonNull(gql.Boolean),
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return true, nil
				},
			},
			&gql.Field{
				Name: "booleanListArgField",
				Type: gql.NewList(gql.Boolean),
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "booleanListArg",
						Type: gql.NewNonNull(gql.NewList(gql.Boolean)),
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return 42, nil
				},
			},
			&gql.Field{
				Name: "optionalNonNullBooleanArgField",
				Type: gql.Int,
				Arguments: gql.Arguments{
					&gql.Argument{
						Name:         "optionalBooleanArg",
						Type:         gql.NewNonNull(gql.Boolean),
						DefaultValue: false,
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return true, nil
				},
			},
		},
	}

	ArgumentsSchema = &gql.Schema{
		Query: &gql.Object{
			Name: "RootQuery",
			Fields: gql.Fields{
				&gql.Field{
					Name: "arguments",
					Type: Arguments,
					Resolver: func(c gql.Context) (interface{}, error) {
						return true, nil
					},
				},
			},
		},
	}
)
