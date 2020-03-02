package schema

import (
	"context"
	"sync"
	"time"
)

type gqlCtxKey string

const (
	keyRawQuery      string = "gql_rawQuery"
	keyOperationName string = "gql_operation_name"
	keyVariables     string = "gql_variables"
	keyQuery         string = "gql_query"
	keySchema        string = "gql_schema"
	keyOperation     string = "gql_operation"
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

var _ context.Context = &eCtx{}
