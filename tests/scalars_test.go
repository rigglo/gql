package tests

import (
	"testing"

	"github.com/rigglo/gql/schema"
)

func TestStringScalar_InputCoercion(t *testing.T) {
	s := schema.ID
	o, err := s.InputCoercion(`"3"`)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("'%s'", o)
}

func TestStringScalar_OutputCoercion(t *testing.T) {
	s := schema.ID
	o, err := s.OutputCoercion(42)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", string(o))
}
