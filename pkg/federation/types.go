package federation

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rigglo/gql"
	"github.com/rigglo/gql/pkg/language/ast"
)

var (
	anyScalar *gql.Scalar = &gql.Scalar{
		Name:        "_Any",
		Description: "This is the built-in '_Any' scalar type for federation",
		CoerceResultFunc: func(i interface{}) (interface{}, error) {
			switch i := i.(type) {
			case json.Marshaler:
				return i, nil
			case string, *string:
				return i, nil
			case []byte:
				return string(i), nil
			case fmt.Stringer:
				return i.String(), nil
			default:
				return fmt.Sprintf("%v", i), nil
			}
		},
		CoerceInputFunc: func(i interface{}) (interface{}, error) {
			switch i := i.(type) {
			case map[string]interface{}:
				return i, nil
			default:
				return nil, fmt.Errorf("invalid value for _Any scalar, got type: '%T'", i)
			}
		},
		AstValidator: func(v ast.Value) error {
			switch v.(type) {
			case *ast.ObjectValue:
				return nil
			default:
				return errors.New("invalid value type for _Any scalar")
			}
		},
	}

	fieldSetScalar *gql.Scalar = &gql.Scalar{
		Name:        "_FieldSet",
		Description: "This is the built-in '_FieldSet' scalar type",
		CoerceResultFunc: func(i interface{}) (interface{}, error) {
			switch i := i.(type) {
			case json.Marshaler:
				return i, nil
			case string, *string:
				return i, nil
			case []byte:
				return string(i), nil
			case fmt.Stringer:
				return i.String(), nil
			default:
				return fmt.Sprintf("%v", i), nil
			}
		},
		CoerceInputFunc: func(i interface{}) (interface{}, error) {
			switch i := i.(type) {
			case string:
				return i, nil
			default:
				return nil, fmt.Errorf("invalid value for _FieldSet scalar, got type: '%T'", i)
			}
		},
		AstValidator: func(v ast.Value) error {
			switch v.(type) {
			case *ast.StringValue:
				return nil
			default:
				return errors.New("invalid value type for _FieldSet scalar")
			}
		},
	}
)
