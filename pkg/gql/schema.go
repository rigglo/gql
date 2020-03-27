package gql

import (
	"context"
)

type Schema struct {
	Query        *Object
	Mutation     *Object
	Subscription *Object
	Directives   Directives
}

type Params ExecuteParams

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

func (s *Schema) Exec(ctx context.Context, p Params) *Result {
	if s.Query == nil {
		s.Query = &Object{
			Name: "Query",
		}
	}
	if s.Mutation == nil {
		s.Mutation = &Object{
			Name: "Mutation",
		}
	}
	return Execute(ctx, s, ExecuteParams(p))
}
