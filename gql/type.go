package gql

import "github.com/rigglo/gql/schema"

type Type interface {
	GetName() string
	GetDescription() string
	GetKind() schema.TypeKind
}

type List struct {
	Name        string
	Description string
	Wrapped     Type
}

func NewList(t Type) Type {
	return &List{
		Name:        "List",
		Description: "Built-in 'List' type",
		Wrapped:     t,
	}
}

func (l *List) GetName() string {
	return l.Name
}

func (l *List) GetDescription() string {
	return l.Description
}

func (l *List) GetKind() schema.TypeKind {
	return schema.ListKind
}

func (l *List) Unwrap() schema.Type {
	return l.Wrapped
}

var _ schema.List = &List{}
