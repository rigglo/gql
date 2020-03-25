package gql

func init() {
	typeIntrospection.AddFields(
		&Field{
			Name: "interfaces",
			Type: NewList(NewNonNull(typeIntrospection)),
		},
		&Field{
			Name: "possibleTypes",
			Type: NewList(NewNonNull(typeIntrospection)),
		},
		&Field{
			Name: "ofType",
			Type: typeIntrospection,
		},
	)

	inputValueIntrospection.AddFields(
		&Field{
			Name: "type",
			Type: NewNonNull(typeIntrospection),
		},
	)

	fieldIntrospection.AddFields(
		&Field{
			Name: "type",
			Type: NewNonNull(typeIntrospection),
		},
	)
}

var (
	schemaIntrospection = &Object{
		Name: "__Schema",
		Fields: Fields{
			&Field{
				Name: "types",
				Type: NewNonNull(NewList(NewNonNull(typeIntrospection))),
			},
			&Field{
				Name: "queryType",
				Type: NewNonNull(typeIntrospection),
			},
			&Field{
				Name: "mutationType",
				Type: typeIntrospection,
			},
			&Field{
				Name: "subscriptionType",
				Type: typeIntrospection,
			},
			&Field{
				Name: "directives",
				Type: typeIntrospection,
			},
		},
	}

	typeIntrospection = &Object{
		Name: "__Type",
		Fields: Fields{
			&Field{
				Name: "kind",
				Type: NewNonNull(typeKindIntrospection),
			},
			&Field{
				Name: "name",
				Type: String,
			},
			&Field{
				Name: "description",
				Type: String,
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
				Type: NewList(NewNonNull(fieldIntrospection)),
			},
			&Field{
				Name: "inputFields",
				Type: NewList(NewNonNull(inputValueIntrospection)),
			},
		},
	}

	fieldIntrospection = &Object{
		Name: "__Field",
		Fields: Fields{
			&Field{
				Name: "name",
				Type: NewNonNull(String),
			},
			&Field{
				Name: "description",
				Type: String,
			},
			&Field{
				Name: "args",
				Type: NewNonNull(NewList(NewNonNull(inputValueIntrospection))),
			},
			&Field{
				Name: "isDepricated",
				Type: NewNonNull(Boolean),
			},
			&Field{
				Name: "deprecationReason",
				Type: String,
			},
		},
	}

	inputValueIntrospection = &Object{
		Name: "__InputValue",
		Fields: Fields{
			&Field{
				Name: "name",
				Type: NewNonNull(String),
			},
			&Field{
				Name: "description",
				Type: String,
			},
			&Field{
				Name: "defaultValue",
				Type: String,
			},
		},
	}

	enumValueIntrospection = &Object{
		Name: "__EnumValue",
		Fields: Fields{
			&Field{
				Name: "name",
				Type: NewNonNull(String),
			},
			&Field{
				Name: "description",
				Type: String,
			},
			&Field{
				Name: "isDepricated",
				Type: NewNonNull(Boolean),
			},
			&Field{
				Name: "deprecationReason",
				Type: String,
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
			},
			&Field{
				Name: "description",
				Type: String,
			},
			&Field{
				Name: "locations",
				Type: NewNonNull(NewList(NewNonNull(directiveLocationIntrospection))),
			},
			&Field{
				Name: "args",
				Type: NewNonNull(NewList(NewNonNull(inputValueIntrospection))),
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
