package schema

import "github.com/rigglo/gql/pkg/language/ast"

func resolveMetaFields(ctx *eCtx, f *ast.Field, t Type, res map[string]interface{}) {
	switch f.Name {
	case "__typename":
		res[f.Alias] = t.GetName()
	}
}
