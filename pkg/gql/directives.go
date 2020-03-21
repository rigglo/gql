package gql

import "github.com/rigglo/gql/pkg/schema"

type Directives []schema.Directive

type Directive struct {
	Description string
	Name        string
	Arguments   Arguments
	Locations   []schema.DirectiveLocation
}

func (d *Directive) GetDescription() string {
	return d.Description
}

func (d *Directive) GetName() string {
	return d.Name
}

func (d *Directive) GetArguments() []schema.Argument {
	return d.Arguments
}

func (d *Directive) GetLocations() []schema.DirectiveLocation {
	return d.Locations
}

const (
	// Executable directive locations
	QueryLoc              schema.DirectiveLocation = "QUERY"
	MutationLoc           schema.DirectiveLocation = "MUTATION"
	SubscriptionLoc       schema.DirectiveLocation = "SUBSCRIPTION"
	FieldLoc              schema.DirectiveLocation = "FIELD"
	FragmentDefinitionLoc schema.DirectiveLocation = "FRAGMENT_DEFINITION"
	FragmentSpreadLoc     schema.DirectiveLocation = "FRAGMENT_SPREAD"
	InlineFragmentLoc     schema.DirectiveLocation = "INLINE_FRAGMENT"

	// Type System directive locations
	SchemaLoc               schema.DirectiveLocation = "SCHEMA"
	ScalarLoc               schema.DirectiveLocation = "SCALAR"
	ObjectLoc               schema.DirectiveLocation = "OBJECT"
	FieldDefinitionLoc      schema.DirectiveLocation = "FIELD_DEFINITION"
	ArgumentDefinitionLoc   schema.DirectiveLocation = "ARGUMENT_DEFINITION"
	InterfaceLoc            schema.DirectiveLocation = "INTERFACE"
	UnionLoc                schema.DirectiveLocation = "UNION"
	EnumLoc                 schema.DirectiveLocation = "ENUM"
	EnumValueLoc            schema.DirectiveLocation = "ENUM_VALUE"
	InputObjectLoc          schema.DirectiveLocation = "INPUT_OBJECT"
	InputFieldDefinitionLoc schema.DirectiveLocation = "INPUT_FIELD_DEFINITION"
)

var DepricatiedDirective *Directive = &Directive{
	Name: "depricated",
	Arguments: Arguments{
		&Argument{
			Name:        "reason",
			Type:        String,
			Description: "Reason of the deprication",
		},
	},
	Locations: []schema.DirectiveLocation{
		FieldDefinitionLoc,
		EnumValueLoc,
	},
}

var _ schema.Directive = DepricatiedDirective
