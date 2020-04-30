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
			return coerceString(v)
		}
		return coerceString(i)
	},
	CoerceInputFunc: coerceString,
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
			return coerceString(v)
		}
		return coerceString(i)
	},
	CoerceInputFunc: func(i interface{}) (interface{}, error) {
		if bs, ok := i.([]byte); ok {
			var str string
			err := json.Unmarshal(bs, &str)
			if err != nil {
				return nil, err
			}
			i = str
		} else if v, err := coerceInt(i); err == nil {
			return v, nil
		}
		return coerceString(i)
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
			return coerceInt(v)
		}
		if i == nil {
			return nil, nil
		}

		return coerceInt(i)
	},
	CoerceInputFunc: coerceInt,
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
			return coerceFloat(v)
		}
		return coerceFloat(i)
	},
	CoerceInputFunc: coerceFloat,
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
			return coerceBool(v)
		}
		return coerceBool(i)
	},
	CoerceInputFunc: coerceBool,
}

var DateTime *Scalar = &Scalar{
	Name:        "DateTime",
	Description: "The `DateTime` scalar type represents a DateTime (RFC 3339)",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		if m, ok := i.(json.Marshaler); ok {
			return m, nil
		}

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

func trimString(value string) string {
	return strings.Trim(value, `"`)
}

func coerceBool(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	switch value := value.(type) {
	case bool:
		return value, nil
	case *bool:
		if value == nil {
			return nil, nil
		}
		return *value, nil
	case string:
		switch value {
		case "", "false":
			return false, nil
		}
		return true, nil
	case *string:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	case float64:
		if value != 0 {
			return true, nil
		}
		return false, nil
	case *float64:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	case float32:
		if value != 0 {
			return true, nil
		}
		return false, nil
	case *float32:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	case int:
		if value != 0 {
			return true, nil
		}
		return false, nil
	case *int:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	case int8:
		if value != 0 {
			return true, nil
		}
		return false, nil
	case *int8:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	case int16:
		if value != 0 {
			return true, nil
		}
		return false, nil
	case *int16:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	case int32:
		if value != 0 {
			return true, nil
		}
		return false, nil
	case *int32:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	case int64:
		if value != 0 {
			return true, nil
		}
		return false, nil
	case *int64:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	case uint:
		if value != 0 {
			return true, nil
		}
		return false, nil
	case *uint:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	case uint8:
		if value != 0 {
			return true, nil
		}
		return false, nil
	case *uint8:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	case uint16:
		if value != 0 {
			return true, nil
		}
		return false, nil
	case *uint16:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	case uint32:
		if value != 0 {
			return true, nil
		}
		return false, nil
	case *uint32:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	case uint64:
		if value != 0 {
			return true, nil
		}
		return false, nil
	case *uint64:
		if value == nil {
			return nil, nil
		}
		return coerceBool(*value)
	}
	return false, nil
}

func coerceInt(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	switch value := value.(type) {
	case bool:
		if value == true {
			return 1, nil
		}
		return 0, nil
	case *bool:
		if value == nil {
			return nil, nil
		}
		return coerceInt(*value)
	case int:
		if value < int(math.MinInt32) || value > int(math.MaxInt32) {
			return nil, errors.New("invalid int32 value")
		}
		return value, nil
	case *int:
		if value == nil {
			return nil, nil
		}
		return coerceInt(*value)
	case int8:
		return int(value), nil
	case *int8:
		if value == nil {
			return nil, nil
		}
		return int(*value), nil
	case int16:
		return int(value), nil
	case *int16:
		if value == nil {
			return nil, nil
		}
		return int(*value), nil
	case int32:
		return int(value), nil
	case *int32:
		if value == nil {
			return nil, nil
		}
		return int(*value), nil
	case int64:
		if value < int64(math.MinInt32) || value > int64(math.MaxInt32) {
			return nil, errors.New("invalid int32 value")
		}
		return int(value), nil
	case *int64:
		if value == nil {
			return nil, nil
		}
		return coerceInt(*value)
	case uint:
		if value > math.MaxInt32 {
			return nil, errors.New("invalid int32 value")
		}
		return int(value), nil
	case *uint:
		if value == nil {
			return nil, nil
		}
		return coerceInt(*value)
	case uint8:
		return int(value), nil
	case *uint8:
		if value == nil {
			return nil, nil
		}
		return int(*value), nil
	case uint16:
		return int(value), nil
	case *uint16:
		if value == nil {
			return nil, nil
		}
		return int(*value), nil
	case uint32:
		if value > uint32(math.MaxInt32) {
			return nil, errors.New("invalid int32 value")
		}
		return int(value), nil
	case *uint32:
		if value == nil {
			return nil, nil
		}
		return coerceInt(*value)
	case uint64:
		if value > uint64(math.MaxInt32) {
			return nil, errors.New("invalid int32 value")
		}
		return int(value), nil
	case *uint64:
		if value == nil {
			return nil, nil
		}
		return coerceInt(*value)
	case float32:
		if value < float32(math.MinInt32) || value > float32(math.MaxInt32) {
			return nil, errors.New("invalid int32 value")
		}
		return int(value), nil
	case *float32:
		if value == nil {
			return nil, nil
		}
		return coerceInt(*value)
	case float64:
		if value < float64(math.MinInt32) || value > float64(math.MaxInt32) {
			return nil, errors.New("invalid int32 value")
		}
		return int(value), nil
	case *float64:
		if value == nil {
			return nil, nil
		}
		return coerceInt(*value)
	case string:
		val, err := strconv.ParseFloat(value, 0)
		if err != nil {
			return nil, errors.New("invalid int32 value")
		}
		return coerceInt(val)
	case *string:
		if value == nil {
			return nil, nil
		}
		return coerceInt(*value)
	case []byte:
		if value == nil {
			return nil, nil
		}
		return coerceInt(string(value))
	}
	return nil, errors.New("invalid int32 value")
}

func coerceFloat(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	switch value := value.(type) {
	case bool:
		if value == true {
			return 1.0, nil
		}
		return 0.0, nil
	case *bool:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case int:
		return float64(value), nil
	case *int:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case int8:
		return float64(value), nil
	case *int8:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case int16:
		return float64(value), nil
	case *int16:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case int32:
		return float64(value), nil
	case *int32:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case int64:
		return float64(value), nil
	case *int64:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case uint:
		return float64(value), nil
	case *uint:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case uint8:
		return float64(value), nil
	case *uint8:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case uint16:
		return float64(value), nil
	case *uint16:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case uint32:
		return float64(value), nil
	case *uint32:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case uint64:
		return float64(value), nil
	case *uint64:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case float32:
		return float64(value), nil
	case *float32:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case float64:
		return value, nil
	case *float64:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case string:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, errors.New("invalid float value")
		}
		return val, nil
	case *string:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(*value)
	case []byte:
		if value == nil {
			return nil, nil
		}
		return coerceFloat(string(value))
	}
	return nil, nil
}

func coerceString(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	} else if v, ok := value.(*string); ok {
		if v == nil {
			return nil, nil
		}
		return *v, nil
	} else if v, ok := value.([]byte); ok {
		if v == nil {
			return nil, nil
		}
		return string(v), nil
	}
	return fmt.Sprintf("%v", value), nil
}
