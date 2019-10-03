package gql_test

import (
	"context"
	"fmt"
	"github.com/rigglo/gql"
	"testing"
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

func BenchmarkListQuery_1K(b *testing.B) {
	banchN(1000)(b)
}

func BenchmarkListQuery_100K(b *testing.B) {
	banchN(100000)(b)
}

func banchN(n int) func(b *testing.B) {
	return func(b *testing.B) {
		schema := ListWithNItem(n)
		for i := 0; i < b.N; i++ {

			query := `
				query {
					colors {
						hex
						r
						g
						b
					}
				}
			`
			benchGraphql(schema, query, b)
		}
	}
}

func benchGraphql(s *gql.Schema, q string, t testing.TB) {
	result, _ := s.Resolve(context.Background(), q)
	if len(result.Errors) > 0 {
		t.Fatalf("wrong result, unexpected errors: %v", result.Errors)
	}
}

func ListWithNItem(n int) *gql.Schema {
	color := &gql.Object{
		Name:   "Color",
		Fields: gql.Fields{},
	}
	query := &gql.Object{
		Name: "Query",
		Fields: gql.Fields{
			&gql.Field{
				Name: "colors",
				Type: gql.NewList(color),
				Resolver: func(ctx context.Context, args map[string]interface{}, parent interface{}) (interface{}, error) {
					return nil, nil
				},
			},
		},
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
