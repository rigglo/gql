package gql

import (
	"context"
)

type Union struct {
	Description  string
	Name         string
	Members      Members
	Directives   Directives
	TypeResolver TypeResolver
}

type Members []Type

func (u *Union) GetDescription() string {
	return u.Description
}

func (u *Union) GetName() string {
	return u.Name
}

func (u *Union) GetKind() TypeKind {
	return UnionKind
}

func (u *Union) GetMembers() []Type {
	return u.Members
}

func (u *Union) GetDirectives() []*Directive {
	return u.Directives
}

func (u *Union) Resolve(ctx context.Context, v interface{}) *Object {
	return u.TypeResolver(ctx, v)
}

// var _ UnionType = &Union{}
