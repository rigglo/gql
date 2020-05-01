package gql

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/rigglo/gql/pkg/language/ast"
)

var String *Scalar = &Scalar{
	Name:        "String",
	Description: "This is the built-in 'String' scalar type",
	CoerceResultFunc: func(i interface{}) (interface{}, error) {
		if m, ok := i.(json.Marshaler); ok {
			return m, nil
		}
		if v, ok := i.(string); ok {
			return v, nil
		} else if v, ok := i.(*string); ok {
			if v == nil {
				return nil, nil
			}
			return *v, nil
		} else if v, ok := i.([]byte); ok {
			if v == nil {
				return nil, nil
			}
			return string(v), nil
		}
		return fmt.Sprintf("%v", i), nil
	},
	CoerceInputFunc: func(i interface{}) (interface{}, error) {
		switch i := i.(type) {
		case ast.Value:
			if v, ok := i.(*ast.StringValue); ok {
				return trimString(v.Value), nil
			}
			return nil, errors.New("invalid value for String scalar")
		case string:
			return i, nil
		case *string:
			return *i, nil
		case []byte:
			return string(i), nil
		}
		return nil, errors.New("invalid value for String scalar")
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
			return coerceString(trimString(string(v)), false)
		}
		return coerceString(i, false)
	},
	CoerceInputFunc: func(i interface{}) (interface{}, error) {
		if v, ok := i.(ast.Value); ok {
			if sv, ok := v.(*ast.StringValue); ok {
				return trimString(sv.Value), nil
			} else if iv, ok := v.(*ast.IntValue); ok {
				return coerceInt(iv)
			}
		} else if i, err := coerceInt(i); err == nil {
			return i, nil
		}
		return coerceString(i, true)
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
		return serializeDateTime(i)
	},
	CoerceInputFunc: unserializeDateTime,
}

func trimString(value string) string {
	return strings.Trim(value, `"`)
}

func unserializeDateTime(value interface{}) (interface{}, error) {
	switch value := value.(type) {
	case ast.Value:
		if v, ok := value.(*ast.StringValue); ok {
			return unserializeDateTime(trimString(v.Value))
		}
		return nil, errors.New("invalid value for DateTime scalar")
	case time.Time:
		return value, nil
	case *time.Time:
		if value == nil {
			return nil, nil
		}
		return *value, nil
	case string:
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return nil, errors.New("invalid value for DateTime scalar")
		}
		return t, nil
	}
	return nil, errors.New("invalid value for DateTime scalar")
}

func serializeDateTime(value interface{}) (interface{}, error) {
	switch value := value.(type) {
	case time.Time:
		buff, err := value.MarshalText()
		if err != nil {
			return nil, errors.New("invalid value for DateTime scalar")
		}
		return string(buff), nil
	case *time.Time:
		if value == nil {
			return nil, nil
		}
		buff, err := value.MarshalText()
		if err != nil {
			return nil, errors.New("invalid value for DateTime scalar")
		}
		return string(buff), nil
	case int64:
		return serializeDateTime(time.Unix(value, 0))
	}
	return nil, errors.New("invalid value for DateTime scalar")
}

func coerceBool(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	switch value := value.(type) {
	case ast.Value:
		if v, ok := value.(*ast.BooleanValue); ok {
			return coerceBool(v.Value)
		}
		return nil, errors.New("invalid value for Boolean scalar")
	case bool:
		return value, nil
	case *bool:
		if value == nil {
			return nil, nil
		}
		return *value, nil
	case string:
		if value == `false` {
			return false, nil
		} else if value == `true` {
			return true, nil
		}
	}
	return nil, errors.New("invalid value for Boolean scalar")
}

func coerceInt(value interface{}) (interface{}, error) {
	switch value := value.(type) {
	case ast.Value:
		if v, ok := value.(*ast.IntValue); ok {
			return coerceInt(v.Value)
		}
		return nil, errors.New("invalid value for Int scalar")
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
		} else if value != float32(int(value)) {
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
		} else if value != float64(int(value)) {
			return nil, errors.New("invalid int32 value")
		}
		return int(value), nil
	case *float64:
		if value == nil {
			return nil, nil
		}
		return coerceInt(*value)
	case string:
		val, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, errors.New("invalid int32 value")
		}
		return int(val), nil
	case *string:
		if value == nil {
			return nil, nil
		}
		return coerceInt(*value)
	}
	return nil, errors.New("invalid value for Int")
}

func coerceFloat(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	switch value := value.(type) {
	case ast.Value:
		if v, ok := value.(*ast.IntValue); ok {
			return coerceFloat(v.Value)
		} else if v, ok := value.(*ast.FloatValue); ok {
			return coerceFloat(v.Value)
		}
		return nil, errors.New("invalid value for Float scalar")
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
	return nil, errors.New("invalid float value")
}

func coerceString(value interface{}, input bool) (interface{}, error) {
	if v, ok := value.(string); ok {
		return v, nil
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
	} else if input {
		return nil, errors.New("invalid value for String scalar")
	}
	return fmt.Sprintf("%v", value), nil
}
