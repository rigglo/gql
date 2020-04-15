package gql

import (
	"context"
)

// Members is an alias for a list of types that are set for a Union
type Members []Type

/*
Union represents an object that could be one of a list of graphql objects types,
but provides for no guaranteed fields between those types. They also differ from interfaces
in that Object types declare what interfaces they implement, but are not aware of what unions contain them
*/
type Union struct {
	Description  string
	Name         string
	Members      Members
	Directives   Directives
	TypeResolver TypeResolver
}

// GetDescription show the description of the Union type
func (u *Union) GetDescription() string {
	return u.Description
}

// GetName show the name of the Union type, "Union"
func (u *Union) GetName() string {
	return u.Name
}

// GetKind returns the kind of the type, UnionType
func (u *Union) GetKind() TypeKind {
	return UnionKind
}

// GetMembers returns the composite types contained by the union
func (u *Union) GetMembers() []Type {
	return u.Members
}

// GetDirectives returns all the directives applied to the Union type
func (u *Union) GetDirectives() []Directive {
	return u.Directives
}

// Resolve helps decide the executor which contained type to resolve
func (u *Union) Resolve(ctx context.Context, v interface{}) *Object {
	return u.TypeResolver(ctx, v)
}
