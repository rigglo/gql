package gql

import (
	"context"

	"github.com/rigglo/gql/pkg/schema"
)

type Interfaces []schema.InterfaceType

type Interface struct {
	Description  string
	Name         string
	Directives   Directives
	Fields       Fields
	TypeResolver TypeResolver
}

type ObjectType schema.ObjectType

type TypeResolver func(context.Context, interface{}) ObjectType

func (i *Interface) GetDescription() string {
	return i.Description
}

func (i *Interface) GetName() string {
	return i.Name
}

func (i *Interface) GetKind() schema.TypeKind {
	return schema.InterfaceKind
}

func (i *Interface) GetDirectives() []schema.Directive {
	return i.Directives
}

func (i *Interface) GetFields() []schema.Field {
	return i.Fields
}

func (i *Interface) Resolve(ctx context.Context, v interface{}) schema.ObjectType {
	return i.TypeResolver(ctx, v)
}

var _ schema.InterfaceType = &Interface{}
