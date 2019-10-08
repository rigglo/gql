package gql

import (
	"context"
)

// Params ...
type Params struct {
	Ctx           context.Context        `json:"-"`
	Schema        Schema                 `json:"-"`
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

// Execute ...
func Execute(p *Params) *Result {
	e := &executor{
		schema: &p.Schema,
		types:  map[string]Type{},
	}
	return e.Execute(p.Ctx, p.Query, p.OperationName, p.Variables)
}
