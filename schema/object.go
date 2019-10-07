package schema

// Object ...
type Object struct {
	Name        string
	Description string
	Implements  []Type
	Fields      Fields
}

// Kind returns ObjectTypeDefinition
func (o *Object) Kind() TypeDefinition {
	return ObjectTypeDefinition
}

// Unwrap returns nothing for an Object
func (o *Object) Unwrap() Type {
	return nil
}
