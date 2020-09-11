package gql

import (
	"context"

	"github.com/rigglo/gql/pkg/language/ast"
)

type Directives []Directive

type Directive interface {
	GetName() string
	GetDescription() string
	GetArguments() Arguments
	GetLocations() []DirectiveLocation
}

/*
	type ExecutableDirective interface {
		// any Executable Directive specific function comes here
	}
*/

type TypeSystemDirective interface {
	Directive
	GetValues() map[string]interface{}
}

func (ds TypeSystemDirectives) ast() []*ast.Directive {
	out := []*ast.Directive{}
	for _, d := range ds {
		od := ast.Directive{
			Name:      d.GetName(),
			Arguments: make([]*ast.Argument, 0),
		}
		for an, a := range d.GetArguments() {
			od.Arguments = append(od.Arguments, &ast.Argument{
				Name:  an,
				Value: toAstValue(a.Type, d.GetValues()[an]),
			})
		}
		out = append(out, &od)
	}
	return out
}

type TypeSystemDirectives []TypeSystemDirective

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

type SchemaDirective interface {
	VisitSchema(context.Context, Schema) *Schema
}

type ScalarDirective interface {
	VisitScalar(context.Context, Scalar) *Scalar
}

type ObjectDirective interface {
	VisitObject(context.Context, Object) *Object
}

type FieldDefinitionDirective interface {
	VisitFieldDefinition(context.Context, Field, Resolver) Resolver
}

type ArgumentDirective interface {
	VisitArgument(context.Context, Argument)
}

type InterfaceDirective interface {
	VisitInterface(context.Context, Interface) *Interface
}

type UnionDirective interface {
	VisitUnion(context.Context, Union) *Union
}

type EnumDirective interface {
	VisitEnum(context.Context, Enum) *Enum
}

type EnumValueDirective interface {
	VisitEnumValue(context.Context, EnumValue) *EnumValue
}

type InputObjectDirective interface {
	VisitInputObject(context.Context, InputObject) *InputObject
}

type InputFieldDirective interface {
	VisitInputField(context.Context, InputField) *InputField
}

type skip struct{}

func (s *skip) GetName() string {
	return "skip"
}

func (s *skip) GetDescription() string {
	return "The @skip directive may be provided for fields, fragment spreads, and inline fragments, and allows for conditional exclusion during execution as described by the if argument"
}

func (s *skip) GetArguments() Arguments {
	return Arguments{
		"if": &Argument{
			Type: NewNonNull(Boolean),
		},
	}
}

func (s *skip) GetLocations() []DirectiveLocation {
	return []DirectiveLocation{
		FieldLoc,
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

func (s *include) GetArguments() Arguments {
	return Arguments{
		"if": &Argument{
			Type: NewNonNull(Boolean),
		},
	}
}

func (s *include) GetLocations() []DirectiveLocation {
	return []DirectiveLocation{
		FieldLoc,
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

func Deprecate(reason string) TypeSystemDirective {
	return &deprecated{reason}
}

type deprecated struct {
	reason string
}

func (d *deprecated) GetName() string {
	return "deprecated"
}

func (d *deprecated) GetDescription() string {
	return "The `@deprecated` directive is used within the type system definition language to indicate deprecated portions of a GraphQL service’s schema, such as deprecated fields on a type or deprecated enum values"
}

func (d *deprecated) GetArguments() Arguments {
	return Arguments{
		"reason": &Argument{
			Type:         String,
			DefaultValue: "",
		},
	}
}

func (d *deprecated) GetLocations() []DirectiveLocation {
	return []DirectiveLocation{
		FieldDefinitionLoc,
		EnumValueLoc,
	}
}

func (d *deprecated) Reason() string {
	return d.reason
}

func (d *deprecated) GetValues() map[string]interface{} {
	return map[string]interface{}{
		"reason": d.reason,
	}
}

var (
	skipDirective       = &skip{}
	includeDirective    = &include{}
	deprecatedDirective = &deprecated{}
)
