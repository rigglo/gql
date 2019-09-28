package gql

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	String = &Scalar{
		Name: "String",
		Encode: func(v interface{}) (interface{}, error) {
			return fmt.Sprintf("%v", v), nil
		},
		Decode: func(bs []byte) (interface{}, error) {
			return strings.Trim(string(bs), "\""), nil
		},
	}

	Int = &Scalar{
		Name: "Int",
		Encode: func(v interface{}) (interface{}, error) {
			return v.(int), nil
		},
		Decode: func(bs []byte) (interface{}, error) {
			n, err := strconv.Atoi(string(bs))
			if err != nil {
				return nil, err
			} else {
				return n, nil
			}
			return nil, nil
		},
	}

	UnixTime = &Scalar{
		Name: "Timestamp",
		Encode: func(v interface{}) (interface{}, error) {
			return v.(time.Time).Unix(), nil
		},
		Decode: func(bs []byte) (interface{}, error) {
			i, err := strconv.ParseInt(string(bs), 10, 64)
			if err != nil {
				return nil, err
			}
			return time.Unix(i, 0), nil
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

func (s *Scalar) IsNullable() bool {
	return true
}
