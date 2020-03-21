package gql

import (
	"context"

	"github.com/rigglo/gql/pkg/schema"
)

type Schema struct {
	Query        *Object
	Mutation     *Object
	Subscription *Object
	Directives   Directives
}

type Params schema.ExecuteParams

func (s *Schema) GetRootQuery() schema.Operation {
	return s.Query
}

func (s *Schema) GetRootMutation() schema.Operation {
	return s.Mutation
}

func (s *Schema) GetRootSubsciption() schema.Operation {
	return s.Subscription
}

func (s *Schema) GetDirectives() []schema.Directive {
	return s.Directives
}

func (s *Schema) Exec(ctx context.Context, p Params) *schema.Result {
	return schema.Execute(ctx, s, schema.ExecuteParams(p))
}
