package tests

import (
	"context"
	"fmt"
	"testing"

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
		Fields: gql.NewFields(
			&gql.Field{
				Name: "wide",
				Type: gql.NewList(wide),
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					out := make([]struct{}, 0, y)
					for i := 0; i < y; i++ {
						out = append(out, struct{}{})
					}
					return out, nil
				},
			},
		),
	}

	wideSchema := &gql.Schema{
		Query: queryType,
	}

	return wideSchema
}

func generateXWideFields(x int) *gql.Fields {
	fields := gql.NewFields()
	for i := 0; i < x; i++ {
		fields.Add(generateWideFieldFromX(i))
	}
	return fields
}

func generateWideFieldFromX(x int) *gql.Field {
	return &gql.Field{
		Name:     generateFieldNameFromX(x),
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

func generateWideResolveFromX(x int) gql.ResolverFunc {
	switch x % 8 {
	case 0:
		return func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
			return fmt.Sprint(x), nil
		}
	case 1:
		return func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
			return fmt.Sprint(x), nil
		}
	case 2:
		return func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
			return x, nil
		}
	case 3:
		return func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
			return x, nil
		}
	case 4:
		return func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
			return float64(x), nil
		}
	case 5:
		return func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
			return float64(x), nil
		}
	case 6:
		return func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
			if x%2 == 0 {
				return false, nil
			}
			return true, nil
		}
	case 7:
		return func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
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

func nFieldsyItemsQueryBenchmark(x int, y int) func(b *testing.B) {
	return func(b *testing.B) {
		schema := WideSchemaWithXFieldsAndYItems(x, y)
		query := WideSchemaQuery(x)

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				benchGraphql(schema, query, b)
			}
		})
	}
}

func BenchmarkWideQuery_1_1(b *testing.B) {
	nFieldsyItemsQueryBenchmark(1, 1)(b)
}

func BenchmarkWideQuery_10_1(b *testing.B) {
	nFieldsyItemsQueryBenchmark(10, 1)(b)
}

func BenchmarkWideQuery_100_1(b *testing.B) {
	nFieldsyItemsQueryBenchmark(100, 1)(b)
}

func BenchmarkWideQuery_1K_1(b *testing.B) {
	nFieldsyItemsQueryBenchmark(1000, 1)(b)
}

func BenchmarkWideQuery_1_10(b *testing.B) {
	nFieldsyItemsQueryBenchmark(1, 10)(b)
}

func BenchmarkWideQuery_10_10(b *testing.B) {
	nFieldsyItemsQueryBenchmark(10, 10)(b)
}

func BenchmarkWideQuery_100_10(b *testing.B) {
	nFieldsyItemsQueryBenchmark(100, 10)(b)
}

func BenchmarkWideQuery_1K_10(b *testing.B) {
	nFieldsyItemsQueryBenchmark(1000, 10)(b)
}
