package gql

type Type interface {
	Type() FieldType
	Value() interface{}
}

type FieldType int

const (
	ArrayType FieldType = iota
	ObjectType
	ScalarType
)
