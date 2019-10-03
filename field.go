package gql

import (
	"context"
	"fmt"
)

type Argument struct {
	Name              string
	Description       string
	Depricated        bool
	DepricationReason string
	Type              Type
}

type Directive struct {
	Name      string
	Arguments []*Argument
}

type Field struct {
	Name              string
	Description       string
	Depricated        bool
	DepricationReason string
	Type              Type
	Arguments         []*Argument
	Resolver          ResolverFunc
}

type ResolverFunc func(context.Context, map[string]interface{}, interface{}) (interface{}, error)

type Fields []*Field

func (fs Fields) Get(name string) (*Field, error) {
	for f := range fs {
		if fs[f].Name == name {
			return fs[f], nil
		}
	}
	return nil, fmt.Errorf("there's no field named '%s'", name)
}
