package gql

type Object struct {
	Description string
	Name        string
	Implements  Interfaces
	Directives  Directives
	Fields      Fields
}

func (o *Object) GetDescription() string {
	return o.Description
}

func (o *Object) GetName() string {
	return o.Name
}

func (o *Object) GetKind() TypeKind {
	return ObjectKind
}

func (o *Object) GetInterfaces() []*Interface {
	return o.Implements
}

func (o *Object) GetDirectives() []*Directive {
	return o.Directives
}

func (o *Object) GetFields() []*Field {
	return o.Fields
}
