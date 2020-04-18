package benchutil

import (
	"fmt"

	"github.com/rigglo/gql"
)

type color struct {
	Hex string
	R   int
	G   int
	B   int
}

func ListSchemaWithXItems(x int) *gql.Schema {

	list := generateXListItems(x)

	color := &gql.Object{
		Name:        "Color",
		Description: "A color",
		Fields: gql.Fields{
			"hex": &gql.Field{
				Name:        "hex",
				Type:        gql.NewNonNull(gql.String),
				Description: "Hex color code.",
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if c, ok := ctx.Parent().(color); ok {
						return c.Hex, nil
					}
					return nil, nil
				},
			},
			"r": &gql.Field{
				Name:        "r",
				Type:        gql.NewNonNull(gql.Int),
				Description: "Red value.",
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if c, ok := ctx.Parent().(color); ok {
						return c.R, nil
					}
					return nil, nil
				},
			},
			"g": &gql.Field{
				Name:        "g",
				Type:        gql.NewNonNull(gql.Int),
				Description: "Green value.",
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if c, ok := ctx.Parent().(color); ok {
						return c.G, nil
					}
					return nil, nil
				},
			},
			"b": &gql.Field{
				Name:        "b",
				Type:        gql.NewNonNull(gql.Int),
				Description: "Blue value.",
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if c, ok := ctx.Parent().(color); ok {
						return c.B, nil
					}
					return nil, nil
				},
			},
		},
	}

	queryType := &gql.Object{
		Name: "Query",
		Fields: gql.Fields{
			"colors": {
				Name: "colors",
				Type: gql.NewList(color),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return list, nil
				},
			},
		},
	}

	colorSchema := &gql.Schema{
		Query: queryType,
	}

	return colorSchema
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

func generateXListItems(x int) []color {
	if x > len(colors) {
		x = len(colors)
	}
	return colors[0:x]
}
