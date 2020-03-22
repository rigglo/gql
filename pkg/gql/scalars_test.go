package gql_test

import (
	"testing"

	"github.com/rigglo/gql/pkg/gql"
)

type scalarResultCoercionCase struct {
	name     string
	res      interface{}
	expected scalarResultCoercion
}

type scalarResultCoercion struct {
	res interface{}
	err error
}

func TestIntScalarResultCoercion(t *testing.T) {
	cs := []scalarResultCoercionCase{
		{
			name: "int32",
			res:  int32(2),
			expected: scalarResultCoercion{
				res: int(2),
				err: nil,
			},
		},
		{
			name: "string",
			res:  "54",
			expected: scalarResultCoercion{
				res: int(54),
				err: nil,
			},
		},
	}
	for _, c := range cs {
		if r, err := gql.Int.CoerceResult(c.res); r != c.expected.res || err != c.expected.err {
			t.Fatalf("%s: Expected result '%v', err '%v'; got result '%v', err '%v'", c.name, c.expected.res, c.expected.err, r, err)
		}
	}

}
