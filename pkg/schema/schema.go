package schema

type Schema interface {
	GetRootQuery() Operation
	GetRootMutation() Operation
	GetRootSubsciption() Operation
	GetDirectives() []Directive
}

type Operation interface {
	GetName() string
	GetFields() []Field
	GetDescription() string
}
