package gql

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

// InputCoercion func
type InputCoercion func(string) (interface{}, error)

// OutputCoercion func
type OutputCoercion func(interface{}) ([]byte, error)

// Scalar ...
type Scalar struct {
	Name           string
	Description    string
	InputCoercion  InputCoercion
	OutputCoercion OutputCoercion
}

// Unwrap ...
func (s *Scalar) Unwrap() Type {
	return nil
}

// Kind ...
func (s *Scalar) Kind() TypeDefinition {
	return ScalarTypeDefinition
}

var (
	// String is a built-in type in GraphQL
	String = &Scalar{
		Name:        "String",
		Description: "This is the built-in String scalar",
		InputCoercion: func(s string) (interface{}, error) {
			if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
				t := s[1 : len(s)-1]
				if utf8.ValidString(t) {
					return t, nil
				}
			}
			return nil, fmt.Errorf("value '%s' couldn't be coerced as input for String", s)
		},
		OutputCoercion: func(v interface{}) ([]byte, error) {
			r := reflect.ValueOf(v)
			switch {
			case v == nil:
				return []byte("null"), nil
			case r.Kind() == reflect.String:
				if utf8.ValidString(r.String()) {
					return json.Marshal(r.String())
				}
			case r.Kind() == reflect.Bool:
				if r.Bool() {
					return []byte("true"), nil
				}
				return []byte("false"), nil
			case r.Kind() == reflect.Int || r.Kind() == reflect.Int16 || r.Kind() == reflect.Int32:
				return []byte(fmt.Sprintf(`"%v"`, r.Int())), nil
			case r.Kind() == reflect.Float32 || r.Kind() == reflect.Float64:
				return []byte(fmt.Sprintf(`"%v"`, r.Float())), nil
			}
			return nil, fmt.Errorf("value '%v' couldn't be coerced as output for String", v)
		},
	}
	// ID is a built-in type in GraphQL
	ID = &Scalar{
		Name:        "ID",
		Description: "This is the built-in ID scalar",
		InputCoercion: func(s string) (interface{}, error) {
			if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
				t := s[1 : len(s)-1]
				if utf8.ValidString(t) {
					return t, nil
				}
			}
			return nil, fmt.Errorf("value '%s' couldn't be coerced as input for String", s)
		},
		OutputCoercion: func(v interface{}) ([]byte, error) {
			r := reflect.ValueOf(v)
			switch {
			case v == nil:
				return []byte("null"), nil
			case r.Kind() == reflect.String:
				if utf8.ValidString(r.String()) {
					return json.Marshal(r.String())
				}
			case r.Kind() == reflect.Int || r.Kind() == reflect.Int16 || r.Kind() == reflect.Int32:
				return []byte(fmt.Sprintf(`"%v"`, r.Int())), nil
			}
			return nil, fmt.Errorf("value '%v' couldn't be coerced as output for String", v)
		},
	}
)
