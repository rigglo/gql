package gql

import (
	"context"

	"github.com/rigglo/gql/pkg/language/ast"
)

type Directives []Directive

type Directive interface {
	GetName() string
	GetDescription() string
	GetArguments() []*Argument
	GetLocations() []DirectiveLocation
}

type skip struct{}

func (s *skip) GetName() string {
	return "skip"
}

func (s *skip) GetDescription() string {
	return "The @skip directive may be provided for fields, fragment spreads, and inline fragments, and allows for conditional exclusion during execution as described by the if argument"
}

func (s *skip) GetArguments() []*Argument {
	return []*Argument{
		&Argument{
			Name: "if",
			Type: NewNonNull(Boolean),
		},
	}
}

func (s *skip) GetLocations() []DirectiveLocation {
	return []DirectiveLocation{
		FieldDefinitionLoc,
		FragmentSpreadLoc,
		InlineFragmentLoc,
	}
}

func (s *skip) Skip(args []*ast.Argument) bool {
	if args[0].Value.GetValue().(string) == "true" {
		return true
	}
	return false
}

type include struct{}

func (s *include) GetName() string {
	return "include"
}

func (s *include) GetDescription() string {
	return "The @skip directive may be provided for fields, fragment spreads, and inline fragments, and allows for conditional exclusion during execution as described by the if argument"
}

func (s *include) GetArguments() []*Argument {
	return []*Argument{
		&Argument{
			Name: "if",
			Type: NewNonNull(Boolean),
		},
	}
}

func (s *include) GetLocations() []DirectiveLocation {
	return []DirectiveLocation{
		FieldDefinitionLoc,
		FragmentSpreadLoc,
		InlineFragmentLoc,
	}
}

func (s *include) Include(args []*ast.Argument) bool {
	if args[0].Value.GetValue().(string) == "true" {
		return true
	}
	return false
}

var (
	skipDirective    = &skip{}
	includeDirective = &include{}
)

type DirectiveLocation string

const (
	// Executable directive locations
	QueryLoc              DirectiveLocation = "QUERY"
	MutationLoc           DirectiveLocation = "MUTATION"
	SubscriptionLoc       DirectiveLocation = "SUBSCRIPTION"
	FieldLoc              DirectiveLocation = "FIELD"
	FragmentDefinitionLoc DirectiveLocation = "FRAGMENT_DEFINITION"
	FragmentSpreadLoc     DirectiveLocation = "FRAGMENT_SPREAD"
	InlineFragmentLoc     DirectiveLocation = "INLINE_FRAGMENT"

	// Type System directive locations
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

/* var DeprecatedDirective *Directive = &Directive{
	Name: "deprecated",
	Arguments: Arguments{
		&Argument{
			Name:        "reason",
			Type:        String,
			Description: "Reason of the deprecation",
		},
	},
	Locations: []DirectiveLocation{
		FieldDefinitionLoc,
		EnumValueLoc,
	},
} */

// var _ Directive = DeprecatedDirective

type SchemaDirective interface {
	visitSchema(context.Context, Schema) *Schema
}

type ScalarDirective interface {
	visitScalar(context.Context, Scalar) *Scalar
}

type ObjectDirective interface {
	visitObject(context.Context, Object) *Object
}

type FieldDirective interface {
	visitField(context.Context, Field) *Field
}

type ArgumentDirective interface {
	visitArgument(context.Context, Argument) *Argument
}

type InterfaceDirective interface {
	visitInterface(context.Context, Interface) *Interface
}

type UnionDirective interface {
	visitUnion(context.Context, Union) *Union
}

type EnumDirective interface {
	visitEnum(context.Context, Enum) *Enum
}

type EnumValueDirective interface {
	visitEnumValue(context.Context, EnumValue) *EnumValue
}

type InputObjectDirective interface {
	visitInputObject(context.Context, InputObject) *InputObject
}

type InputFieldDirective interface {
	visitInputField(context.Context, InputField) *InputField
}
