package gql

import (
	"context"
	"sync"
)

const (
	keyOperationName string = "gql_operation_name"
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

func newCtx(ctx context.Context, store map[string]interface{}, semLimit int, enableGoroutines bool) *eCtx {
	return &eCtx{
		ctx:              ctx,
		store:            store,
		sem:              make(chan struct{}, semLimit),
		enableGoroutines: enableGoroutines,
		res:              &Result{},
	}
}

func (c *eCtx) Get(key string) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.store[key]; ok {
		return v
	}
	return nil
}

func (c *eCtx) Set(key string, v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = v
}

func (c *eCtx) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

func (c *eCtx) addErr(err *Error) {
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
	eCtx   *eCtx
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
