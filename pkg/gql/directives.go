package gql

type Directives []*Directive

type Directive struct {
	Description string
	Name        string
	Arguments   Arguments
	Locations   []DirectiveLocation
}

func (d *Directive) GetDescription() string {
	return d.Description
}

func (d *Directive) GetName() string {
	return d.Name
}

func (d *Directive) GetArguments() []*Argument {
	return d.Arguments
}

func (d *Directive) GetLocations() []DirectiveLocation {
	return d.Locations
}

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

var DeprecatedDirective *Directive = &Directive{
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
}

type Skip struct {
	*Directive
	Func string
}

func asd() {

}

// var _ Directive = DeprecatedDirective
