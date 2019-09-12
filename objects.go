package gql

type Object struct {
	Name     string
	Fields   Fields
	isArray  bool
	isObject bool
}

func (o *Object) IsArray() bool {
	return o.isArray
}

func (o *Object) IsObject() bool {
	return o.isObject
}

func NewObject() *Object {
	return &Object{}
}

func (o *Object) Type() FieldType {
	return ObjectType
}

func (o *Object) Value() interface{} {
	return o
}
