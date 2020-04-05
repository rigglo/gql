package gql

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type CoerceResultFunc func(interface{}) (interface{}, error)
type CoerceInputFunc func(interface{}) (interface{}, error)

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

func (s *Scalar) GetKind() TypeKind {
	return ScalarKind
}

func (s *Scalar) GetDirectives() []Directive {
	return s.Directives
}

func (s *Scalar) CoerceResult(i interface{}) (interface{}, error) {
	return s.CoerceResultFunc(i)
}

func (s *Scalar) CoerceInput(bs interface{}) (interface{}, error) {
	return s.CoerceInputFunc(bs)
}

var String *Scalar = &Scalar{
	Name:        "String",
	Description: "This is the built-in 'String' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		if i == nil {
			return nil, nil
		}
		switch i.(type) {
		case string:
			return i.(string), nil
		default:
			return fmt.Sprintf("%v", i), nil
		}
	},
	CoerceInputFunc: func(i interface{}) (interface{}, error) {
		switch i.(type) {
		case []byte:
			if bs, ok := i.([]byte); ok {
				if !strings.HasPrefix(string(bs), "\"") || !strings.HasSuffix(string(bs), "\"") {
					return nil, errors.New("Invalid input, could not coerce value")
				}
				return strings.TrimPrefix(strings.TrimSuffix(string(bs), "\""), "\""), nil
			}
		}
		return nil, fmt.Errorf("couldn't coerce input value '%v'", i)

	},
}

var ID *Scalar = &Scalar{
	Name:        "ID",
	Description: "This is the built-in 'ID' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		if i == nil {
			return nil, nil
		}
		switch i.(type) {
		case string:
			return i.(string), nil
		default:
			return fmt.Sprintf("%v", i), nil
		}
	},
	CoerceInputFunc: func(i interface{}) (interface{}, error) {
		switch i.(type) {
		case []byte:
			if bs, ok := i.([]byte); ok {
				if !strings.HasPrefix(string(bs), "\"") || !strings.HasSuffix(string(bs), "\"") {
					return nil, errors.New("Invalid input, could not coerce value")
				}
				return strings.TrimPrefix(strings.TrimSuffix(string(bs), "\""), "\""), nil
			}
		}
		return nil, fmt.Errorf("couldn't coerce input value '%v'", i)
	},
}

var Int *Scalar = &Scalar{
	Name:        "Int",
	Description: "This is the built-in 'Int' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		switch i.(type) {
		case float32:
			{
				if i == float32(int(i.(float32))) {
					return int(i.(float32)), nil
				}
				return nil, errors.New("Data loss, expected Int, got float32")
			}
		case float64:
			{
				if i == float64(int(i.(float64))) {
					return int(i.(float64)), nil
				}
				return nil, errors.New("Data loss, expected Int, got float64")
			}
		case string:
			{
				if n, err := strconv.ParseInt(i.(string), 10, 32); err == nil {
					return int(n), nil
				}
				return nil, errors.New("Data loss, expected Int, got string")
			}
		case bool:
			{
				if i.(bool) {
					return 1, nil
				}
				return 0, nil
			}
		case int:
			{
				return i.(int), nil
			}
		case int8:
			{
				return int(i.(int8)), nil
			}
		case int16:
			{
				return int(i.(int16)), nil
			}
		case int32:
			{
				return int(i.(int32)), nil
			}
		case int64:
			{
				if i.(int64) > math.MaxInt32 || i.(int64) < math.MinInt32 {
					return nil, errors.New("Data loss, expected Int (32bit), got int64")
				}
				return int(i.(int64)), nil
			}
		case uint:
			{
				if i.(uint) > math.MaxInt32 {
					return nil, errors.New("Data loss, expected Int (32bit), got uint")
				}
				return int(i.(uint)), nil
			}
		case uint8:
			{
				return int(i.(uint8)), nil
			}
		case uint16:
			{
				return int(i.(uint16)), nil
			}
		case uint32:
			{
				if i.(uint32) > math.MaxInt32 {
					return nil, errors.New("Data loss, expected Int (32bit), got uint32")
				}
				return int(i.(int64)), nil
			}
		case uint64:
			{
				if i.(uint64) > math.MaxInt32 {
					return nil, errors.New("Data loss, expected Int (32bit), got uint64")
				}
				return int(i.(uint64)), nil
			}
		}
		return i, nil
	},
	CoerceInputFunc: func(i interface{}) (interface{}, error) {
		switch i.(type) {
		case []byte:
			if bs, ok := i.([]byte); ok {
				if n, err := strconv.ParseInt(string(bs), 10, 32); err == nil {
					return int(n), nil
				}
			}
		}
		return nil, fmt.Errorf("couldn't coerce input value '%v'", i)
	},
}

var Float *Scalar = &Scalar{
	Name:        "Float",
	Description: "This is the built-in 'Float' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		return i, nil
	},
	CoerceInputFunc: func(i interface{}) (interface{}, error) {
		switch i.(type) {
		case []byte:
			if bs, ok := i.([]byte); ok {
				if n, err := strconv.ParseFloat(string(bs), 32); err == nil {
					return float32(n), nil
				}
			}
		}
		return nil, fmt.Errorf("couldn't coerce input value '%v'", i)
	},
}

var Boolean *Scalar = &Scalar{
	Name:        "Boolean",
	Description: "This is the built-in 'Boolean' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		return i, nil
	},
	CoerceInputFunc: func(i interface{}) (interface{}, error) {
		switch i.(type) {
		case []byte:
			if bs, ok := i.([]byte); ok {
				if string(bs) == "true" {
					return true, nil
				} else if string(bs) == "false" {
					return false, nil
				}
			}
		}
		return nil, fmt.Errorf("couldn't coerce input value '%v'", i)
	},
}
