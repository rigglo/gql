package gql

import "github.com/rigglo/gql/pkg/schema"

type Object struct {
	Description string
	Name        string
	Implements  []interface{}
	Directives  Directives
	Fields      Fields
}

func (o *Object) GetDescription() string {
	return o.Description
}

func (o *Object) GetName() string {
	return o.Name
}

func (o *Object) GetKind() schema.TypeKind {
	return schema.ObjectKind
}

func (o *Object) GetImplements() []schema.InterfaceType {
	return nil
}

func (o *Object) GetDirectives() []schema.Directive {
	return o.Directives
}

func (o *Object) GetFields() []schema.Field {
	return o.Fields
}
