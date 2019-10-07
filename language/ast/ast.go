package ast

type Document struct {
	Operations []*Operation
	Fragments  map[string]*Fragment
}

func NewDocument() *Document {
	return &Document{
		Operations: []*Operation{},
		Fragments:  map[string]*Fragment{},
	}
}

type Fragment struct {
	Name          string
	TypeCondition string
	Directives    []*Directive
	SelectionSet  []Selection
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
	SelectionSet  []Selection
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
	return t.Name
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

// SELECTIONSET

type SelectionKind int

const (
	FieldSelectionKind SelectionKind = iota
	FragmentSpreadSelectionKind
	InlineFragmentSelectionKind
)

type Selection interface {
	Kind() SelectionKind
	GetDirectives() []*Directive
}

type Fields []*Field

type Field struct {
	Alias        string
	Name         string
	Arguments    []*Argument
	Directives   []*Directive
	SelectionSet []Selection
}

func (f *Field) Kind() SelectionKind {
	return FieldSelectionKind
}

func (f *Field) GetDirectives() []*Directive {
	return f.Directives
}

type FragmentSpread struct {
	Name       string
	Directives []*Directive
}

func (fs *FragmentSpread) Kind() SelectionKind {
	return FragmentSpreadSelectionKind
}

func (fs *FragmentSpread) GetDirectives() []*Directive {
	return fs.Directives
}

type InlineFragment struct {
	TypeCondition string
	Directives    []*Directive
	SelectionSet  []Selection
}

func (inf *InlineFragment) Kind() SelectionKind {
	return InlineFragmentSelectionKind
}

func (inf *InlineFragment) GetDirectives() []*Directive {
	return inf.Directives
}

// VALUES

type ValueKind int

const (
	VariableValueKind ValueKind = iota
	IntValueKind
	FloatValueKind
	BooleanValueKind
	StringValueKind
	NullValueKind
	EnumValueKind
	ListValueKind
	ObjectValueKind
	ObjectFieldValueKind
)

type Value interface {
	GetValue() interface{}
	Kind() ValueKind
}

type VariableValue struct {
	Name string
}

func (v *VariableValue) GetValue() interface{} {
	return v.Name
}

func (v *VariableValue) Kind() ValueKind {
	return VariableValueKind
}

type IntValue struct {
	Value string
}

func (v *IntValue) GetValue() interface{} {
	return v.Value
}

func (v *IntValue) Kind() ValueKind {
	return IntValueKind
}

type FloatValue struct {
	Value string
}

func (v *FloatValue) GetValue() interface{} {
	return v.Value
}

func (v *FloatValue) Kind() ValueKind {
	return FloatValueKind
}

type StringValue struct {
	Value string
}

func (v *StringValue) GetValue() interface{} {
	return v.Value
}

func (v *StringValue) Kind() ValueKind {
	return StringValueKind
}

type BooleanValue struct {
	Value string
}

func (v *BooleanValue) GetValue() interface{} {
	return v.Value
}

func (v *BooleanValue) Kind() ValueKind {
	return BooleanValueKind
}

type NullValue struct {
	Value string
}

func (v *NullValue) GetValue() interface{} {
	return v.Value
}

func (v *NullValue) Kind() ValueKind {
	return NullValueKind
}

type EnumValue struct {
	Value string
}

func (v *EnumValue) GetValue() interface{} {
	return v.Value
}

func (v *EnumValue) Kind() ValueKind {
	return EnumValueKind
}

type ListValue struct {
	Values []Value
}

func (v *ListValue) GetValue() interface{} {
	return v.Values
}

func (v *ListValue) Kind() ValueKind {
	return ListValueKind
}

type ObjectValue struct {
	Fields []*ObjectFieldValue
}

func (v *ObjectValue) GetValue() interface{} {
	return v.Fields
}

func (v *ObjectValue) Kind() ValueKind {
	return ObjectValueKind
}

type ObjectFieldValue struct {
	Name  string
	Value Value
}

func (v *ObjectFieldValue) GetValue() interface{} {
	return v.Value
}

func (v *ObjectFieldValue) Kind() ValueKind {
	return ObjectFieldValueKind
}
