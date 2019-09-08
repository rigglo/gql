package gql

import "context"

type ResolverFunc func(context.Context, Arguments, ResolverInfo) (interface{}, error)
