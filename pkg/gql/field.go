package gql

import (
	"reflect"
	"strings"
)

type Fields []*Field

func (fs Fields) Add(f *Field) {
	fs = append(fs, f)
}

type Resolver func(Context) (interface{}, error)

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

func (f *Field) GetArguments() []*Argument {
	return f.Arguments
}

func (f *Field) GetType() Type {
	return f.Type
}

func (f *Field) GetDirectives() []Directive {
	return f.Directives
}

func (f *Field) Resolve(ctx Context) (interface{}, error) {
	if f.Resolver == nil {
		f.Resolver = defaultResolver(f.Name)
	}
	return f.Resolver(ctx)
}

func (f *Field) IsDeprecated() bool {
	for _, d := range f.Directives {
		if _, ok := d.(*deprecated); ok {
			return true
		}
	}
	return false
}

// var _ schema.Field = &Field{}

func defaultResolver(fname string) Resolver {
	return func(ctx Context) (interface{}, error) {
		t := reflect.TypeOf(ctx.Parent())
		v := reflect.ValueOf(ctx.Parent())
		for i := 0; i < t.NumField(); i++ {
			// Get the field, returns https://golang.org/pkg/reflect/#StructField
			field := t.Field(i)
			// Get the field tag value
			// TODO: check 'gql' tag first and if that does not exist, check 'json'
			// TODO: json tags need to be checked for additional options
			// 			like `,omitempty`, so can NOT use simple "==" op on them
			tag := field.Tag.Get("json")
			if strings.Split(tag, ",")[0] == fname {
				return v.FieldByName(field.Name).Interface(), nil
			}
		}
		return nil, nil
	}
}
