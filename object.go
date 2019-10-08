package gql

import (
	"fmt"
	"strings"
)

// Object ...
type Object struct {
	Name        string
	Description string
	Implements  []Type
	Fields      Fields
	fieldNames  map[string]int
}

// Kind returns ObjectTypeDefinition
func (o *Object) Kind() TypeDefinition {
	return ObjectTypeDefinition
}

// Unwrap returns nothing for an Object
func (o *Object) Unwrap() Type {
	return nil
}

// Validate returns any error caused by invalid object definition
func (o *Object) Validate() error {
	if strings.HasPrefix(o.Name, "__") {
		return fmt.Errorf("invalid name (%s) for Object", o.Name)
	}
	if len(o.Fields) == 0 {
		return fmt.Errorf("object (%s) must have at least one field", o.Name)
	}
	o.fieldNames = map[string]int{}
	for i, f := range o.Fields {
		if _, ok := o.fieldNames[f.Name]; ok {
			return fmt.Errorf("field names must be unique, there are more than one fields named '%s'", f.Name)
		}
		o.fieldNames[f.Name] = i
		if err := f.Validate(); err != nil {
			return err
		}
	}
	return nil
}
