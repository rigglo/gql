package gql

import (
	"context"
	"github.com/rigglo/gql/schema"
)

type Params struct {
	Ctx           context.Context
	Schema        schema.Schema
	Query         string
	OperationName string
	Variables     map[string]interface{}
}

func Execute(p *Params) *Result {
	e := &executor{
		schema: &p.Schema,
		types:  map[string]schema.Type{},
	}
	return e.Execute(p.Ctx, p.Query, p.OperationName, p.Variables)
}
