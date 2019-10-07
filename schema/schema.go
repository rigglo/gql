package schema

// Schema ...
type Schema struct {
	Mutation    *Object
	Query       *Object
	Subsciption *Object
	types       map[string]Type
}