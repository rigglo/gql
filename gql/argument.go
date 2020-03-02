package gql

import (
	"reflect"

	"github.com/rigglo/gql/schema"
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
