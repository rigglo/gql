package gql

import (
	"context"
	"sync"

	"github.com/rigglo/gql/pkg/language/ast"
)

const (
	keyOperationName string = "gql_operation_name"
	keyVariableDefs  string = "gql_variable_defs"
	keyRawVariables  string = "gql_raw_variables"
	keyVariables     string = "gql_variables"
	keyRawQuery      string = "gql_raw_query"
	keyQuery         string = "gql_query"
	keySchema        string = "gql_schema"
	keyOperation     string = "gql_operation"
	keyTypes         string = "gql_types"
	keyDirectives    string = "gql_directives"
	keyImplementors  string = "gql_implementors"
	keyFragments     string = "gql_fragments"
	keyFragmentUsage string = "gql_fragment_usage"
)

type eCtx struct {
	ctx              context.Context
	store            map[string]interface{}
	mu               sync.Mutex
	res              *Result
	sem              chan struct{}
	enableGoroutines bool
	errMu            sync.Mutex
}

type gqlCtx struct {
	ctx           context.Context
	ectx          *eCtx
	mu            sync.Mutex
	sem           chan struct{}
	concurrency   bool
	res           *Result
	errMu         sync.Mutex
	schema        *Schema
	doc           *ast.Document
	params        *Params
	operation     *ast.Operation
	variables     map[string]interface{}
	types         map[string]Type
	implementors  map[string][]Type
	directives    map[string]Directive
	fragments     map[string]*ast.Fragment
	fragmentUsage map[string]bool
	variableDefs  map[string]map[string]*ast.Variable
	fieldsCache   map[string]map[string]*Field
}

func newContext(ctx context.Context, schema *Schema, doc *ast.Document, params *Params, concurrencyLimit int, concurrency bool) *gqlCtx {
	return &gqlCtx{
		sem:           make(chan struct{}, concurrencyLimit),
		concurrency:   concurrency,
		res:           &Result{},
		schema:        schema,
		doc:           doc,
		params:        params,
		types:         map[string]Type{},
		implementors:  map[string][]Type{},
		directives:    map[string]Directive{},
		fragments:     map[string]*ast.Fragment{},
		fragmentUsage: map[string]bool{},
		variables:     map[string]interface{}{},
		variableDefs:  map[string]map[string]*ast.Variable{},
		fieldsCache:   map[string]map[string]*Field{},
	}
}

func (c *gqlCtx) addErr(err *Error) {
	c.errMu.Lock()
	defer c.errMu.Unlock()
	c.res.Errors = append(c.res.Errors, err)
}

type Context interface {
	Context() context.Context
	Path() []interface{}
	// 	Fields() []string
	Args() map[string]interface{}
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
