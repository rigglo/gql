package gql

import (
	"reflect"

	"github.com/rigglo/gql/pkg/schema"
)

type Arguments []schema.Argument

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

func (a *Argument) GetType() schema.Type {
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

var _ schema.InputObjectType = &InputObject{}

func (o *InputObject) GetDescription() string {
	return o.Description
}

func (o *InputObject) GetName() string {
	return o.Name
}

func (o *InputObject) GetKind() schema.TypeKind {
	return schema.InputObjectKind
}

func (o *InputObject) GetDirectives() []schema.Directive {
	return o.Directives
}

func (o *InputObject) GetFields() map[string]schema.InputField {
	return o.Fields
}

type InputField struct {
	Description  string
	Name         string
	Type         Type
	DefaultValue interface{}
	Directives   Directives
}

type InputFields map[string]schema.InputField

func (o *InputField) GetDescription() string {
	return o.Description
}

func (o *InputField) GetName() string {
	return o.Name
}

func (o *InputField) GetType() schema.Type {
	return o.Type
}

func (o *InputField) GetDefaultValue() interface{} {
	return o.DefaultValue
}

func (o *InputField) IsDefaultValueSet() bool {
	return reflect.ValueOf(o.DefaultValue).IsValid()
}

func (o *InputField) GetDirectives() []schema.Directive {
	return o.Directives
}

var _ schema.InputField = &InputField{}
