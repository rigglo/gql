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
