# Type Definitions

## Scalars

Scalar is an end type, a type of a field that's either a built-in `Int`, `Float`, `String`, `ID`, `Boolean`, `DateTime` or a custom scalar of yours.

Creating a custom scalar 

```go
var UnixTimestampScalar *gql.Scalar = &gql.Scalar{
    Name:        "UnixTimestamp",
    Description: "This is a custom scalar that converts time.Time into a unix timestamp and a string formatted unix timestamp to time.Time",
    CoerceResultFunc: func(i interface{}) (interface{}, error) {
        switch i.(type) {
        case time.Time:
            return fmt.Sprintf("%v", i.(time.Time).Unix()), nil
        default:
            return nil, errors.New("invalid value to coerce")
        }
    },
    CoerceInputFunc: func(i interface{}) (interface{}, error) {
        switch i.(type) {
        case string:
            unix, err := strconv.ParseInt(i.(string), 10, 64)
            if err != nil {
                return nil, err
            }
            return time.Unix(unix, 0), nil
        default:
            return nil, errors.New("invalid value for UnixTimestamp scalar")
        }
    },
	AstValidator: func(v ast.Value) error {
		switch v.(type) {
		case *ast.StringValue:
			return nil
		default:
			return errors.New("invalid value type for String scalar")
		}
	},
}
```