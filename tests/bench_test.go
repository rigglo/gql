package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/rigglo/gql"
)

type color struct {
	Hex string
	R   int
	G   int
	B   int
}

var colors []color

func init() {
	colors = make([]color, 0, 256*16*16)

	for r := 0; r < 256; r++ {
		for g := 0; g < 16; g++ {
			for b := 0; b < 16; b++ {
				colors = append(colors, color{
					Hex: fmt.Sprintf("#%x%x%x", r, g, b),
					R:   r,
					G:   g,
					B:   b,
				})
			}
		}
	}
}

func BenchmarkListQuery_1(b *testing.B) {
	banchN(1)(b)
}

func BenchmarkListQuery_100(b *testing.B) {
	banchN(100)(b)
}

func BenchmarkListQuery_1K(b *testing.B) {
	banchN(1000)(b)
}

func BenchmarkListQuery_10K(b *testing.B) {
	banchN(10000)(b)
}

func BenchmarkListQuery_100K(b *testing.B) {
	banchN(100000)(b)
}

func banchN(n int) func(b *testing.B) {
	return func(b *testing.B) {
		schema := ListWithNItem(n)
		query := `
		{
			colors {
				hex
				r
				g
				b
			}
		}
	`
		/*
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					benchGraphql(schema, query, b)
				}
			})
		*/
		for i := 0; i < b.N; i++ {
			benchGraphql(schema, query, b)
		}

	}
}

func benchGraphql(s *gql.Schema, q string, t testing.TB) {
	result := s.Execute(&gql.Params{
		Ctx:   context.Background(),
		Query: q,
	})
	if len(result.Errors) > 0 {
		t.Fatalf("wrong result, unexpected errors: %v", result.Errors)
	}
}

func ListWithNItem(n int) *gql.Schema {
	color := &gql.Object{
		Name: "Color",
		Fields: gql.NewFields(
			&gql.Field{
				Name: "hex",
				Type: gql.String,
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return parent.(color).Hex, nil
				},
			},
			&gql.Field{
				Name: "r",
				Type: gql.String,
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return parent.(color).R, nil
				},
			},
			&gql.Field{
				Name: "g",
				Type: gql.String,
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return parent.(color).G, nil
				},
			},
			&gql.Field{
				Name: "b",
				Type: gql.String,
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return parent.(color).B, nil
				},
			},
		),
	}
	query := &gql.Object{
		Name: "Query",
		Fields: gql.NewFields(
			&gql.Field{
				Name: "colors",
				Type: gql.NewList(color),
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return generateXListItems(n), nil
				},
			},
		),
	}
	schema := gql.Schema{
		Query: query,
	}

	return &schema
}

func generateXListItems(x int) []color {
	if x > len(colors) {
		x = len(colors)
	}
	return colors[0:x]
}
