package main

import (
	"context"
	"log"

	"github.com/rigglo/gql"
)

type simpleTracer struct {
}

func (e *simpleTracer) Call(ctx context.Context, ev gql.ExtensionEvent, arg interface{}) {
	switch ev {
	case gql.EventParseStart:
		log.Println("simpleTracer - EventParseStart!")
	case gql.EventParseFinish:
		log.Println("simpleTracer - EventParseFinish!")
	case gql.EventValidationStart:
		log.Println("simpleTracer - EventValidationStart!")
	case gql.EventValidationFinish:
		log.Println("simpleTracer - EventValidationFinish!")
	case gql.EventExecutionStart:
		log.Println("simpleTracer - EventExecutionStart!")
	case gql.EventExecutionFinish:
		log.Println("simpleTracer - EventExecutionFinish!")
	case gql.EventFieldResolverStart:
		log.Println("simpleTracer - EventFieldResolverStart!")
	case gql.EventFieldResolverFinish:
		log.Println("simpleTracer - EventFieldResolverFinish!")
	default:
		log.Println("someEvent", ev)
	}
}

func (e *simpleTracer) GetName() string {
	return "simpleTracer"
}

func (e *simpleTracer) Init(ctx context.Context, _ gql.Params) {

}

func (e *simpleTracer) Result(ctx context.Context) interface{} {
	return nil
}

var _ gql.Extension = &simpleTracer{}
