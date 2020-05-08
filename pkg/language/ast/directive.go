package ast

import (
	"fmt"
	"strings"
)

const (
	// Executable directive locations
	QueryDirectiveLocation              string = "QUERY"
	MutationDirectiveLocation           string = "MUTATION"
	SubscriptionDirectiveLocation       string = "SUBSCRIPTION"
	FieldDirectiveLocation              string = "FIELD"
	FragmentDefinitionDirectiveLocation string = "FRAGMENT_DEFINITION"
	FragmentSpreadDirectiveLocation     string = "FRAGMENT_SPREAD"
	InlineFragmentDirectiveLocation     string = "INLINE_FRAGMENT"

	// TypeSystem directive locations
	SchemaDirectiveLocation               string = "SCHEMA"
	ScalarDirectiveLocation               string = "SCALAR"
	ObjectDirectiveLocation               string = "OBJECT"
	FieldDefinitionDirectiveLocation      string = "FIELD_DEFINITION"
	ArguentDefinitionDirectiveLocation    string = "ARGUMENT_DEFINITION"
	InterfaceDirectiveLocation            string = "INTERFACE"
	UnionDirectiveLocation                string = "UNION"
	EnumDirectiveLocation                 string = "ENUM"
	EnumValueDirectiveLocation            string = "ENUM_VALUE"
	InputObjectDirectiveLocation          string = "INPUT_OBJECT"
	InputFieldDefinitionDirectiveLocation string = "INPUT_FIELD_DEFINITION"
)

func IsValidDirective(d string) bool {
	return strings.Contains("QUERY MUTATION SUBSCRIPTION FIELD FRAGMENT_DEFINITION FRAGMENT_SPREAD INLINE_FRAGMENT SCHEMA SCALAR OBJECT FIELD_DEFINITION ARGUMENT_DEFINITION INTERFACE UNION ENUM ENUM_VALUE INPUT_OBJECT INPUT_FIELD_DEFINITION", fmt.Sprintf(" %s ", d))
}
