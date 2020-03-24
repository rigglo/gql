package schema

type Schema interface {
	GetRootQuery() ObjectType
	GetRootMutation() ObjectType
	GetRootSubsciption() ObjectType
	GetDirectives() []Directive
}

func getTypes(s Schema) map[string]Type {
	// TODO: getTypes from the Schema into a map[string]Type
	types := map[string]Type{}
	typeWalker(types, s.GetRootQuery())
	typeWalker(types, s.GetRootMutation())
	return types
}

func typeWalker(types map[string]Type, t Type) {
	wt := unwrapper(t)
	if _, ok := types[wt.GetName()]; !ok {
		types[wt.GetName()] = wt
	}
	if hf, ok := t.(HasFields); ok {
		for _, f := range hf.GetFields() {
			typeWalker(types, f.GetType())
		}
	} else if u, ok := wt.(UnionType); ok {
		for _, m := range u.GetMembers() {
			typeWalker(types, m)
		}
	}
}

type Operation interface {
	GetName() string
	GetFields() []Field
	GetDescription() string
}
