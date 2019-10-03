package gql

/*
TypeDefinition:
	ScalarTypeDefinition
	ObjectTypeDefinition
	InterfaceTypeDefinition
	UnionTypeDefinition
	EnumTypeDefinition
	InputObjectTypeDefinition
*/

type TypeDefinition int

const (
	NonNullTypeDefinition TypeDefinition = iota
	ListTypeDefinition
	ScalarTypeDefinition
	ObjectTypeDefinition
	InterfaceTypeDefinition
	UnionTypeDefinition
	EnumTypeDefinition
	InputObjectTypeDefinition
)

type Type interface {
	Unwrap() Type
	Kind() TypeDefinition
}

func NewNonNull(t Type) Type {
	return &NonNull{
		Type: t,
	}
}

type NonNull struct {
	Type Type
}

func (nn *NonNull) Unwrap() Type {
	return nn.Type
}

func (nn *NonNull) Kind() TypeDefinition {
	return NonNullTypeDefinition
}

func NewList(t Type) Type {
	return &List{
		Type: t,
	}
}

type List struct {
	Type Type
}

func (l *List) Unwrap() Type {
	return l.Type
}

func (l *List) Kind() TypeDefinition {
	return ListTypeDefinition
}

func isInputType(t Type) bool {
	if t.Kind() == ListTypeDefinition || t.Kind() == NonNullTypeDefinition {
		return isInputType(t.Unwrap())
	}
	if t.Kind() == ScalarTypeDefinition || t.Kind() == EnumTypeDefinition || t.Kind() == InputObjectTypeDefinition {
		return true
	}
	return false
}

func isOutputType(t Type) bool {
	if t.Kind() == ListTypeDefinition || t.Kind() == NonNullTypeDefinition {
		return isOutputType(t.Unwrap())
	}
	if t.Kind() == ScalarTypeDefinition || t.Kind() == ObjectTypeDefinition || t.Kind() == InterfaceTypeDefinition || t.Kind() == UnionTypeDefinition || t.Kind() == EnumTypeDefinition {
		return true
	}
	return false
}
