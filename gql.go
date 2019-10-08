package gql

import (
	"context"
)

type Params struct {
	Ctx           context.Context
	Schema        Schema
	Query         string
	OperationName string
	Variables     map[string]interface{}
}

func Execute(p *Params) *Result {
	e := &executor{
		schema: &p.Schema,
		types:  map[string]Type{},
	}
	return e.Execute(p.Ctx, p.Query, p.OperationName, p.Variables)
}
