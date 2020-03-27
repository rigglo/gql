package gql

import (
	"reflect"
)

type Object struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Implements  Interfaces
	Directives  Directives
	Fields      Fields
}

func (o *Object) GetDescription() string {
	return o.Description
}

func (o *Object) GetName() string {
	return o.Name
}

func (o *Object) GetKind() TypeKind {
	return ObjectKind
}

func (o *Object) GetInterfaces() []*Interface {
	return o.Implements
}

func (o *Object) GetDirectives() []Directive {
	return o.Directives
}

func (o *Object) GetFields() []*Field {
	return o.Fields
}

func (o *Object) AddFields(fs ...*Field) {
	o.Fields = append(o.Fields, fs...)
}

func (o *Object) DoesImplement(i *Interface) bool {
	for _, in := range o.Implements {
		if reflect.DeepEqual(i, in) {
			return true
		}
	}
	return false
}
