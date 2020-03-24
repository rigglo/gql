package schema

import (
	"fmt"
	"strings"

	"github.com/rigglo/gql/pkg/language/ast"
)

func validate(ctx *eCtx) {
	// TODO: parse type system definitions and return error that those are not supported in a query
	ops := map[string]bool{}
	for _, o := range ctx.Get(keyQuery).(*ast.Document).Operations {
		// 5.2.1 - Named Operation Definitions
		if _, ok := ops[o.Name]; ok && o.Name != "" {
			ctx.res.addErr(NewError(ctx, fmt.Sprintf(ErrValidateOperationName, o.Name), nil))
		} else {
			ops[o.Name] = true
		}

		// 5.2.2 - Anonymous Operation Definitions
		if o.Name == "" && len(ctx.Get(keyQuery).(*ast.Document).Operations) > 1 {
			ctx.res.addErr(NewError(ctx, fmt.Sprintf(ErrAnonymousOperationDefinitions), nil))
		}

		// TODO: 5.2.3 - Subscription Operation Definitions

		// 5.3 - Fields
		switch o.OperationType {
		case ast.Query:
			validateSelectionSet(ctx, o.SelectionSet, ctx.Get(keySchema).(Schema).GetRootQuery())
		case ast.Mutation:
			validateSelectionSet(ctx, o.SelectionSet, ctx.Get(keySchema).(Schema).GetRootMutation())
		}
		// validateSelectionSet(ctx, o)
	}
}

// unwrapper unwraps Type t, until it results in a real type, one of (scalar, enum, interface, union or an object)
func unwrapper(t Type) Type {
	if w, ok := t.(WrappingType); ok {
		return unwrapper(w.Unwrap())
	}
	return t
}

func validateMetaField(ctx *eCtx, f *ast.Field, t Type) {
	switch f.Name {
	case "__typename":
		{
			// TODO: validate meta fields
		}
	default:
		{
			ctx.res.addErr(NewError(ctx, fmt.Sprintf(ErrFieldDoesNotExist, f.Name, t.GetName()), nil))
		}
	}
}

func validateSelectionSet(ctx *eCtx, set []ast.Selection, t Type) {
	//types := ctx.Get(keyTypes).(map[string]Type)
	for _, s := range set {
		if s.Kind() == ast.FieldSelectionKind {
			f := s.(*ast.Field)

			// check if it's a meta field then check if that meta field can be queried on the specific type or does it even exists
			if strings.HasPrefix(f.Name, "__") {
				validateMetaField(ctx, f, t)
			} else {
				// check if the type 't' is an Object
				if o, ok := t.(ObjectType); ok {
					ok = false
					for _, tf := range o.GetFields() {

						// 5.3.1 - Field Selections on Objects, Interfaces, and Unions Types
						if tf.GetName() == f.Name {

							// 5.3.3 - Leaf Field Selections
							selType := unwrapper(tf.GetType())
							if selType.GetKind() == InterfaceKind || selType.GetKind() == UnionKind || selType.GetKind() == ObjectKind {
								if len(f.SelectionSet) > 0 {
									validateSelectionSet(ctx, f.SelectionSet, selType)
								} else {
									ctx.res.addErr(NewError(ctx, fmt.Sprintf(ErrLeafFieldSelectionsSelectionMissing, selType.GetName()), nil))
								}
							} else if selType.GetKind() == ScalarKind || selType.GetKind() == EnumKind {
								if len(f.SelectionSet) != 0 {
									ctx.res.addErr(NewError(ctx, fmt.Sprintf(ErrLeafFieldSelectionsSelectionNotAllowed, selType.GetName()), nil))
								}
							}

							ok = true
						}
					}

					// 5.3.1 - Field Selections on Objects, Interfaces, and Unions Types
					// if field does NOT exist on type
					if !ok {
						ctx.res.addErr(NewError(ctx, fmt.Sprintf(ErrFieldDoesNotExist, f.Name, t.GetName()), nil))
					}
				} else if i, ok := t.(InterfaceType); ok {
					ok = false
					for _, tf := range i.GetFields() {

						// 5.3.1 - Field Selections on Objects, Interfaces, and Unions Types
						if tf.GetName() == f.Name {

							// 5.3.3 - Leaf Field Selections
							selType := unwrapper(tf.GetType())
							if selType.GetKind() == InterfaceKind || selType.GetKind() == UnionKind || selType.GetKind() == ObjectKind {
								if len(f.SelectionSet) > 0 {
									validateSelectionSet(ctx, f.SelectionSet, selType)
								} else {
									ctx.res.addErr(NewError(ctx, fmt.Sprintf(ErrLeafFieldSelectionsSelectionMissing, selType.GetName()), nil))
								}
							} else if selType.GetKind() == ScalarKind || selType.GetKind() == EnumKind {
								if len(f.SelectionSet) != 0 {
									ctx.res.addErr(NewError(ctx, fmt.Sprintf(ErrLeafFieldSelectionsSelectionNotAllowed, selType.GetName()), nil))
								}
							}

							ok = true
						}
					}

					// 5.3.1 - Field Selections on Objects, Interfaces, and Unions Types
					// if field does NOT exist on type
					if !ok {
						ctx.res.addErr(NewError(ctx, fmt.Sprintf(ErrFieldDoesNotExist, f.Name, t.GetName()), nil))
					}
				} else if _, ok := t.(UnionType); ok {
					// TODO: add error, that field selection on Union type does not supported (only fragments)
				} else {
					// TODO: add error ..
				}
			}
		}
	}
}

func validateOperation(ctx *eCtx, o *ast.Operation) {

}
