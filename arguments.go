package gql

type Argument struct {
	Name         string
	Type         Type
	DefaultValue string
}

type Arguments map[string]*Argument

type ResolverArgs interface {
	Get(string) interface{}
}
