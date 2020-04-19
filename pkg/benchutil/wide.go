package benchutil

import (
	"fmt"

	"github.com/rigglo/gql"
)

func WideSchemaWithXFieldsAndYItems(x int, y int) *gql.Schema {
	wide := &gql.Object{
		Name:        "Wide",
		Description: "An object",
		Fields:      generateXWideFields(x),
	}

	queryType := &gql.Object{
		Name: "Query",
		Fields: gql.Fields{
			"wide": {
				Type: gql.NewList(wide),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					out := make([]struct{}, 0, y)
					for i := 0; i < y; i++ {
						out = append(out, struct{}{})
					}
					return out, nil
				},
			},
		},
	}

	wideSchema := &gql.Schema{
		Query: queryType,
	}

	return wideSchema
}

func generateXWideFields(x int) gql.Fields {
	fields := gql.Fields{}
	for i := 0; i < x; i++ {
		fn, f := generateWideFieldFromX(i)
		fields[fn] = f
	}
	return fields
}

func generateWideFieldFromX(x int) (string, *gql.Field) {
	return generateFieldNameFromX(x), &gql.Field{
		Type:     generateWideTypeFromX(x),
		Resolver: generateWideResolveFromX(x),
	}
}

func generateWideTypeFromX(x int) gql.Type {
	switch x % 8 {
	case 0:
		return gql.String
	case 1:
		return gql.NewNonNull(gql.String)
	case 2:
		return gql.Int
	case 3:
		return gql.NewNonNull(gql.Int)
	case 4:
		return gql.Float
	case 5:
		return gql.NewNonNull(gql.Float)
	case 6:
		return gql.Boolean
	case 7:
		return gql.NewNonNull(gql.Boolean)
	}

	return nil
}

func generateFieldNameFromX(x int) string {
	var out string
	alphabet := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "z"}
	v := x
	for {
		r := v % 10
		out = alphabet[r] + out
		v /= 10
		if v == 0 {
			break
		}
	}
	return out
}

func generateWideResolveFromX(x int) func(p gql.Context) (interface{}, error) {
	switch x % 8 {
	case 0:
		return func(ctx gql.Context) (interface{}, error) {
			return fmt.Sprint(x), nil
		}
	case 1:
		return func(ctx gql.Context) (interface{}, error) {
			return fmt.Sprint(x), nil
		}
	case 2:
		return func(ctx gql.Context) (interface{}, error) {
			return x, nil
		}
	case 3:
		return func(ctx gql.Context) (interface{}, error) {
			return x, nil
		}
	case 4:
		return func(ctx gql.Context) (interface{}, error) {
			return float64(x), nil
		}
	case 5:
		return func(ctx gql.Context) (interface{}, error) {
			return float64(x), nil
		}
	case 6:
		return func(ctx gql.Context) (interface{}, error) {
			if x%2 == 0 {
				return false, nil
			}
			return true, nil
		}
	case 7:
		return func(ctx gql.Context) (interface{}, error) {
			if x%2 == 0 {
				return false, nil
			}
			return true, nil
		}
	}

	return nil
}

func WideSchemaQuery(x int) string {
	var fields string
	for i := 0; i < x; i++ {
		fields = fields + generateFieldNameFromX(i) + " "
	}

	return fmt.Sprintf("query { wide { %s} }", fields)
}
