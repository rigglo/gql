package ast

import "fmt"

type Location struct {
	Column int
	Line   int
}

type Fragment struct {
	Name          string
	TypeCondition string
	Directives    []*Directive
	SelectionSet  []Selection
	Location      Location
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
	Location      Location
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
	Location     Location
}

type Argument struct {
	Name     string
	Value    Value
	Location Location
}

func argumentToString(a *Argument) string {
	return fmt.Sprintf("%s: %s", a.Name, a.Value.String())
}

type Directive struct {
	Name      string
	Arguments []*Argument
	Location  Location
}

func (d *Directive) String() string {
	out := "@" + d.Name
	if d.Arguments != nil && len(d.Arguments) != 0 {
		out += "("
		for i, a := range d.Arguments {
			if i != 0 {
				out += ", "
			}
			out += argumentToString(a)
		}
		out += ")"
	}
	return out
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
	String() string
}

type NamedType struct {
	Name     string
	Location Location
}

func (t *NamedType) Kind() TypeKind {
	return Named
}

func (t *NamedType) GetValue() interface{} {
	return t.Name
}

func (t *NamedType) String() string {
	return t.Name
}

type ListType struct {
	Type
	Location Location
}

func (t *ListType) Kind() TypeKind {
	return List
}

func (t *ListType) GetValue() interface{} {
	return t.Type
}

func (t *ListType) String() string {
	return "[" + t.Type.String() + "]"
}

type NonNullType struct {
	Type
	Location Location
}

func (t *NonNullType) Kind() TypeKind {
	return NonNull
}

func (t *NonNullType) GetValue() interface{} {
	return t.Type
}

func (t *NonNullType) String() string {
	return t.Type.String() + "!"
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
	ParentType   string
	Arguments    []*Argument
	Directives   []*Directive
	SelectionSet []Selection
	Location     Location
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
	Location   Location
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
	Location      Location
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
	GetLocation() Location
	String() string
}

type VariableValue struct {
	Name     string
	Location Location
}

func (v *VariableValue) GetValue() interface{} {
	return v.Name
}

func (v *VariableValue) Kind() ValueKind {
	return VariableValueKind
}

func (v *VariableValue) GetLocation() Location {
	return v.Location
}

func (v *VariableValue) String() string {
	return "$" + v.Name
}

type IntValue struct {
	Value    string
	Location Location
}

func (v *IntValue) GetValue() interface{} {
	return v.Value
}

func (v *IntValue) Kind() ValueKind {
	return IntValueKind
}

func (v *IntValue) GetLocation() Location {
	return v.Location
}

func (v *IntValue) String() string {
	return v.Value
}

type FloatValue struct {
	Value    string
	Location Location
}

func (v *FloatValue) GetValue() interface{} {
	return v.Value
}

func (v *FloatValue) Kind() ValueKind {
	return FloatValueKind
}

func (v *FloatValue) GetLocation() Location {
	return v.Location
}

func (v *FloatValue) String() string {
	return v.Value
}

type StringValue struct {
	Value    string
	Location Location
}

func (v *StringValue) GetValue() interface{} {
	return v.Value
}

func (v *StringValue) Kind() ValueKind {
	return StringValueKind
}

func (v *StringValue) GetLocation() Location {
	return v.Location
}

func (v *StringValue) String() string {
	return `"` + v.Value + `"`
}

type BooleanValue struct {
	Value    string
	Location Location
}

func (v *BooleanValue) GetValue() interface{} {
	return v.Value
}

func (v *BooleanValue) Kind() ValueKind {
	return BooleanValueKind
}

func (v *BooleanValue) GetLocation() Location {
	return v.Location
}

func (v *BooleanValue) String() string {
	return v.Value
}

type NullValue struct {
	Value    string
	Location Location
}

func (v *NullValue) GetValue() interface{} {
	return v.Value
}

func (v *NullValue) Kind() ValueKind {
	return NullValueKind
}

func (v *NullValue) GetLocation() Location {
	return v.Location
}

func (v *NullValue) String() string {
	return "null"
}

type EnumValue struct {
	Value    string
	Location Location
}

func (v *EnumValue) GetValue() interface{} {
	return v.Value
}

func (v *EnumValue) Kind() ValueKind {
	return EnumValueKind
}

func (v *EnumValue) GetLocation() Location {
	return v.Location
}

func (v *EnumValue) String() string {
	return v.Value
}

type ListValue struct {
	Values   []Value
	Location Location
}

func (v *ListValue) GetValue() interface{} {
	return v.Values
}

func (v *ListValue) Kind() ValueKind {
	return ListValueKind
}

func (v *ListValue) GetLocation() Location {
	return v.Location
}

func (v *ListValue) String() string {
	out := ""
	for i, vi := range v.Values {
		out += " " + vi.String()
		if i != len(v.Values)-1 {
			out += ","
		}
	}
	return "[" + out + "]"
}

type ObjectValue struct {
	Fields   []*ObjectFieldValue
	Location Location
}

func (v *ObjectValue) GetValue() interface{} {
	return v.Fields
}

func (v *ObjectValue) Kind() ValueKind {
	return ObjectValueKind
}

func (v *ObjectValue) GetLocation() Location {
	return v.Location
}

func (v *ObjectValue) String() string {
	out := "{ "
	for _, f := range v.Fields {
		out += f.String() + " "
	}
	return out + "}"
}

type ObjectFieldValue struct {
	Name     string
	Value    Value
	Location Location
}

func (v *ObjectFieldValue) GetValue() interface{} {
	return v.Value
}

func (v *ObjectFieldValue) Kind() ValueKind {
	return ObjectFieldValueKind
}

func (v *ObjectFieldValue) GetLocation() Location {
	return v.Location
}

func (v *ObjectFieldValue) String() string {
	return v.Name + ": " + v.Value.String()
}
