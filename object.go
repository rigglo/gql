package gql

type Object struct {
	Name        string
	Description string
	Implements  []Type
	Fields      Fields
}

func (o *Object) Kind() TypeDefinition {
	return ObjectTypeDefinition
}

func (o *Object) Unwrap() Type {
	return nil
}
