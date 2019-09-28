package gql

type Type interface {
	Type() FieldType
	Value() interface{}
	IsNullable() bool
}

type FieldType int

const (
	ArrayType FieldType = iota
	ObjectType
	ScalarType
)

func NewNonNull(t Type) Type {
	return &nonnull{
		T: t,
	}
}

type nonnull struct {
	T Type
}

func (n *nonnull) Type() FieldType {
	return n.T.Type()
}

func (n *nonnull) Value() interface{} {
	return n.T.Value()
}

func (n *nonnull) IsNullable() bool {
	return false
}
