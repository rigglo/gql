package gql

import (
	"fmt"
	"reflect"

	"github.com/rigglo/gql/pkg/language/ast"
)

func typeToAst(t Type) ast.Type {
	switch t.GetKind() {
	case NonNullKind:
		return &ast.NonNullType{
			Type: typeToAst(t.(*NonNull).Wrapped),
		}
	case ListKind:
		return &ast.ListType{
			Type: typeToAst(t.(*List).Wrapped),
		}
	default:
		return &ast.NamedType{
			Name: t.GetName(),
		}
	}
}

func toAstValue(t Type, i interface{}) ast.Value {
	if i == nil {
		return &ast.NullValue{
			Value: "null",
		}
	}
	switch t := t.(type) {
	case *NonNull:
		return toAstValue(t.Wrapped, i)
	case *List:
		lv := ast.ListValue{
			Values: make([]ast.Value, 0),
		}
		rv := reflect.ValueOf(i)
		for li := 0; li < rv.Len(); li++ {
			lv.Values = append(lv.Values, toAstValue(t.Wrapped, rv.Index(li).Interface()))
		}
		return &lv
	case *Scalar:
		switch i := i.(type) {
		case string:
			return &ast.StringValue{
				Value: i,
			}
		case bool:
			return &ast.BooleanValue{
				Value: fmt.Sprintf("%v", i),
			}
		case float32, float64:
			return &ast.FloatValue{
				Value: fmt.Sprintf("%f", i),
			}
		default:
			return &ast.IntValue{
				Value: fmt.Sprintf("%v", i),
			}
		}
	case *Enum:
		return &ast.EnumValue{
			Value: fmt.Sprintf("%v", i),
		}
	case *InputObject:
		switch i := i.(type) {
		case map[string]interface{}:
			io := &ast.ObjectValue{
				Fields: make([]*ast.ObjectFieldValue, 0),
			}
			for fn, fv := range i {
				io.Fields = append(io.Fields, &ast.ObjectFieldValue{
					Name:  fn,
					Value: toAstValue(t.Fields[fn].Type, fv),
				})
			}
			return io
		}
	}
	return nil
}

type sdlBuilder struct {
	schema        *Schema
	defs          []ast.Definition
	typeDefs      map[string]bool
	directiveDefs map[string]bool
}

func newSDLBuilder(s *Schema) *sdlBuilder {
	return &sdlBuilder{
		schema: s,
		defs:   make([]ast.Definition, 0),
		directiveDefs: map[string]bool{
			"skip":       true,
			"include":    true,
			"deprecated": true,
		},
		typeDefs: map[string]bool{
			"String":   true,
			"Boolean":  true,
			"Int":      true,
			"ID":       true,
			"Float":    true,
			"DateTime": true,
		},
	}
}

func (b *sdlBuilder) Build() string {
	b.defs = make([]ast.Definition, 0)
	if b.schema == nil {
		return ""
	}
	sdef := &ast.SchemaDefinition{
		Directives:     make([]*ast.Directive, 0),
		RootOperations: make(map[ast.OperationType]*ast.NamedType),
	}
	if b.schema.Query != nil {
		sdef.RootOperations[ast.Query] = &ast.NamedType{
			Name: b.schema.Query.Name,
		}
		b.visitType(b.schema.Query)
	}
	if b.schema.Mutation != nil {
		sdef.RootOperations[ast.Mutation] = &ast.NamedType{
			Name: b.schema.Mutation.Name,
		}
		b.visitType(b.schema.Mutation)
	}
	if b.schema.Subscription != nil {
		sdef.RootOperations[ast.Subscription] = &ast.NamedType{
			Name: b.schema.Subscription.Name,
		}
		b.visitType(b.schema.Subscription)
	}
	for _, t := range b.schema.AdditionalTypes {
		b.visitType(t)
	}

	b.defs = append(b.defs, sdef)
	out := ""
	for i, def := range b.defs {
		if i != 0 {
			out += "\n"
		}
		out += def.String()
	}
	return out
}

