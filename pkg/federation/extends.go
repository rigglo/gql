package federation

import (
	"context"

	"github.com/rigglo/gql"
)

func Extend() gql.Directive {
	return &extends{}
}

type extends struct{}

func (d *extends) GetName() string {
	return "extends"
}

func (d *extends) GetDescription() string {
	return ""
}

func (d *extends) GetArguments() gql.Arguments {
	return gql.Arguments{}
}

func (d *extends) GetLocations() []gql.DirectiveLocation {
	return []gql.DirectiveLocation{gql.ObjectLoc, gql.InterfaceLoc}
}

func (d *extends) Variables() map[string]interface{} {
	return map[string]interface{}{}
}

func (d *extends) VisitObject(ctx context.Context, o gql.Object) *gql.Object {
	return &o
}

func (d *extends) VisitInterface(ctx context.Context, i gql.Interface) *gql.Interface {
	return &i
}
