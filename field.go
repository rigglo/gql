package gql

type Fields map[string]*Field

func (fs Fields) Add(name string, f *Field) {
	fs[name] = f
}

type Resolver func(Context) (interface{}, error)

type Field struct {
	Description string
	Arguments   Arguments
	Type        Type
	Directives  Directives
	Resolver    Resolver
}

func (f *Field) GetDescription() string {
	return f.Description
}

func (f *Field) GetArguments() []*Argument {
	return f.Arguments
}

func (f *Field) GetType() Type {
	return f.Type
}

func (f *Field) GetDirectives() []Directive {
	return f.Directives
}

func (f *Field) IsDeprecated() bool {
	for _, d := range f.Directives {
		if _, ok := d.(*deprecated); ok {
			return true
		}
	}
	return false
}
