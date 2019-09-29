package ast

type Document struct {
	Operation *Operation
}

func NewDocument() *Document {
	return &Document{}
}

type OperationType int

const (
	Query OperationType = iota
	Mutation
	Subscription
)

type Operation struct {
	OperationType OperationType
	Name          string
	Variables     []*Variable
	Directives    []*Directive
	SelectionSet  []*Selection
}

func NewOperation(ot OperationType) *Operation {
	return &Operation{
		OperationType: ot,
	}
}

type Variable struct {
	Name         string
	Type         Type
	DefaultValue Value
}

type Argument struct {
	Name string
	Value
}

type Directive struct {
	Name      string
	Arguments []*Argument
}

type TypeKind int

const (
	Named TypeKind = iota
	List
	NonNull
)

type Type interface {
	Kind() TypeKind
	GetValue() interface{}
}

type NamedType struct {
	Name string
}

func (t *NamedType) Kind() TypeKind {
	return Named
}

func (t *NamedType) GetValue() interface{} {
	return t
}

type ListType struct {
	Type
}

func (t *ListType) Kind() TypeKind {
	return List
}

func (t *ListType) GetValue() interface{} {
	return t.Type
}

type NonNullType struct {
	Type
}

func (t *NonNullType) Kind() TypeKind {
	return NonNull
}

func (t *NonNullType) GetValue() interface{} {
	return t.Type
}
