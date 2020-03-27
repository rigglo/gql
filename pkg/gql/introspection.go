package gql

import (
	"context"
)

func init() {
	typeIntrospection.AddFields(
		&Field{
			Name: "interfaces",
			Type: NewList(NewNonNull(typeIntrospection)),
			Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
				return []interface{}{}, nil
			},
		},
		&Field{
			Name: "possibleTypes",
			Type: NewList(NewNonNull(typeIntrospection)),
			Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
				return []interface{}{}, nil
			},
		},
		&Field{
			Name: "ofType",
			Type: typeIntrospection,
			Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
				t := parent.(Type)
				switch t.GetKind() {
				case NonNullKind:
					return t.(WrappingType).Unwrap(), nil
				case ListKind:
					return t.(WrappingType).Unwrap(), nil
				}
				return nil, nil
			},
		},
	)

	inputValueIntrospection.AddFields(
		&Field{
			Name: "type",
			Type: NewNonNull(typeIntrospection),
			Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
				if v, ok := parent.(*Argument); ok {
					return v.GetType(), nil
				}
				if v, ok := parent.(*InputField); ok {
					return v.GetType(), nil
				}
				return nil, nil
			},
		},
	)

	fieldIntrospection.AddFields(
		&Field{
			Name: "type",
			Type: NewNonNull(typeIntrospection),
			Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
				if v, ok := parent.(*Field); ok {
					return v.Type, nil
				}
				return nil, nil
			},
		},
	)
}

func addIntrospectionTypes(types map[string]Type) {
	types["__Schema"] = schemaIntrospection
	types["__Type"] = typeIntrospection
	types["__Field"] = fieldIntrospection
	types["__InputValue"] = inputValueIntrospection
	types["__EnumValue"] = enumValueIntrospection
	types["__TypeKind"] = typeKindIntrospection
	types["__Directive"] = directiveIntrospection
	types["__DirectiveLocation"] = directiveLocationIntrospection
}

