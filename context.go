package gql

import (
	"context"
	"sync"

	"github.com/rigglo/gql/pkg/language/ast"
)

type gqlCtx struct {
	ctx              context.Context
	mu               sync.Mutex
	sem              chan struct{}
	concurrency      bool
	concurrencyLimit int
	res              *Result
	errMu            sync.Mutex
	schema           *Schema
	doc              *ast.Document
	params           *Params
	operation        *ast.Operation
	variables        map[string]interface{}
	types            map[string]Type
	implementors     map[string][]Type
	directives       map[string]Directive
	fragments        map[string]*ast.Fragment
	fragmentUsage    map[string]bool
	variableDefs     map[string]map[string]*ast.Variable
	variableUsages   map[string]map[string]struct{}
	extensions       []Extension
}

func newContext(ctx context.Context, schema *Schema, doc *ast.Document, params *Params, concurrencyLimit int, concurrency bool) *gqlCtx {
	return &gqlCtx{
		sem:              make(chan struct{}, concurrencyLimit),
		concurrency:      concurrency,
		concurrencyLimit: concurrencyLimit,
		res:              &Result{},
		schema:           schema,
		doc:              doc,
		params:           params,
		types:            map[string]Type{},
		implementors:     map[string][]Type{},
		directives:       map[string]Directive{},
		fragments:        map[string]*ast.Fragment{},
		fragmentUsage:    map[string]bool{},
		variables:        map[string]interface{}{},
		variableDefs:     map[string]map[string]*ast.Variable{},
		variableUsages:   map[string]map[string]struct{}{},
	}
}

func (c *gqlCtx) addErr(err *Error) {
	c.errMu.Lock()
	defer c.errMu.Unlock()
	c.res.Errors = append(c.res.Errors, err)
}

/*
Context for the field resolver functions
*/
type Context interface {
	// Context returns the original context that was given to the executor
	Context() context.Context
	// Path of the field
	Path() []interface{}
	// Args of the field
	Args() map[string]interface{}
	// Parent object's data
	Parent() interface{}
}

type resolveContext struct {
	ctx    context.Context
	gqlCtx *gqlCtx
	path   []interface{}
	fields []string
	args   map[string]interface{}
	parent interface{}
}

func (r *resolveContext) Context() context.Context {
	return r.ctx
}

func (r *resolveContext) Path() []interface{} {
	return r.path
}

func (r *resolveContext) Fields() []string {
	return r.fields
}

func (r *resolveContext) Args() map[string]interface{} {
	return r.args
}

func (r *resolveContext) Parent() interface{} {
	return r.parent
}
