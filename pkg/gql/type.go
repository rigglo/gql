package gql

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

type Type interface {
	GetName() string
	GetDescription() string
	GetKind() TypeKind
}

type List struct {
	Name        string
	Description string
	Wrapped     Type
}

func NewList(t Type) Type {
	return &List{
		Name:        "List",
		Description: "Built-in 'List' type",
		Wrapped:     t,
	}
}

func (l *List) GetName() string {
	return l.Name
}

func (l *List) GetDescription() string {
	return l.Description
}

func (l *List) GetKind() TypeKind {
	return ListKind
}

func (l *List) Unwrap() Type {
	return l.Wrapped
}

// var _ schema.List = &List{}

type NonNull struct {
	Name        string
	Description string
	Wrapped     Type
}

func NewNonNull(t Type) Type {
	return &NonNull{
		Name:        "NonNull",
		Description: "Built-in 'NonNull' type",
		Wrapped:     t,
	}
}

func (l *NonNull) GetName() string {
	return l.Name
}

func (l *NonNull) GetDescription() string {
	return l.Description
}

func (l *NonNull) GetKind() TypeKind {
	return NonNullKind
}

func (l *NonNull) Unwrap() Type {
	return l.Wrapped
}

// var _ schema.NonNull = &NonNull{}

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
