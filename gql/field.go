package gql

import (
	"context"
	"reflect"

	"github.com/rigglo/gql/schema"
)

type Fields []schema.Field
type Resolver func(context.Context, interface{}, map[string]interface{}) (interface{}, error)

type Field struct {
	Description string
	Name        string
	Arguments   Arguments
	Type        Type
	Directives  Directives
	Resolver    Resolver
}

func (f *Field) GetDescription() string {
	return f.Description
}

func (f *Field) GetName() string {
	return f.Name
}

func (f *Field) GetArguments() []schema.Argument {
	return f.Arguments
}

func (f *Field) GetType() schema.Type {
	return f.Type
}

func (f *Field) GetDirectives() []schema.Directive {
	return f.Directives
}

func (f *Field) Resolve(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
	if f.Resolver == nil {
		f.Resolver = defaultResolver(f.Name)
	}
	return f.Resolver(ctx, parent, args)
}

var _ schema.Field = &Field{}

func defaultResolver(fname string) Resolver {
	return func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
		t := reflect.TypeOf(parent)
		v := reflect.ValueOf(parent)
		for i := 0; i < t.NumField(); i++ {
			// Get the field, returns https://golang.org/pkg/reflect/#StructField
			field := t.Field(i)
			// Get the field tag value
			tag := field.Tag.Get("json")
			if tag == fname {
				return v.FieldByName(field.Name).Interface(), nil
			}
		}
		return nil, nil
	}
}
