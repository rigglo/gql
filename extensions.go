package gql

import "context"

type Extension interface {
	Init(ctx context.Context, p Params)
	GetName() string
	Call(ctx context.Context, e ExtensionEvent, args interface{})
	Result(ctx context.Context) interface{}
}

type ExtensionEvent uint

const (
	EventParseStart ExtensionEvent = iota
	EventParseFinish
	EventValidationStart
	EventValidationFinish
	EventExecutionStart
	EventExecutionFinish
	EventFieldResolverStart
	EventFieldResolverFinish
)

func callExtensions(ctx context.Context, exts []Extension, ev ExtensionEvent, arg interface{}) {
	for _, e := range exts {
		e.Call(ctx, ev, arg)
	}
}
