package gql

type InputObject struct {
	Name        string
	Description string
	Fields      InputFields
}

type InputFields []*InputObjectField

func (i InputFields) Get(name string) *InputObjectField {
	for _, f := range i {
		if f.Name == name {
			return f
		}
	}
	return nil
}

type InputObjectField struct {
	Name        string
	Type        Type
	Description string
}

func (o *InputObject) Unwrap() Type {
	return nil
}

func (o *InputObject) Kind() TypeDefinition {
	return InputObjectTypeDefinition
}

func (o *InputObject) Validate() error {
	return nil
}
