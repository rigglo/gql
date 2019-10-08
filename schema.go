package gql

import (
	"context"
)

// Schema ...
type Schema struct {
	Mutation    *Object
	Query       *Object
	Subsciption *Object
}

// Execute an operation on the schema
func (s *Schema) Execute(p *Params) *Result {
	e := &executor{
		schema: s,
		types:  map[string]Type{},
	}
	return e.Execute(p.Ctx, p.Query, p.OperationName, p.Variables)
}

// Params ...
type Params struct {
	Ctx           context.Context        `json:"-"`
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}
