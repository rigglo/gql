package gql

import "context"

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
	introspectionQueryies = map[string]*Field{
		"__schema": &Field{
			Name: "__schema",
			Type: NewNonNull(schemaIntrospection),
			Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
				return nil, nil
			},
		},
		"__type": &Field{},
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
						fs := []EnumValue{}
						for name, f := range parent.(*Enum).GetValues() {
							f.name = name
							fs = append(fs, f)
						}
						return fs, nil
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
					if v, ok := parent.(EnumValue); ok {
						return v, nil
					}
					return "nil", nil
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
					if v, ok := parent.(EnumValue); ok {
						return v.name, nil
					}
					return nil, nil
				},
			},
			&Field{
				Name: "description",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if v, ok := parent.(EnumValue); ok {
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
			"SCALAR": EnumValue{
				Value: "SCALAR",
			},
			"OBJECT": EnumValue{
				Value: "OBJECT",
			},
			"INTERFACE": EnumValue{
				Value: "INTERFACE",
			},
			"UNION": EnumValue{
				Value: "UNION",
			},
			"ENUM": EnumValue{
				Value: "ENUM",
			},
			"INPUT_OBJECT": EnumValue{
				Value: "INPUT_OBJECT",
			},
			"LIST": EnumValue{
				Value: "LIST",
			},
			"NON_NULL": EnumValue{
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
			"QUERY": EnumValue{
				Value: "QUERY",
			},
			"MUTATION": EnumValue{
				Value: "MUTATION",
			},
			"SUBSCRIPTION": EnumValue{
				Value: "SUBSCRIPTION",
			},
			"FIELD": EnumValue{
				Value: "FIELD",
			},
			"FRAGMENT_DEFINITION": EnumValue{
				Value: "FRAGMENT_DEFINITION",
			},
			"FRAGMENT_SPREAD": EnumValue{
				Value: "FRAGMENT_SPREAD",
			},
			"INLINE_FRAGMENT": EnumValue{
				Value: "INLINE_FRAGMENT",
			},
			"SCHEMA": EnumValue{
				Value: "SCHEMA",
			},
			"SCALAR": EnumValue{
				Value: "SCALAR",
			},
			"OBJECT": EnumValue{
				Value: "OBJECT",
			},
			"FIELD_DEFINITION": EnumValue{
				Value: "FIELD_DEFINITION",
			},
			"ARGUMENT_DEFINITION": EnumValue{
				Value: "ARGUMENT_DEFINITION",
			},
			"INTERFACE": EnumValue{
				Value: "INTERFACE",
			},
			"UNION": EnumValue{
				Value: "UNION",
			},
			"ENUM": EnumValue{
				Value: "ENUM",
			},
			"ENUM_VALUE": EnumValue{
				Value: "ENUM_VALUE",
			},
			"INPUT_OBJECT": EnumValue{
				Value: "INPUT_OBJECT",
			},
			"INPUT_FIELD_DEFINITION": EnumValue{
				Value: "INPUT_FIELD_DEFINITION",
			},
		},
	}
)
