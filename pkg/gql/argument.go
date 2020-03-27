package gql

import (
	"reflect"
)

type Arguments []*Argument

type Argument struct {
	Description  string
	Name         string
	Type         Type
	DefaultValue interface{}
}

func (a *Argument) GetDescirption() string {
	return a.Description
}

func (a *Argument) GetName() string {
	return a.Name
}

func (a *Argument) GetType() Type {
	return a.Type
}

func (a *Argument) GetDefaultValue() interface{} {
	return a.DefaultValue
}

func (a *Argument) IsDefaultValueSet() bool {
	return reflect.ValueOf(a.DefaultValue).IsValid()
}

type InputObject struct {
	Description string
	Name        string
	Directives  Directives
	Fields      InputFields
}

// var _ schema.InputObjectType = &InputObject{}

func (o *InputObject) GetDescription() string {
	return o.Description
}

func (o *InputObject) GetName() string {
	return o.Name
}

func (o *InputObject) GetKind() TypeKind {
	return InputObjectKind
}

func (o *InputObject) GetDirectives() []Directive {
	return o.Directives
}

func (o *InputObject) GetFields() []*InputField {
	return o.Fields
}

type InputField struct {
	Description  string
	Name         string
	Type         Type
	DefaultValue interface{}
	Directives   Directives
}

type InputFields []*InputField

func (o *InputField) GetDescription() string {
	return o.Description
}

func (o *InputField) GetName() string {
	return o.Name
}

func (o *InputField) GetType() Type {
	return o.Type
}

func (o *InputField) GetDefaultValue() interface{} {
	return o.DefaultValue
}

func (o *InputField) IsDefaultValueSet() bool {
	return reflect.ValueOf(o.DefaultValue).IsValid()
}

func (o *InputField) GetDirectives() []Directive {
	return o.Directives
}

// var _ InputField = &InputField{}
