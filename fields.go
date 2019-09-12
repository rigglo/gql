package gql

type Field struct {
	Type     Type
	Name     string
	Fields   Fields
	Args     Arguments
	Resolver ResolverFunc
}

func (f *Field) getFields() Fields {
	return f.Fields
}

type Fields map[string]*Field
