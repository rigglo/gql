package gql

type Schema struct {
	Query        *Object
	Mutation     *Object
	Subscription *Object
	Directives   Directives
}

func (s *Schema) GetRootQuery() *Object {
	return s.Query
}

func (s *Schema) GetRootMutation() *Object {
	return s.Mutation
}

func (s *Schema) GetRootSubsciption() *Object {
	return s.Subscription
}

func (s *Schema) GetDirectives() []Directive {
	return s.Directives
}
