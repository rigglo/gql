package schema

type Schema interface {
	GetRootQuery() ObjectType
	GetRootMutation() ObjectType
	GetRootSubsciption() ObjectType
	GetDirectives() []Directive
}

func getTypes(s Schema) map[string]Type {
	return nil
}

type Operation interface {
	GetName() string
	GetFields() []Field
	GetDescription() string
}
