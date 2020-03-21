package schema

import (
	"strings"
)

type Directive interface {
	GetDescription() string
	GetName() string
	GetArguments() []Argument
	GetLocations() []DirectiveLocation
}

type DirectiveLocation string

func IsExecutableDirectiveLocation(loc DirectiveLocation) bool {
	return strings.Contains("QUERY MUTATION SUBSCRIPTION FIELD FRAGMENT_DEFINITION FRAGMENT_SPREAD INLINE_FRAGMENT", string(loc))
}

func IsTypeSystemDirectiveLocation(loc DirectiveLocation) bool {
	return strings.Contains("SCHEMA SCALAR OBJECT FIELD_DEFINITION ARGUMENT_DEFINITION INTERFACE UNION ENUM ENUM_VALUE INPUT_OBJECT INPUT_FIELD_DEFINITION", string(loc))
}

const (
	QueryLoc                DirectiveLocation = "QUERY"
	MutationLoc             DirectiveLocation = "MUTATION"
	SubscriptionLoc         DirectiveLocation = "SUBSCRIPTION"
	FieldLoc                DirectiveLocation = "FIELD"
	FragmentDefinitionLoc   DirectiveLocation = "FRAGMENT_DEFINITION"
	FragmentSpreadLoc       DirectiveLocation = "FRAGMENT_SPREAD"
	InlineFragmentLoc       DirectiveLocation = "INLINE_FRAGMENT"
	SchemaLoc               DirectiveLocation = "SCHEMA"
	ScalarLoc               DirectiveLocation = "SCALAR"
	ObjectLoc               DirectiveLocation = "OBJECT"
	FieldDefinitionLoc      DirectiveLocation = "FIELD_DEFINITION"
	ArgumentDefinitionLoc   DirectiveLocation = "ARGUMENT_DEFINITION"
	InterfaceLoc            DirectiveLocation = "INTERFACE"
	UnionLoc                DirectiveLocation = "UNION"
	EnumLoc                 DirectiveLocation = "ENUM"
	EnumValueLoc            DirectiveLocation = "ENUM_VALUE"
	InputObjectLoc          DirectiveLocation = "INPUT_OBJECT"
	InputFieldDefinitionLoc DirectiveLocation = "INPUT_FIELD_DEFINITION"
)

/*
type directive struct {
	name string
}

func (d *directive) Name() string {
	return d.name
}

func (d *directive) Arguments() []Argument {
	return nil
}

func (d *directive) GetLocations() []DirectiveLocation {
	return nil
}

var DirectiveSkip Directive = &directive{
	name: "skip",
}
var DirectiveInclude Directive = &directive{
	name: "include",
}
*/
