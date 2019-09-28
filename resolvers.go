package gql

import (
	"context"
	"reflect"
)

type ResolverFunc func(context.Context, ResolverArgs, ResolverInfo) (interface{}, error)

var defaultObjectResolver = func(ctx context.Context, args ResolverArgs, info ResolverInfo) (interface{}, error) {
	dv := reflect.ValueOf(info.Parent)
	for i := 0; i < dv.NumField(); i++ {
		if dv.Type().Field(i).Tag.Get(fieldTagName) == info.Field.Name {
			return dv.Field(i).Interface(), nil
		}
	}
	return nil, &Error{
		Path:    info.Path,
		Message: ErrNoResolver,
	}
}