func (b *sdlBuilder) visitType(t Type) {
	if _, ok := b.typeDefs[t.GetName()]; ok {
		return
	}
	b.typeDefs[t.GetName()] = true
	switch t := t.(type) {
	case *NonNull:
		b.visitType(t.Wrapped)
	case *List:
		b.visitType(t.Wrapped)
	case *Scalar:
		b.defs = append(b.defs, &ast.ScalarDefinition{
			Name:        t.Name,
			Description: t.Description,
			Directives:  t.Directives.ast(),
		})
	case *Object:
		def := &ast.ObjectDefinition{
			Name:        t.Name,
			Description: t.Description,
			Implements:  make([]*ast.NamedType, 0),
			Directives:  t.Directives.ast(),
			Fields:      make([]*ast.FieldDefinition, 0),
		}
		for _, i := range t.Implements {
			def.Implements = append(def.Implements, &ast.NamedType{
				Name: i.Name,
			})
		}
		for fn, f := range t.Fields {
			fdef := &ast.FieldDefinition{
				Name:        fn,
				Description: f.Description,
				Type:        typeToAst(f.Type),
				Directives:  f.Directives.ast(),
				Arguments:   make([]*ast.InputValueDefinition, 0),
			}
			for an, a := range f.Arguments {
				fdef.Arguments = append(fdef.Arguments, &ast.InputValueDefinition{
					Name:         an,
					Type:         typeToAst(a.Type),
					DefaultValue: toAstValue(a.Type, a.DefaultValue),
					Description:  a.Description,
				})
			}
			def.Fields = append(def.Fields, fdef)
			b.visitType(f.Type)
		}
		b.defs = append(b.defs, def)
	case *Enum:
		def := &ast.EnumDefinition{
			Name:        t.Name,
			Description: t.Description,
			Values:      make([]*ast.EnumValueDefinition, len(t.Values)),
		}
		for _, v := range t.Values {
			def.Values = append(def.Values, &ast.EnumValueDefinition{
				Description: v.Description,
				Value: &ast.EnumValue{
					Value: v.Name,
				},
				Directives: make([]*ast.Directive, 0),
			})
		}
		b.defs = append(b.defs, def)
	case *Interface:
		def := &ast.InterfaceDefinition{
			Name:        t.Name,
			Description: t.Description,
			Directives:  t.Directives.ast(),
			Fields:      make([]*ast.FieldDefinition, 0),
		}
		for fn, f := range t.Fields {
			fdef := &ast.FieldDefinition{
				Name:        fn,
				Description: f.Description,
				Type:        typeToAst(f.Type),
				Directives:  f.Directives.ast(),
				Arguments:   make([]*ast.InputValueDefinition, 0),
			}

			def.Fields = append(def.Fields, fdef)
			b.visitType(f.Type)
		}
		b.defs = append(b.defs, def)
	case *Union:
		def := &ast.UnionDefinition{
			Name:        t.Name,
			Description: t.Description,
			Directives:  t.Directives.ast(),
			Members:     make([]*ast.NamedType, 0),
		}
		for _, m := range t.Members {
			def.Members = append(def.Members, &ast.NamedType{
				Name: m.GetName(),
			})
			b.visitType(m)
		}
		b.defs = append(b.defs, def)
	case *InputObject:
		def := &ast.InputObjectDefinition{
			Name:        t.Name,
			Description: t.Description,
			Directives:  t.Directives.ast(),
			Fields:      make([]*ast.InputValueDefinition, 0),
		}
		for fn, f := range t.Fields {
			def.Fields = append(def.Fields, &ast.InputValueDefinition{
				Name:        fn,
				Description: f.Description,
				Type:        typeToAst(f.Type),
				Directives:  f.Directives.ast(),
			})
			b.visitType(f.Type)
		}
		b.defs = append(b.defs, def)
	}
}

func (b *sdlBuilder) visitDirective(d Directive) {
	if _, ok := b.directiveDefs[d.GetName()]; ok {
		return
	}
}
