package gql

import (
	"fmt"
	"strconv"
)

var (
	String = &Scalar{
		Name: "String",
		Encode: func(v interface{}) (interface{}, error) {
			return fmt.Sprintf("%v", v), nil
		},
		Decode: func(bs []byte) (interface{}, error) {
			return string(bs), nil
		},
	}

	Int = &Scalar{
		Name: "Int",
		Encode: func(v interface{}) (interface{}, error) {
			return v.(int), nil
		},
		Decode: func(bs []byte) (interface{}, error) {
			return strconv.Atoi(string(bs))
		},
	}
)

type Encoder func(interface{}) (interface{}, error)
type Decoder func([]byte) (interface{}, error)

type Scalar struct {
	Name   string
	Encode Encoder
	Decode Decoder
}

func (s *Scalar) Type() FieldType {
	return ScalarType
}

func (s *Scalar) Value() interface{} {
	return s
}
