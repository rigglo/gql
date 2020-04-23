package gql

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

var String *Scalar = &Scalar{
	Name:        "String",
	Description: "This is the built-in 'String' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		if m, ok := i.(json.Marshaler); ok {
			v, err := m.MarshalJSON()
			if err != nil {
				return nil, err
			}
			return json.RawMessage(v), nil
		}

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
		case string:
			return i, nil
		}
		return nil, fmt.Errorf("couldn't coerce input value '%v'", i)

	},
}

var ID *Scalar = &Scalar{
	Name:        "ID",
	Description: "This is the built-in 'ID' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		if m, ok := i.(json.Marshaler); ok {
			v, err := m.MarshalJSON()
			if err != nil {
				return nil, err
			}
			return json.RawMessage(v), nil
		}

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
		if m, ok := i.(json.Marshaler); ok {
			v, err := m.MarshalJSON()
			if err != nil {
				return nil, err
			}
			return json.RawMessage(v), nil
		}

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
		case float64:
			if i == float64(int(i.(float64))) {
				return int(i.(float64)), nil
			}
		}
		return nil, fmt.Errorf("couldn't coerce input value '%v'", i)
	},
}

var Float *Scalar = &Scalar{
	Name:        "Float",
	Description: "This is the built-in 'Float' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		if m, ok := i.(json.Marshaler); ok {
			v, err := m.MarshalJSON()
			if err != nil {
				return nil, err
			}
			return json.RawMessage(v), nil
		}
		return i, nil
	},
	CoerceInputFunc: func(i interface{}) (interface{}, error) {
		switch i.(type) {
		case []byte:
			if bs, ok := i.([]byte); ok {
				if n, err := strconv.ParseFloat(string(bs), 64); err == nil {
					return float64(n), nil
				}
			}
		case float64:
			return i, nil
		}
		return nil, fmt.Errorf("couldn't coerce input value '%v'", i)
	},
}

var Boolean *Scalar = &Scalar{
	Name:        "Boolean",
	Description: "This is the built-in 'Boolean' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		if m, ok := i.(json.Marshaler); ok {
			v, err := m.MarshalJSON()
			if err != nil {
				return nil, err
			}
			return json.RawMessage(v), nil
		}
		switch i.(type) {
		case bool:
			return i, nil
		}
		if i == nil {
			return i, nil
		}
		return nil, fmt.Errorf("couldn't coerce value to Boolean")
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
		case bool:
			return i, nil
		}
		return nil, fmt.Errorf("couldn't coerce input value '%v'", i)
	},
}

var DateTime *Scalar = &Scalar{
	Name:        "DateTime",
	Description: "The `DateTime` scalar type represents a DateTime (RFC 3339)",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		switch value := i.(type) {
		case time.Time:
			buff, err := value.MarshalText()
			if err != nil {
				return nil, err
			}
			return string(buff), nil
		case *time.Time:
			if value == nil {
				return nil, nil
			}
			buff, err := value.MarshalText()
			if err != nil {
				return nil, err
			}
			return string(buff), nil
		default:
			return nil, nil
		}

	},
	CoerceInputFunc: func(i interface{}) (interface{}, error) {
		switch value := i.(type) {
		case []byte:
			t := time.Time{}
			err := t.UnmarshalText(value)
			if err != nil {
				return nil, err
			}

			return t, nil
		case string:
			t := time.Time{}
			err := t.UnmarshalText([]byte(value))
			if err != nil {
				return nil, err
			}

			return t, nil
		case *string:
			if value == nil {
				return nil, nil
			}
			t := time.Time{}
			err := t.UnmarshalText([]byte(*value))
			if err != nil {
				return nil, err
			}

			return t, nil
		case time.Time:
			return value, nil
		default:
			return nil, nil
		}
	},
}
