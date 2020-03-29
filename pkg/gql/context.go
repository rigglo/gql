package gql

import (
	"context"
	"sync"
	"time"
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
)

type eCtx struct {
	ctx   context.Context
	store map[string]interface{}
	mu    sync.Mutex
	res   *Result
}

func newCtx(ctx context.Context, store map[string]interface{}) *eCtx {
	return &eCtx{
		ctx:   ctx,
		store: store,
		res: &Result{
			ctx: ctx,
		},
	}
}

func (c *eCtx) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

func (c *eCtx) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func (c *eCtx) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *eCtx) Err() error {
	return c.ctx.Err()
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

var _ context.Context = (*eCtx)(nil)

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
