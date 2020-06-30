package federation

import (
	"context"

	"github.com/rigglo/gql"
)

func Key(fieldSet string) gql.Directive {
	return &key{fieldSet: fieldSet}
}

type key struct {
	fieldSet string
}

func (d *key) GetName() string {
	return "key"
}

func (d *key) GetDescription() string {
	return ""
}

func (d *key) GetArguments() gql.Arguments {
	return gql.Arguments{
		"fields": &gql.Argument{
			Type: gql.NewNonNull(fieldset),
		},
	}
}

func (d *key) GetLocations() []gql.DirectiveLocation {
	return []gql.DirectiveLocation{gql.ObjectLoc, gql.InterfaceLoc}
}

func (d *key) Variables() map[string]interface{} {
	return map[string]interface{}{"fields": d.fieldSet}
}

func (d *key) VisitObject(ctx context.Context, o gql.Object) *gql.Object {
	return &o
}

func (d *key) VisitInterface(ctx context.Context, i gql.Interface) *gql.Interface {
	return &i
}
