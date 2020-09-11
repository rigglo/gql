package federation

import (
	"context"

	"github.com/rigglo/gql"
)

type (
	federation struct {
		Schema   *gql.Schema
		Types    map[string]*federatedType
		Entities gql.Members
	}

	federatedType struct {
		RequiredFields map[string]gql.Type
		Type           gql.Type
	}
)

// Federate adds the Apollo Federation fields and types to the schema, specified in the Apollo Federation specification
// https://www.apollographql.com/docs/apollo-server/federation/federation-spec
func Federate(s *gql.Schema) *gql.Schema {
	f := loadFederation(s)
	s.Query.Fields["_service"] = &gql.Field{
		Type: &gql.Object{
			Name:        "_Service",
			Description: "A Federated service",
			Fields: gql.Fields{
				"sdl": &gql.Field{
					Type:        gql.String,
					Description: "SDL of the service",
					Resolver: func(ctx gql.Context) (interface{}, error) {
						return s.SDL(), nil
					},
				},
			},
		},
		Resolver: func(ctx gql.Context) (interface{}, error) {
			return struct{}{}, nil
		},
	}
	s.Query.Fields["_entities"] = &gql.Field{
		Type: gql.NewNonNull(gql.NewList(&gql.Union{
			Name:        "_Entity",
			Description: "A Federated service",
			Members:     f.Entities,
			TypeResolver: func(ctx context.Context, v interface{}) *gql.Object {
				if v, ok := v.(map[string]interface{}); ok {
					if t, ok := v["__typename"]; ok {
						if tname, ok := t.(string); ok {
							if o, ok := f.Types[tname].Type.(*gql.Object); ok {
								return o
							} else if i, ok := f.Types[tname].Type.(*gql.Interface); ok {
								return i.Resolve(ctx, v)
							}
						}
					}
				}
				return nil
			},
		})),
		Arguments: gql.Arguments{
			"representations": &gql.Argument{
				Type: gql.NewNonNull(gql.NewList(gql.NewNonNull(anyScalar))),
			},
		},
		Resolver: func(ctx gql.Context) (interface{}, error) {
			return ctx.Args()["representations"], nil
		},
	}
	return s
}

func loadFederation(s *gql.Schema) *federation {
	f := &federation{
		Schema:   s,
		Entities: gql.Members{},
		Types:    make(map[string]*federatedType),
	}
	for _, t := range s.AdditionalTypes {
		f.visit(unwrap(t))
	}
	if s.Query != nil {
		for _, q := range s.Query.Fields {
			f.visit(unwrap(q.Type))
		}
	}
	if s.Mutation != nil {
		for _, m := range s.Mutation.Fields {
			f.visit(unwrap(m.Type))
		}
	}
	return f
}

func (f *federation) visit(t gql.Type) {
	switch t := t.(type) {
	case *gql.Object:
		for _, d := range t.Directives {
			if _, ok := d.(*keyDirective); ok {
				if _, ok := f.Types[t.Name]; !ok {
					for _, field := range t.Fields {
						f.visit(unwrap(field.Type))
					}
					f.Types[t.Name] = &federatedType{
						Type: t,
					}
					f.Entities = append(f.Entities, t)
				}
			}
		}
	case *gql.Interface:
		for _, d := range t.Directives {
			if _, ok := d.(*keyDirective); ok {
				if _, ok := f.Types[t.Name]; !ok {
					for _, field := range t.Fields {
						f.visit(unwrap(field.Type))
					}
					f.Types[t.Name] = &federatedType{
						Type: t,
					}
					f.Entities = append(f.Entities, t)
				}
			}
		}
	}
}

func unwrap(t gql.Type) gql.Type {
	if t, ok := t.(gql.WrappingType); ok {
		return unwrap(t.Unwrap())
	}
	return t
}
