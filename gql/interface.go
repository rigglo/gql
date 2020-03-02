package gql

import (
	"github.com/rigglo/gql/schema"
)

type Interface struct {
	Description string
	Name        string
	Directives  Directives
	Fields      Fields
}

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
