package schema

import (
	"context"
	"fmt"
)

// Directive ..
type Directive struct {
	Name      string
	Arguments []*Argument
}

// Field ...
type Field struct {
	Name              string
	Description       string
	Depricated        bool
	DepricationReason string
	Type              Type
	Arguments         []*Argument
	Resolver          ResolverFunc
}

// ResolverFunc ..
type ResolverFunc func(context.Context, map[string]interface{}, interface{}) (interface{}, error)

// Fields ...
type Fields []*Field

// Get returns the field with the given name (if exists)
func (fs Fields) Get(name string) (*Field, error) {
	for f := range fs {
		if fs[f].Name == name {
			return fs[f], nil
		}
	}
	return nil, fmt.Errorf("there's no field named '%s'", name)
}
