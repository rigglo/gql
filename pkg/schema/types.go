package schema

import (
	"context"
)

type TypeKind uint

const (
	ScalarKind TypeKind = iota
	ObjectKind
	InterfaceKind
	UnionKind
	EnumKind
	InputObjectKind
	NonNullKind
	ListKind
)

type Type interface {
	GetName() string
	GetDescription() string
	GetKind() TypeKind
}

type ScalarType interface {
	Type
	GetDirectives() []Directive
	CoerceResult(interface{}) (interface{}, error)
	CoerceInput([]byte) (interface{}, error)
}

type CoerceResultFunc func(interface{}) (interface{}, error)
type CoerceInputFunc func([]byte) (interface{}, error)

type ObjectType interface {
	Type
	GetInterfaces() []InterfaceType
	GetDirectives() []Directive
	GetFields() []Field
}

type InterfaceType interface {
	Type
	GetDirectives() []Directive
	GetFields() []Field
	Resolve(context.Context, interface{}) Type
}

type UnionType interface {
	Type
	GetDirectives() []Directive
	GetMembers() []Type
}

type EnumType interface {
	Type
	GetDirectives() []Directive
	GetValues() map[string]EnumValue
}

type EnumValue interface {
	GetDescription() string
	GetValue() interface{}
	GetDirectives() []Directive
}

type InputObjectType interface {
	Type
	GetDirectives() []Directive
	GetFields() []InputField
}

type InputField interface {
	GetDescription() string
	GetName() string
	GetType() Type
	GetDefaultValue() interface{}
	IsDefaultValueSet() bool
	GetDirectives() []Directive
}

type Field interface {
	GetDescription() string
	GetName() string
	GetArguments() []Argument
	GetType() Type
	GetDirectives() []Directive
	Resolve(context.Context, interface{}, map[string]interface{}) (interface{}, error)
}

type Argument interface {
	GetDescirption() string
	GetName() string
	GetType() Type
	GetDefaultValue() interface{}
	IsDefaultValueSet() bool
}

type NonNull interface {
	Type
	Unwrap() Type
}

type List interface {
	Type
	Unwrap() Type
}
