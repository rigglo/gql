package gql

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

func (e *Enum) GetDirectives() []*Directive {
	return e.Directives
}

func (e *Enum) GetKind() TypeKind {
	return EnumKind
}

func (e *Enum) GetValues() []*EnumValue {
	return e.Values
}

// TODO: enum values has to be a list, return

// var _ EnumType = &Enum{}

type EnumValues []*EnumValue

type EnumValue struct {
	Name        string
	Description string
	Directives  Directives
	Value       interface{}
}

func (e EnumValue) GetDescription() string {
	return e.Description
}

func (e EnumValue) GetDirectives() []*Directive {
	return e.Directives
}

func (e EnumValue) GetValue() interface{} {
	return e.Value
}

// var _ EnumValue = &EnumValue{}
