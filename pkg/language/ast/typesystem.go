package ast

type DefinitionKind uint

const (
	SchemaKind DefinitionKind = iota
	ScalarKind
	ObjectKind
	InterfaceKind
	UnionKind
	EnumKind
	InputObjectKind
	DirectiveKind
)

type Definition interface {
	Kind() DefinitionKind
}

type SchemaDefinition struct {
	Name           string
	Directives     []*Directive
	RootOperations map[OperationType]*NamedType
}

func (d *SchemaDefinition) Kind() DefinitionKind {
	return SchemaKind
}

type ScalarDefinition struct {
	Description string
	Name        string
	Directives  []*Directive
}

func (d *ScalarDefinition) Kind() DefinitionKind {
	return ScalarKind
}

type ObjectDefinition struct {
	Description string
	Name        string
	Implements  []*NamedType
	Directives  []*Directive
	Fields      []*FieldDefinition
}

func (d *ObjectDefinition) Kind() DefinitionKind {
	return ObjectKind
}

type FieldDefinition struct {
	Description string
	Name        string
	Arguments   []*InputValueDefinition
	Type        Type
	Directives  []*Directive
}

type InputValueDefinition struct {
	Description  string
	Name         string
	Type         Type
	DefaultValue Value
	Directives   []*Directive
}

type InterfaceDefinition struct {
	Description string
	Name        string
	Directives  []*Directive
	Fields      []*FieldDefinition
}

func (d *InterfaceDefinition) Kind() DefinitionKind {
	return InterfaceKind
}

type UnionDefinition struct {
	Description string
	Name        string
	Directives  []*Directive
	Members     []*NamedType
}

func (d *UnionDefinition) Kind() DefinitionKind {
	return UnionKind
}

type EnumDefinition struct {
	Description string
	Name        string
	Directives  []*Directive
	Values      []*EnumValueDefinition
}

func (d *EnumDefinition) Kind() DefinitionKind {
	return EnumKind
}

type EnumValueDefinition struct {
	Description string
	Value       *EnumValue
	Directives  []*Directive
}

type InputObjectDefinition struct {
	Description string
	Name        string
	Directives  []*Directive
	Fields      []*InputValueDefinition
}

func (d *InputObjectDefinition) Kind() DefinitionKind {
	return InputObjectKind
}

type DirectiveDefinition struct {
	Description string
	Name        string
	Locations   []string
	Arguments   []*InputValueDefinition
}

func (d *DirectiveDefinition) Kind() DefinitionKind {
	return DirectiveKind
}
