package gql

import (
	"context"
	"github.com/rigglo/gql/language/parser"
)

// Schema ...
type Schema struct {
	Mutation    *Object
	Query       *Object
	Subsciption *Object
}

// Execute an operation on the schema
func (s *Schema) Execute(p *Params) *Result {
	_, doc, err := parser.Parse(p.Query)
	if err != nil {
		return &Result{
			Errors: []error{err},
		}
	}

	ctx := &gqlCtx{
		Context:       p.Ctx,
		doc:           doc,
		operationName: p.OperationName,
		query:         p.Query,
		vars:          p.Variables,
	}

	v := newValidator(s)

	e, errs := v.Validate(ctx)
	if len(errs) > 0 {
		return &Result{
			Errors: errs,
		}
	}

	return e.Execute(ctx)
}

// Params ...
type Params struct {
	Ctx           context.Context        `json:"-"`
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

// ValidationContext ...
type ValidationContext struct {
	types map[string]Type
}
