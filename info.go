package gql

type ResolverInfo struct {
	Path   []interface{}
	Parent interface{}
	Field  Field
}
