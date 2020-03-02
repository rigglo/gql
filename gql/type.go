package gql

import "github.com/rigglo/gql/schema"

type Type interface {
	GetName() string
	GetDescription() string
	GetKind() schema.TypeKind
}
