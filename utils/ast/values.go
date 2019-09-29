package ast

/*
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
*/
