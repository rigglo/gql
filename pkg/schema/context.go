package schema

import (
	"context"
	"sync"
	"time"
)

type gqlCtxKey string

type GqlContext interface {
	context.Context
	Get(key string) interface{}
	Set(key string, v interface{})
	Delete(key string)
}

const (
	keyOperationName string = "gql_operation_name"
	keyRawVariables  string = "gql_raw_variables"
	keyVariables     string = "gql_variables"
	keyRawQuery      string = "gql_raw_query"
	keyQuery         string = "gql_query"
	keySchema        string = "gql_schema"
	keyOperation     string = "gql_operation"
	keyTypes         string = "gql_types"
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
	return c.Value(key)
}

func (c *eCtx) Done() <-chan struct{} {
	return c.Done()
}

func (c *eCtx) Err() error {
	return c.Err()
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

var _ context.Context = &eCtx{}
