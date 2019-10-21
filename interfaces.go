package gql

import (
	"context"
	"fmt"
	"strings"
)

// Interface ...
type Interface struct {
	Name        string
	Description string
	Fields      *Fields
	ResolveType TypeResolver
}

// TypeResolver resolves the Object for an Interface or Union based on the resolver's result
type TypeResolver func(ctx context.Context, value interface{}) *Object

// Kind returns InterfaceTypeDefinition
func (i *Interface) Kind() TypeDefinition {
	return InterfaceTypeDefinition
}

// Unwrap returns nothing for an Interface
func (i *Interface) Unwrap() Type {
	return nil
}

// Validate the type
func (i *Interface) Validate(ctx *ValidationContext) error {
	if strings.HasPrefix(i.Name, "__") {
		return fmt.Errorf("invalid name (%s) for Object", i.Name)
	}

	if i.Fields.Len() == 0 {
		return fmt.Errorf("object (%s) must have at least one field", i.Name)
	}
	fieldNames := map[string]int{}
	for n, f := range i.Fields.Slice() {
		if _, ok := fieldNames[f.Name]; ok {
			return fmt.Errorf("field names must be unique, there are more than one fields named '%s'", f.Name)
		}
		fieldNames[f.Name] = n
		if err := f.Validate(ctx); err != nil {
			return err
		}
	}
	ctx.types[i.Name] = i
	return nil
}

// GetFields implemets the HasFields interface
func (i *Interface) GetFields() *Fields {
	return i.Fields
}
