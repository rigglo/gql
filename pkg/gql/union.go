package gql

import (
	"context"

	"github.com/rigglo/gql/pkg/schema"
)

type Union struct {
	Description  string
	Name         string
	Members      Members
	Directives   Directives
	TypeResolver TypeResolver
}

type Members []schema.Type

func (u *Union) GetDescription() string {
	return u.Description
}

func (u *Union) GetName() string {
	return u.Name
}

func (u *Union) GetKind() schema.TypeKind {
	return schema.UnionKind
}

func (u *Union) GetMembers() []schema.Type {
	return u.Members
}

func (u *Union) GetDirectives() []schema.Directive {
	return u.Directives
}

func (u *Union) Resolve(ctx context.Context, v interface{}) schema.ObjectType {
	return u.TypeResolver(ctx, v)
}

var _ schema.UnionType = &Union{}
