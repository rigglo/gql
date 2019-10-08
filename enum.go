package gql

import (
	"fmt"
)

type Enum struct {
	Name        string
	Description string
	Values      EnumValues
}

type EnumValues []*EnumValue

type EnumValue struct {
	Name        string
	Value       interface{}
	Description string
}

func (e *Enum) Unwrap() Type {
	return nil
}

func (e *Enum) Kind() TypeDefinition {
	return EnumTypeDefinition
}

func (e *Enum) InputCoercion(s string) (interface{}, error) {
	for _, eV := range e.Values {
		if eV.Name == s {
			return eV.Value, nil
		}
	}
	return nil, fmt.Errorf("invalid key for %s enum: %s", e.Name, s)
}

func (e *Enum) OutputCoercion(v interface{}) ([]byte, error) {
	for _, eV := range e.Values {
		if eV.Value == v {
			return []byte(`"` + eV.Name + `"`), nil
		}
	}
	return nil, fmt.Errorf("invalid value for %s enum: %v", e.Name, v)
}

func (e *Enum) Validate() error {
	return nil
}
