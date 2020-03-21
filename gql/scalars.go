package gql

import (
	"fmt"

	"github.com/rigglo/gql/schema"
)

type CoerceResultFunc schema.CoerceResultFunc
type CoerceInputFunc schema.CoerceInputFunc

type Scalar struct {
	Name             string
	Description      string
	Directives       Directives
	CoerceResultFunc CoerceResultFunc
	CoerceInputFunc  CoerceInputFunc
}

func (s *Scalar) GetName() string {
	return s.Name
}

func (s *Scalar) GetDescription() string {
	return s.Description
}

func (s *Scalar) GetKind() schema.TypeKind {
	return schema.ScalarKind
}

func (s *Scalar) GetDirectives() []schema.Directive {
	return s.Directives
}

func (s *Scalar) CoerceResult(i interface{}) (interface{}, error) {
	return s.CoerceResultFunc(i)
}

func (s *Scalar) CoerceInput(bs []byte) (interface{}, error) {
	return s.CoerceInputFunc(bs)
}

var String *Scalar = &Scalar{
	Name:        "String",
	Description: "This is the built-in 'String' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		if i == nil {
			return nil, nil
		}
		return fmt.Sprintf("%v", i), nil
	},
	CoerceInputFunc: func(bs []byte) (interface{}, error) {
		return string(bs), nil
	},
}

var Int *Scalar = &Scalar{
	Name:        "Int",
	Description: "This is the built-in 'Int' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		return i, nil
	},
	CoerceInputFunc: func(bs []byte) (interface{}, error) {
		return string(bs), nil
	},
}

var Float *Scalar = &Scalar{
	Name:        "Float",
	Description: "This is the built-in 'Float' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		return i, nil
	},
	CoerceInputFunc: func(bs []byte) (interface{}, error) {
		return string(bs), nil
	},
}

var Boolean *Scalar = &Scalar{
	Name:        "Boolean",
	Description: "This is the built-in 'Boolean' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		return i, nil
	},
	CoerceInputFunc: func(bs []byte) (interface{}, error) {
		return string(bs), nil
	},
}