var (
	introspectionQuery = &Object{
		Fields: Fields{
			&Field{
				Name: "__schema",
				Type: NewNonNull(schemaIntrospection),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return nil, nil
				},
			},
			&Field{
				Name: "__type",
				Type: typeIntrospection,
				Arguments: Arguments{
					&Argument{
						Name: "name",
						Type: NewNonNull(String),
					},
				},
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					ectx := ctx.(*eCtx)
					ts := ectx.Get(keyTypes).(map[string]Type)
					return ts[args["name"].(string)], nil
				},
			},
		},
	}

	schemaIntrospection = &Object{
		Name: "__Schema",
		Fields: Fields{
			&Field{
				Name: "types",
				Type: NewNonNull(NewList(NewNonNull(typeIntrospection))),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					ectx := ctx.(*eCtx)
					ts := ectx.Get(keyTypes).(map[string]Type)

					out := []Type{}
					for _, t := range ts {
						out = append(out, t)
					}

					return out, nil
				},
			},
			&Field{
				Name: "queryType",
				Type: NewNonNull(typeIntrospection),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					ectx := ctx.(*eCtx)
					schema := ectx.Get(keySchema).(*Schema)
					return schema.GetRootQuery(), nil
				},
			},
			&Field{
				Name: "mutationType",
				Type: typeIntrospection,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return nil, nil
				},
			},
			&Field{
				Name: "subscriptionType",
				Type: typeIntrospection,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return nil, nil
				},
			},
			&Field{
				Name: "directives",
				Type: NewNonNull(NewList(NewNonNull(directiveIntrospection))),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					// ectx := ctx.(*eCtx)
					// schema := ectx.Get(keySchema).(*Schema)
					return []interface{}{}, nil
				},
			},
		},
	}

	typeIntrospection = &Object{
		Name: "__Type",
		Fields: Fields{
			&Field{
				Name: "kind",
				Type: NewNonNull(typeKindIntrospection),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					switch parent.(Type).GetKind() {
					case ScalarKind:
						return "SCALAR", nil
					case ObjectKind:
						return "OBJECT", nil
					case InterfaceKind:
						return "INTERFACE", nil
					case UnionKind:
						return "UNION", nil
					case EnumKind:
						return "ENUM", nil
					case InputObjectKind:
						return "INPUT_OBJECT", nil
					case ListKind:
						return "LIST", nil
					case NonNullKind:
						return "NON_NULL", nil
					}
					return "OBJECT", nil
				},
			},
			&Field{
				Name: "name",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					// ectx := ctx.(*eCtx)
					// schema := ectx.Get(keySchema).(*Schema)
					return parent.(Type).GetName(), nil
				},
			},
			&Field{
				Name: "description",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					// ectx := ctx.(*eCtx)
					// schema := ectx.Get(keySchema).(*Schema)
					return parent.(Type).GetDescription(), nil
				},
			},
			&Field{
				Name: "fields",
				Arguments: Arguments{
					&Argument{
						Name:         "includeDeprecated",
						Type:         Boolean,
						DefaultValue: false,
					},
				},
				Type: NewList(NewNonNull(fieldIntrospection)),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if parent.(Type).GetKind() == ObjectKind {
						return parent.(*Object).GetFields(), nil
					}
					if parent.(Type).GetKind() == InterfaceKind {
						return parent.(*Interface).GetFields(), nil
					}
					return nil, nil
				},
			},
			&Field{
				Name: "enumValues",
				Arguments: Arguments{
					&Argument{
						Name:         "includeDeprecated",
						Type:         Boolean,
						DefaultValue: false,
					},
				},
				Type: NewList(NewNonNull(enumValueIntrospection)),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if parent.(Type).GetKind() == EnumKind {
						return parent.(*Enum).GetValues(), nil
					}
					return nil, nil
				},
			},
			&Field{
				Name: "inputFields",
				Type: NewList(NewNonNull(inputValueIntrospection)),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if parent.(Type).GetKind() == InputObjectKind {
						fs := []*InputField{}
						for _, f := range parent.(*InputObject).GetFields() {
							fs = append(fs, f)
						}
						return fs, nil
					}
					return nil, nil
				},
			},
		},
	}

	fieldIntrospection = &Object{
		Name: "__Field",
		Fields: Fields{
			&Field{
				Name: "name",
				Type: NewNonNull(String),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if f, ok := parent.(*Field); ok {
						return f.GetName(), nil
					}
					return nil, nil
				},
			},
			&Field{
				Name: "description",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if f, ok := parent.(*Field); ok {
						return f.GetDescription(), nil
					}
					return nil, nil
				},
			},
			&Field{
				Name: "args",
				Type: NewNonNull(NewList(NewNonNull(inputValueIntrospection))),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if f, ok := parent.(*Field); ok && f.Arguments != nil {
						return f.GetArguments(), nil
					}
					return []interface{}{}, nil
				},
			},
			&Field{
				Name: "isDeprecated",
				Type: NewNonNull(Boolean),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					/* if f, ok := parent.(*Field); ok {
						return f.GetName(), nil
					} */
					return false, nil
				},
			},
			&Field{
				Name: "deprecationReason",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					/* if f, ok := parent.(*Field); ok {
						return f.GetName(), nil
					} */
					return nil, nil
				},
			},
		},
	}

	inputValueIntrospection = &Object{
		Name: "__InputValue",
		Fields: Fields{
			&Field{
				Name: "name",
				Type: NewNonNull(String),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if v, ok := parent.(*EnumValue); ok {
						return v.Name, nil
					} else if v, ok := parent.(*Argument); ok {
						return v.Name, nil
					} else if v, ok := parent.(*Scalar); ok {
						return v.Name, nil
					} else if v, ok := parent.(*InputField); ok {
						return v.Name, nil
					}
					return nil, nil
				},
			},
			&Field{
				Name: "description",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if f, ok := parent.(*Field); ok {
						return f.GetName(), nil
					}
					return nil, nil
				},
			},
			&Field{
				Name: "defaultValue",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					/* if f, ok := parent.(*Field); ok {
						return f.GetName(), nil
					} */
					return nil, nil
				},
			},
		},
	}

	enumValueIntrospection = &Object{
		Name: "__EnumValue",
		Fields: Fields{
			&Field{
				Name: "name",
				Type: NewNonNull(String),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if v, ok := parent.(*EnumValue); ok {
						return v.Name, nil
					}
					return nil, nil
				},
			},
			&Field{
				Name: "description",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if v, ok := parent.(*EnumValue); ok {
						return v.Description, nil
					}
					return nil, nil
				},
			},
			&Field{
				Name: "isDeprecated",
				Type: NewNonNull(Boolean),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					/* if v, ok := parent.(EnumValue); ok {
						return v, nil
					} */
					return false, nil
				},
			},
			&Field{
				Name: "deprecationReason",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					/* if v, ok := parent.(EnumValue); ok {
						return v, nil
					} */
					return nil, nil
				},
			},
		},
	}

	typeKindIntrospection = &Enum{
		Name: "__TypeKind",
		Values: EnumValues{
			&EnumValue{
				Name:  "SCALAR",
				Value: "SCALAR",
			},
			&EnumValue{
				Name:  "OBJECT",
				Value: "OBJECT",
			},
			&EnumValue{
				Name:  "INTERFACE",
				Value: "INTERFACE",
			},
			&EnumValue{
				Name:  "UNION",
				Value: "UNION",
			},
			&EnumValue{
				Name:  "ENUM",
				Value: "ENUM",
			},
			&EnumValue{
				Name:  "INPUT_OBJECT",
				Value: "INPUT_OBJECT",
			},
			&EnumValue{
				Name:  "LIST",
				Value: "LIST",
			},
			&EnumValue{
				Name:  "NON_NULL",
				Value: "NON_NULL",
			},
		},
	}

	directiveIntrospection = &Object{
		Name: "__Directive",
		Fields: Fields{
			&Field{
				Name: "name",
				Type: NewNonNull(String),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if v, ok := parent.(*Directive); ok {
						return v.Name, nil
					}
					return nil, nil
				},
			},
			&Field{
				Name: "description",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if v, ok := parent.(*Directive); ok {
						return v.Description, nil
					}
					return nil, nil
				},
			},
			&Field{
				Name: "locations",
				Type: NewNonNull(NewList(NewNonNull(directiveLocationIntrospection))),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if v, ok := parent.(*Directive); ok {
						return v.Locations, nil
					}
					return nil, nil
				},
			},
			&Field{
				Name: "args",
				Type: NewNonNull(NewList(NewNonNull(inputValueIntrospection))),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if v, ok := parent.(*Directive); ok {
						return v.Arguments, nil
					}
					return []interface{}{}, nil
				},
			},
		},
	}

	directiveLocationIntrospection = &Enum{
		Name: "__DirectiveLocation",
		Values: EnumValues{
			&EnumValue{
				Name:  "QUERY",
				Value: "QUERY",
			},
			&EnumValue{
				Name:  "MUTATION",
				Value: "MUTATION",
			},
			&EnumValue{
				Name:  "SUBSCRIPTION",
				Value: "SUBSCRIPTION",
			},
			&EnumValue{
				Name:  "FIELD",
				Value: "FIELD",
			},
			&EnumValue{
				Name:  "FRAGMENT_DEFINITION",
				Value: "FRAGMENT_DEFINITION",
			},
			&EnumValue{
				Name:  "FRAGMENT_SPREAD",
				Value: "FRAGMENT_SPREAD",
			},
			&EnumValue{
				Name:  "INLINE_FRAGMENT",
				Value: "INLINE_FRAGMENT",
			},
			&EnumValue{
				Name:  "SCHEMA",
				Value: "SCHEMA",
			},
			&EnumValue{
				Name:  "SCALAR",
				Value: "SCALAR",
			},
			&EnumValue{
				Name:  "OBJECT",
				Value: "OBJECT",
			},
			&EnumValue{
				Name:  "FIELD_DEFINITION",
				Value: "FIELD_DEFINITION",
			},
			&EnumValue{
				Name:  "ARGUMENT_DEFINITION",
				Value: "ARGUMENT_DEFINITION",
			},
			&EnumValue{
				Name:  "INTERFACE",
				Value: "INTERFACE",
			},
			&EnumValue{
				Name:  "UNION",
				Value: "UNION",
			},
			&EnumValue{
				Name:  "ENUM",
				Value: "ENUM",
			},
			&EnumValue{
				Name:  "ENUM_VALUE",
				Value: "ENUM_VALUE",
			},
			&EnumValue{
				Name:  "INPUT_OBJECT",
				Value: "INPUT_OBJECT",
			},
			&EnumValue{
				Name:  "INPUT_FIELD_DEFINITION",
				Value: "INPUT_FIELD_DEFINITION",
			},
		},
	}
)
