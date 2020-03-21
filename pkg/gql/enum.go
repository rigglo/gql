package gql

import "github.com/rigglo/gql/pkg/schema"

type Enum struct {
	Name        string
	Description string
	Directives  Directives
	Values      EnumValues
}

func (e *Enum) GetDescription() string {
	return e.Description
}

func (e *Enum) GetName() string {
	return e.Name
}

func (e *Enum) GetDirectives() []schema.Directive {
	return e.Directives
}

func (e *Enum) GetKind() schema.TypeKind {
	return schema.EnumKind
}

func (e *Enum) GetValues() map[string]schema.EnumValue {
	return e.Values
}

// TODO: enum values has to be a list, return

var _ schema.EnumType = &Enum{}

type EnumValues map[string]schema.EnumValue

type EnumValue struct {
	Description string
	Directives  Directives
	Value       interface{}
}

func (e *EnumValue) GetDescription() string {
	return e.Description
}

func (e *EnumValue) GetDirectives() []schema.Directive {
	return e.Directives
}

func (e *EnumValue) GetValue() interface{} {
	return e.Value
}

var _ schema.EnumValue = &EnumValue{}
