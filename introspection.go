package gql

import (
	"encoding/json"
)

func init() {
	typeIntrospection.AddField(
		"interfaces",
		&Field{
			Type: NewList(NewNonNull(typeIntrospection)),
			Resolver: func(ctx Context) (interface{}, error) {
				if ctx.Parent().(Type).GetKind() == ObjectKind {
					if ctx.Parent().(*Object).Implements == nil {
						return []struct{}{}, nil
					}
					return ctx.Parent().(*Object).Implements, nil
				}
				return nil, nil
			},
		})
	typeIntrospection.AddField(
		"possibleTypes",
		&Field{
			Type: NewList(NewNonNull(typeIntrospection)),
			Resolver: func(ctx Context) (interface{}, error) {
				if ctx.Parent().(Type).GetKind() == UnionKind {
					return ctx.Parent().(*Union).GetMembers(), nil
				} else if ctx.Parent().(Type).GetKind() == InterfaceKind {
					gqlctx := ctx.(*resolveContext).gqlCtx

					gqlctx.mu.Lock()
					out := gqlctx.implementors[ctx.Parent().(*Interface).GetName()]
					gqlctx.mu.Unlock()
					return out, nil
				}
				return nil, nil
			},
		})
	typeIntrospection.AddField(
		"ofType",
		&Field{
			Type: typeIntrospection,
			Resolver: func(ctx Context) (interface{}, error) {
				t := ctx.Parent().(Type)
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

	inputValueIntrospection.AddField(
		"type",
		&Field{
			Type: NewNonNull(typeIntrospection),
			Resolver: func(ctx Context) (interface{}, error) {
				if v, ok := ctx.Parent().(*iArgument); ok {
					return v.arg.Type, nil
				} else if v, ok := ctx.Parent().(struct {
					field *InputField
					name  string
				}); ok {
					return v.field.Type, nil
				}
				return nil, nil
			},
		},
	)

	fieldIntrospection.AddField(
		"type",
		&Field{
			Type: NewNonNull(typeIntrospection),
			Resolver: func(ctx Context) (interface{}, error) {
				return ctx.Parent().(*iField).field.GetType(), nil
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

type iField struct {
	field *Field
	name  string
}

type iArgument struct {
	arg  *Argument
	name string
}

type iInputField struct {
	field *InputField
	name  string
}

var (
	introspectionQuery = &Object{
		Fields: Fields{
			"__schema": &Field{
				Type: NewNonNull(schemaIntrospection),
				Resolver: func(ctx Context) (interface{}, error) {
					return ctx.(*resolveContext).gqlCtx.schema, nil
				},
			},
			"__type": &Field{
				Type: typeIntrospection,
				Arguments: Arguments{
					"name": &Argument{
						Type: NewNonNull(String),
					},
				},
				Resolver: func(ctx Context) (interface{}, error) {
					gqlctx := ctx.(*resolveContext).gqlCtx
					gqlctx.mu.Lock()
					t := gqlctx.types[ctx.Args()["name"].(string)]
					gqlctx.mu.Unlock()
					return t, nil
				},
			},
		},
	}

	schemaIntrospection = &Object{
		Name: "__Schema",
		Fields: Fields{
			"types": &Field{
				Type: NewNonNull(NewList(NewNonNull(typeIntrospection))),
				Resolver: func(ctx Context) (interface{}, error) {
					gqlctx := ctx.(*resolveContext).gqlCtx

					gqlctx.mu.Lock()
					out := []Type{}
					for _, t := range gqlctx.types {
						out = append(out, t)
					}
					gqlctx.mu.Unlock()
					return out, nil
				},
			},
			"queryType": &Field{
				Type: NewNonNull(typeIntrospection),
				Resolver: func(ctx Context) (interface{}, error) {
					return ctx.(*resolveContext).gqlCtx.schema.Query, nil
				},
			},
			"mutationType": &Field{
				Type: typeIntrospection,
				Resolver: func(ctx Context) (interface{}, error) {
					if ctx.(*resolveContext).gqlCtx.schema.Mutation == nil {
						return nil, nil
					}
					return ctx.(*resolveContext).gqlCtx.schema.Mutation, nil
				},
			},
			"subscriptionType": &Field{
				Type: typeIntrospection,
				Resolver: func(ctx Context) (interface{}, error) {
					if ctx.(*resolveContext).gqlCtx.schema.Subscription == nil {
						return nil, nil
					}
					return ctx.(*resolveContext).gqlCtx.schema.Subscription, nil
				},
			},
			"directives": &Field{
				Type: NewNonNull(NewList(NewNonNull(directiveIntrospection))),
				Resolver: func(ctx Context) (interface{}, error) {
					gqlctx := ctx.(*resolveContext).gqlCtx

					out := []Directive{}
					gqlctx.mu.Lock()
					for _, d := range gqlctx.directives {
						out = append(out, d)
					}
					gqlctx.mu.Unlock()
					return out, nil
				},
			},
		},
	}

	typeIntrospection = &Object{
		Name: "__Type",
		Fields: Fields{
			"kind": &Field{
				Type: NewNonNull(typeKindIntrospection),
				Resolver: func(ctx Context) (interface{}, error) {
					return ctx.Parent().(Type).GetKind(), nil
				},
			},
			"name": &Field{
				Type: String,
				Resolver: func(ctx Context) (interface{}, error) {
					if _, ok := ctx.Parent().(WrappingType); ok {
						return nil, nil
					}
					return ctx.Parent().(Type).GetName(), nil
				},
			},
			"description": &Field{
				Type: String,
				Resolver: func(ctx Context) (interface{}, error) {
					if _, ok := ctx.Parent().(WrappingType); ok {
						return nil, nil
					}
					return ctx.Parent().(Type).GetDescription(), nil
				},
			},
			"fields": &Field{
				Arguments: Arguments{
					"includeDeprecated": &Argument{
						Type:         Boolean,
						DefaultValue: false,
					},
				},
				Type: NewList(NewNonNull(fieldIntrospection)),
				Resolver: func(ctx Context) (interface{}, error) {
					includeDeprecated := ctx.Args()["includeDeprecated"].(bool)

					if ctx.Parent().(Type).GetKind() == ObjectKind {
						if includeDeprecated {
							out := []*iField{}
							for fn, f := range ctx.Parent().(*Object).Fields {
								out = append(out, &iField{f, fn})
							}
							return out, nil
						} else {
							out := []*iField{}
							for fn, f := range ctx.Parent().(*Object).Fields {
								if !f.IsDeprecated() {
									out = append(out, &iField{f, fn})
								}
							}
							return out, nil
						}
					}
					if ctx.Parent().(Type).GetKind() == InterfaceKind {
						if includeDeprecated {
							out := []*iField{}
							for fn, f := range ctx.Parent().(*Interface).Fields {
								out = append(out, &iField{f, fn})
							}
							return out, nil
						} else {
							out := []*iField{}
							for fn, f := range ctx.Parent().(*Interface).Fields {
								if !f.IsDeprecated() {
									out = append(out, &iField{f, fn})
								}
							}
							return out, nil
						}
					}
					return nil, nil
				},
			},
			"enumValues": &Field{
				Arguments: Arguments{
					"includeDeprecated": &Argument{
						Type:         Boolean,
						DefaultValue: false,
					},
				},
				Type: NewList(NewNonNull(enumValueIntrospection)),
				Resolver: func(ctx Context) (interface{}, error) {
					includeDeprecated := ctx.Args()["includeDeprecated"].(bool)

					if ctx.Parent().(Type).GetKind() == EnumKind {
						if includeDeprecated {
							return ctx.Parent().(*Enum).GetValues(), nil
						} else {
							out := []*EnumValue{}
							for _, v := range ctx.Parent().(*Enum).GetValues() {
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
			"inputFields": &Field{
				Type: NewList(NewNonNull(inputValueIntrospection)),
				Resolver: func(ctx Context) (interface{}, error) {
					if ctx.Parent().(Type).GetKind() == InputObjectKind {
						fs := []struct {
							field *InputField
							name  string
						}{}
						for fn, f := range ctx.Parent().(*InputObject).GetFields() {
							fs = append(fs, struct {
								field *InputField
								name  string
							}{f, fn})
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
			"name": &Field{
				Type: NewNonNull(String),
				Resolver: func(ctx Context) (interface{}, error) {
					return ctx.Parent().(*iField).name, nil
				},
			},
			"description": &Field{
				Type: String,
				Resolver: func(ctx Context) (interface{}, error) {
					return ctx.Parent().(*iField).field.GetDescription(), nil
				},
			},
			"args": &Field{
				Type: NewNonNull(NewList(NewNonNull(inputValueIntrospection))),
				Resolver: func(ctx Context) (interface{}, error) {
					out := []*iArgument{}
					for aName, a := range ctx.Parent().(*iField).field.GetArguments() {
						out = append(out, &iArgument{
							arg:  a,
							name: aName,
						})
					}
					return out, nil
				},
			},
			"isDeprecated": &Field{
				Type: NewNonNull(Boolean),
				Resolver: func(ctx Context) (interface{}, error) {
					for _, d := range ctx.Parent().(*iField).field.Directives {
						if _, ok := d.(*deprecated); ok {
							return true, nil
						}
					}
					return false, nil
				},
			},
			"deprecationReason": &Field{
				Type: String,
				Resolver: func(ctx Context) (interface{}, error) {
					for _, d := range ctx.Parent().(*iField).field.Directives {
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
			"name": &Field{
				Type: NewNonNull(String),
				Resolver: func(ctx Context) (interface{}, error) {
					if v, ok := ctx.Parent().(*EnumValue); ok {
						return v.Name, nil
					} else if v, ok := ctx.Parent().(*iArgument); ok {
						return v.name, nil
					} else if v, ok := ctx.Parent().(*Scalar); ok {
						return v.Name, nil
					} else if v, ok := ctx.Parent().(struct {
						field *InputField
						name  string
					}); ok {
						return v.name, nil
					}
					return nil, nil
				},
			},
			"description": &Field{
				Type: String,
				Resolver: func(ctx Context) (interface{}, error) {
					if f, ok := ctx.Parent().(struct {
						field *InputField
						name  string
					}); ok {
						return f.field.Description, nil
					} else if a, ok := ctx.Parent().(*iArgument); ok {
						return a.arg.Description, nil
					}
					return nil, nil
				},
			},
			"defaultValue": &Field{
				Type: String,
				Resolver: func(ctx Context) (interface{}, error) {
					if v, ok := ctx.Parent().(*iArgument); ok && v.arg.IsDefaultValueSet() {
						if unwrapper(v.arg.Type).GetKind() == EnumKind {
							return v.arg.DefaultValue, nil
						}
						if v.arg.DefaultValue != nil {
							bs, err := json.Marshal(v.arg.DefaultValue)
							if err != nil {
								return nil, err
							}
							return string(bs), nil
						}
					} else if v, ok := ctx.Parent().(iInputField); ok && v.field.IsDefaultValueSet() {
						if unwrapper(v.field.Type).GetKind() == EnumKind {
							return v.field.DefaultValue, nil
						}
						if v.field.DefaultValue != nil {
							bs, err := json.Marshal(v.field.DefaultValue)
							if err != nil {
								return nil, err
							}
							return string(bs), nil
						}
					}
					return nil, nil
				},
			},
		},
	}

	enumValueIntrospection = &Object{
		Name: "__EnumValue",
		Fields: Fields{
			"name": &Field{
				Type: NewNonNull(String),
				Resolver: func(ctx Context) (interface{}, error) {
					return ctx.Parent().(*EnumValue).Name, nil
				},
			},
			"description": &Field{
				Type: String,
				Resolver: func(ctx Context) (interface{}, error) {
					return ctx.Parent().(*EnumValue).Description, nil
				},
			},
			"isDeprecated": &Field{
				Type: NewNonNull(Boolean),
				Resolver: func(ctx Context) (interface{}, error) {
					for _, d := range ctx.Parent().(*EnumValue).GetDirectives() {
						if _, ok := d.(*deprecated); ok {
							return true, nil
						}
					}
					return false, nil
				},
			},
			"deprecationReason": &Field{
				Type: String,
				Resolver: func(ctx Context) (interface{}, error) {
					for _, d := range ctx.Parent().(*EnumValue).GetDirectives() {
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
			"name": &Field{
				Type: NewNonNull(String),
				Resolver: func(ctx Context) (interface{}, error) {
					return ctx.Parent().(Directive).GetName(), nil
				},
			},
			"description": &Field{
				Type: String,
				Resolver: func(ctx Context) (interface{}, error) {
					return ctx.Parent().(Directive).GetDescription(), nil
				},
			},
			"locations": &Field{
				Type: NewNonNull(NewList(NewNonNull(directiveLocationIntrospection))),
				Resolver: func(ctx Context) (interface{}, error) {
					return ctx.Parent().(Directive).GetLocations(), nil
				},
			},
			"args": &Field{
				Type: NewNonNull(NewList(NewNonNull(inputValueIntrospection))),
				Resolver: func(ctx Context) (interface{}, error) {
					out := []*iArgument{}
					for aName, a := range ctx.Parent().(Directive).GetArguments() {
						out = append(out, &iArgument{
							arg:  a,
							name: aName,
						})
					}
					return out, nil
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
