package gql

import (
	"context"
	"reflect"
)

type ResolverFunc func(context.Context, Arguments, ResolverInfo) (interface{}, error)

var defaultObjectResolver = func(ctx context.Context, args Arguments, info ResolverInfo) (interface{}, error) {
	dv := reflect.ValueOf(info.Parent)
	for i := 0; i < dv.NumField(); i++ {
		if dv.Type().Field(i).Tag.Get("json") == info.Field.Name {
			return dv.Field(i).Interface(), nil
		}
	}
	return nil, nil
}
