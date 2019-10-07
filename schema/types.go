package schema

/*
TypeDefinition:
	ScalarTypeDefinition
	ObjectTypeDefinition
	InterfaceTypeDefinition
	UnionTypeDefinition
	EnumTypeDefinition
	InputObjectTypeDefinition
*/

// TypeDefinition ...
type TypeDefinition int

const (
	// NonNullTypeDefinition NonNull type
	NonNullTypeDefinition TypeDefinition = iota
	// ListTypeDefinition List type
	ListTypeDefinition
	// ScalarTypeDefinition scalar type
	ScalarTypeDefinition
	// ObjectTypeDefinition object type
	ObjectTypeDefinition
	// InterfaceTypeDefinition interface type
	InterfaceTypeDefinition
	// UnionTypeDefinition union type
	UnionTypeDefinition
	// EnumTypeDefinition enum type
	EnumTypeDefinition
	// InputObjectTypeDefinition input object type
	InputObjectTypeDefinition
)

// Type ...
type Type interface {
	Unwrap() Type
	Kind() TypeDefinition
}

// NewNonNull ...
func NewNonNull(t Type) Type {
	return &NonNull{
		Type: t,
	}
}

// NonNull ...
type NonNull struct {
	Type Type
}

// Unwrap ...
func (nn *NonNull) Unwrap() Type {
	return nn.Type
}

// Kind ...
func (nn *NonNull) Kind() TypeDefinition {
	return NonNullTypeDefinition
}

// NewList ..
func NewList(t Type) Type {
	return &List{
		Type: t,
	}
}

// List ...
type List struct {
	Type Type
}

// Unwrap ...
func (l *List) Unwrap() Type {
	return l.Type
}

// Kind ...
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

func getFinalType(t Type) Type {
	return t
}
