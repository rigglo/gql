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
				if parent.(Type).GetKind() == ObjectKind {
					return parent.(*Object).GetInterfaces(), nil
				}
				return nil, nil
			},
		},
		&Field{
			Name: "possibleTypes",
			Type: NewList(NewNonNull(typeIntrospection)),
			Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
				if parent.(Type).GetKind() == UnionKind {
					return parent.(*Union).GetMembers(), nil
				} else if parent.(Type).GetKind() == InterfaceKind {
					ectx := ctx.(*eCtx)
					ts := ectx.Get(keyTypes).(map[string]Type)

					out := []Type{}
					for _, t := range ts {
						if t.GetKind() == ObjectKind {
							t.(*Object).DoesImplement(parent.(*Interface))
						}
						out = append(out, t)
					}
					return out, nil
				}
				return nil, nil
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
				} else if v, ok := parent.(*InputField); ok {
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
				return parent.(*Field).GetType(), nil
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
					ectx := ctx.(*eCtx)
					s := ectx.Get(keySchema).(*Schema)
					return s, nil
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
					ectx := ctx.(*eCtx)
					schema := ectx.Get(keySchema).(*Schema)
					if schema.GetRootMutation() == nil {
						return nil, nil
					}
					return schema.GetRootMutation(), nil
				},
			},
			&Field{
				Name: "subscriptionType",
				Type: typeIntrospection,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					ectx := ctx.(*eCtx)
					schema := ectx.Get(keySchema).(*Schema)
					if schema.GetRootSubsciption() == nil {
						return nil, nil
					}
					return schema.GetRootSubsciption(), nil
				},
			},
			&Field{
				Name: "directives",
				Type: NewNonNull(NewList(NewNonNull(directiveIntrospection))),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					ectx := ctx.(*eCtx)
					ds := ectx.Get(keyDirectives).(map[string]Directive)
					out := []Directive{}
					for _, d := range ds {
						out = append(out, d)
					}
					return out, nil
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
					return parent.(Type).GetKind(), nil
				},
			},
			&Field{
				Name: "name",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if _, ok := parent.(WrappingType); ok {
						return nil, nil
					}
					return parent.(Type).GetName(), nil
				},
			},
			&Field{
				Name: "description",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if _, ok := parent.(WrappingType); ok {
						return nil, nil
					}
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
					includeDeprecated := args["includeDeprecated"].(bool)

					if parent.(Type).GetKind() == ObjectKind {
						if includeDeprecated {
							return parent.(*Object).GetFields(), nil
						} else {
							out := []*Field{}
							for _, f := range parent.(*Object).GetFields() {
								if !f.IsDeprecated() {
									out = append(out, f)
								}
							}
							return out, nil
						}
					}
					if parent.(Type).GetKind() == InterfaceKind {
						if includeDeprecated {
							return parent.(*Interface).GetFields(), nil
						} else {
							out := []*Field{}
							for _, f := range parent.(*Interface).GetFields() {
								if !f.IsDeprecated() {
									out = append(out, f)
								}
							}
							return out, nil
						}
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
					includeDeprecated := args["includeDeprecated"].(bool)

					if parent.(Type).GetKind() == EnumKind {
						if includeDeprecated {
							return parent.(*Enum).GetValues(), nil
						} else {
							out := []*EnumValue{}
							for _, v := range parent.(*Enum).GetValues() {
								if !v.IsDeprecated() {
									out = append(out, v)
								}
							}
							return out, nil
						}
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
					return parent.(*Field).GetName(), nil
				},
			},
			&Field{
				Name: "description",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return parent.(*Field).GetDescription(), nil
				},
			},
			&Field{
				Name: "args",
				Type: NewNonNull(NewList(NewNonNull(inputValueIntrospection))),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return parent.(*Field).GetArguments(), nil
				},
			},
			&Field{
				Name: "isDeprecated",
				Type: NewNonNull(Boolean),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					for _, d := range parent.(*Field).GetDirectives() {
						if _, ok := d.(*deprecated); ok {
							return true, nil
						}
					}
					return false, nil
				},
			},
			&Field{
				Name: "deprecationReason",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					for _, d := range parent.(*Field).GetDirectives() {
						if dep, ok := d.(*deprecated); ok {
							return dep.reason, nil
						}
					}
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
					if v, ok := parent.(*Argument); ok && v.IsDefaultValueSet() {
						return v.DefaultValue, nil
					} else if v, ok := parent.(*InputField); ok && v.IsDefaultValueSet() {
						return v.DefaultValue, nil
					}
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
					return parent.(*EnumValue).Name, nil
				},
			},
			&Field{
				Name: "description",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return parent.(*EnumValue).Description, nil
				},
			},
			&Field{
				Name: "isDeprecated",
				Type: NewNonNull(Boolean),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					for _, d := range parent.(*EnumValue).GetDirectives() {
						if _, ok := d.(*deprecated); ok {
							return true, nil
						}
					}
					return false, nil
				},
			},
			&Field{
				Name: "deprecationReason",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					for _, d := range parent.(*EnumValue).GetDirectives() {
						if dep, ok := d.(*deprecated); ok {
							return dep.reason, nil
						}
					}
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
				Value: ScalarKind,
			},
			&EnumValue{
				Name:  "OBJECT",
				Value: ObjectKind,
			},
			&EnumValue{
				Name:  "INTERFACE",
				Value: InterfaceKind,
			},
			&EnumValue{
				Name:  "UNION",
				Value: UnionKind,
			},
			&EnumValue{
				Name:  "ENUM",
				Value: EnumKind,
			},
			&EnumValue{
				Name:  "INPUT_OBJECT",
				Value: InputObjectKind,
			},
			&EnumValue{
				Name:  "LIST",
				Value: ListKind,
			},
			&EnumValue{
				Name:  "NON_NULL",
				Value: NonNullKind,
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
					return parent.(Directive).GetName(), nil
				},
			},
			&Field{
				Name: "description",
				Type: String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return parent.(Directive).GetDescription(), nil
				},
			},
			&Field{
				Name: "locations",
				Type: NewNonNull(NewList(NewNonNull(directiveLocationIntrospection))),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return parent.(Directive).GetLocations(), nil
				},
			},
			&Field{
				Name: "args",
				Type: NewNonNull(NewList(NewNonNull(inputValueIntrospection))),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return parent.(Directive).GetArguments(), nil
				},
			},
		},
	}

	directiveLocationIntrospection = &Enum{
		Name: "__DirectiveLocation",
		Values: EnumValues{
			&EnumValue{
				Name:  "QUERY",
				Value: QueryLoc,
			},
			&EnumValue{
				Name:  "MUTATION",
				Value: MutationLoc,
			},
			&EnumValue{
				Name:  "SUBSCRIPTION",
				Value: SubscriptionLoc,
			},
			&EnumValue{
				Name:  "FIELD",
				Value: FieldLoc,
			},
			&EnumValue{
				Name:  "FRAGMENT_DEFINITION",
				Value: FragmentDefinitionLoc,
			},
			&EnumValue{
				Name:  "FRAGMENT_SPREAD",
				Value: FragmentSpreadLoc,
			},
			&EnumValue{
				Name:  "INLINE_FRAGMENT",
				Value: InlineFragmentLoc,
			},
			&EnumValue{
				Name:  "SCHEMA",
				Value: SchemaLoc,
			},
			&EnumValue{
				Name:  "SCALAR",
				Value: ScalarLoc,
			},
			&EnumValue{
				Name:  "OBJECT",
				Value: ObjectLoc,
			},
			&EnumValue{
				Name:  "FIELD_DEFINITION",
				Value: FieldDefinitionLoc,
			},
			&EnumValue{
				Name:  "ARGUMENT_DEFINITION",
				Value: ArgumentDefinitionLoc,
			},
			&EnumValue{
				Name:  "INTERFACE",
				Value: InterfaceLoc,
			},
			&EnumValue{
				Name:  "UNION",
				Value: UnionLoc,
			},
			&EnumValue{
				Name:  "ENUM",
				Value: EnumLoc,
			},
			&EnumValue{
				Name:  "ENUM_VALUE",
				Value: EnumValueLoc,
			},
			&EnumValue{
				Name:  "INPUT_OBJECT",
				Value: InputObjectLoc,
			},
			&EnumValue{
				Name:  "INPUT_FIELD_DEFINITION",
				Value: InputFieldDefinitionLoc,
			},
		},
	}
)
