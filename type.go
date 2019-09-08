package gql

type Type interface {
	FieldType() FieldType
	Value() interface{}
}

type FieldType int

const (
	ArrayType FieldType = iota
	ObjectType
	ScalarType
)

var (
	StringType = &Scalar{
		Name: "String",
	}
)

type Scalar struct {
	Name          string
	MarshalText   func(interface{}) []byte
	UnmarshalText func([]byte) (interface{}, error)
}

func (s *Scalar) FieldType() FieldType {
	return ScalarType
}

func (s *Scalar) Value() interface{} {
	return s
}
