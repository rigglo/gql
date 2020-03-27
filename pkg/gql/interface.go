package gql

import (
	"context"
)

type Interfaces []*Interface

type Interface struct {
	Description  string
	Name         string
	Directives   Directives
	Fields       Fields
	TypeResolver TypeResolver
}

// type ObjectType ObjectType

type TypeResolver func(context.Context, interface{}) *Object

func (i *Interface) GetDescription() string {
	return i.Description
}

func (i *Interface) GetName() string {
	return i.Name
}

func (i *Interface) GetKind() TypeKind {
	return InterfaceKind
}

func (i *Interface) GetDirectives() []*Directive {
	return i.Directives
}

func (i *Interface) GetFields() []*Field {
	return i.Fields
}

func (i *Interface) Resolve(ctx context.Context, v interface{}) *Object {
	return i.TypeResolver(ctx, v)
}

// var _ InterfaceType = &Interface{}
