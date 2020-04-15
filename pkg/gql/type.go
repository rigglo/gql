package gql

// TypeKind shows the kind of a Type
type TypeKind uint

const (
	ScalarKind TypeKind = iota
	ObjectKind
	InterfaceKind
	UnionKind
	EnumKind
	InputObjectKind
	NonNullKind
	ListKind
)

// Type represents a Type in the GraphQL Type System
type Type interface {
	GetName() string
	GetDescription() string
	GetKind() TypeKind
}

func isInputType(t Type) bool {
	if t.GetKind() == ListKind || t.GetKind() == NonNullKind {
		return isInputType(t.(WrappingType).Unwrap())
	}
	if t.GetKind() == ScalarKind || t.GetKind() == EnumKind || t.GetKind() == InputObjectKind {
		return true
	}
	return false
}

func isOutputType(t Type) bool {
	if t.GetKind() == ListKind || t.GetKind() == NonNullKind {
		return isOutputType(t.(WrappingType).Unwrap())
	}
	if t.GetKind() == ScalarKind || t.GetKind() == ObjectKind || t.GetKind() == InterfaceKind || t.GetKind() == UnionKind || t.GetKind() == EnumKind {
		return true
	}
	return false
}

// List in the GraphQL Type System
type List struct {
	Name        string
	Description string
	Wrapped     Type
}

// NewList returns a new List
// In every case, you should use NewList function to create a new List, instead of creating on your own.
func NewList(t Type) Type {
	return &List{
		Name:        "List",
		Description: "Built-in 'List' type",
		Wrapped:     t,
	}
}

// GetName returns the name of the type, "List"
func (l *List) GetName() string {
	return l.Name
}

// GetDescription shows the description of the type
func (l *List) GetDescription() string {
	return l.Description
}

// GetKind returns the kind of the type, ListKind
func (l *List) GetKind() TypeKind {
	return ListKind
}

// Unwrap unwraps the inner type from the List
func (l *List) Unwrap() Type {
	return l.Wrapped
}

// NonNull type from the GraphQL type system
type NonNull struct {
	Name        string
	Description string
	Wrapped     Type
}

// NewNonNull function helps create a new Non-Null type
// It must be used, instead of creating a non null type "manually"
func NewNonNull(t Type) Type {
	return &NonNull{
		Name:        "NonNull",
		Description: "Built-in 'NonNull' type",
		Wrapped:     t,
	}
}

// GetName returns the name of the type, "NonNull"
func (l *NonNull) GetName() string {
	return l.Name
}

// GetDescription returns the description of the NonNull type
func (l *NonNull) GetDescription() string {
	return l.Description
}

// GetKind returns the kind of the type, NonNullKind
func (l *NonNull) GetKind() TypeKind {
	return NonNullKind
}

// Unwrap unwraps the inner type of the NonNull type
func (l *NonNull) Unwrap() Type {
	return l.Wrapped
}
