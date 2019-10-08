package gql

// InputObject ...
type InputObject struct {
	Name        string
	Description string
	Fields      InputFields
}

// InputFields contains InputObjectFields
type InputFields []*InputObjectField

// Get returns the InputObjectField with the given name
func (i InputFields) Get(name string) *InputObjectField {
	for _, f := range i {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// InputObjectField describes a field of an InputObject
type InputObjectField struct {
	Name        string
	Type        Type
	Description string
}

// Unwrap is defined to implement the Type interface
func (o *InputObject) Unwrap() Type {
	return nil
}

// Kind returns the kind of the type, which is InputObjectTypeDefinition
func (o *InputObject) Kind() TypeDefinition {
	return InputObjectTypeDefinition
}

// Validate runs a check if everything is okay with the Type
func (o *InputObject) Validate(ctx *ValidationContext) error {
	return nil
}
