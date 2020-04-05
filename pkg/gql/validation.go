package gql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rigglo/gql/pkg/language/ast"
)

func validate(ctx *eCtx) {
	// TODO: parse type system definitions and return error that those are not supported in a query
	ops := map[string]bool{}
	for _, o := range ctx.Get(keyQuery).(*ast.Document).Operations {
		// 5.2.1 - Named Operation Definitions
		if _, ok := ops[o.Name]; ok && o.Name != "" {
			ctx.addErr(&Error{fmt.Sprintf(ErrValidateOperationName, o.Name), nil, nil, nil})
		} else {
			ops[o.Name] = true
		}

		// 5.2.2 - Anonymous Operation Definitions
		if o.Name == "" && len(ctx.Get(keyQuery).(*ast.Document).Operations) > 1 {
			ctx.addErr(&Error{fmt.Sprintf(ErrAnonymousOperationDefinitions), nil, nil, nil})
		}

		// TODO: 5.2.3 - Subscription Operation Definitions

		// 5.3 - Fields
		switch o.OperationType {
		case ast.Query:
			validateSelectionSet(ctx, o.SelectionSet, ctx.Get(keySchema).(*Schema).GetRootQuery())
		case ast.Mutation:
			validateSelectionSet(ctx, o.SelectionSet, ctx.Get(keySchema).(*Schema).GetRootMutation())
		}
		// validateSelectionSet(ctx, o)
	}
}

func validateMetaField(ctx *eCtx, f *ast.Field, t Type) {
	switch f.Name {
	case "__typename":
		{
			// TODO: validate meta fields
		}
	case "__schema":
		{
			// TODO: validate meta fields
		}
	case "__type":
		{
			// TODO: validate meta fields
		}
	default:
		{
			ctx.addErr(&Error{fmt.Sprintf(ErrFieldDoesNotExist, f.Name, t.GetName()), nil, nil, nil})
		}
	}
}

func fieldsInSetCanMerge(ctx *eCtx, set []ast.Selection, t Type) {
	types := ctx.Get(keyTypes).(map[string]Type)

	fieldsForName := collectFields(ctx, t.(*Object), set, []string{})
	for _, fields := range fieldsForName {
		if len(fields) > 1 {
			for i := 1; i < len(fields); i++ {
				if !sameResponseShape(ctx, fields[0], fields[i], t) {
					// TODO: raise error for selection set can not be merged
					ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrResponseShapeMismatch, "response shape is not the same")), nil, nil, nil})
				}

				var pa, pb Type
				pa = types[fields[0].ParentType]
				pb = types[fields[i].ParentType]

				// this is bad, we should check the PARENT TYPE..
				if reflect.DeepEqual(pa, pb) || (pa.GetKind() != ObjectKind || pb.GetKind() != ObjectKind) {
					if fields[0].Name != fields[i].Name {
						// TODO: raise error that selection set can not be merged
						ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrResponseShapeMismatch, "field names are not equal")), nil, nil, nil})
					}
					if !reflect.DeepEqual(fields[0].Arguments, fields[i].Arguments) {
						// TODO: raise error that selection set can not be merged (due to arguments don't match)
						ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrResponseShapeMismatch, "arguments don't match")), nil, nil, nil})
					}
					mergedSet := append(fields[0].SelectionSet, fields[i].SelectionSet...)
					fieldsInSetCanMerge(ctx, mergedSet, t)
				}
			}
		}
	}
}

func sameResponseShape(ctx *eCtx, fa *ast.Field, fb *ast.Field, t Type) bool {
	var typeA, typeB Type
	for _, f := range t.(hasFields).GetFields() {
		if f.GetName() == fa.Name {
			typeA = f.GetType()
		}
		if f.GetName() == fb.Name {
			typeB = f.GetType()
		}
	}

	for {
		if typeA.GetKind() == NonNullKind || typeB.GetKind() == NonNullKind {
			if typeA.GetKind() != NonNullKind || typeB.GetKind() != NonNullKind {
				return false
			}
			typeA = typeA.(*NonNull).Unwrap()
			typeB = typeB.(*NonNull).Unwrap()
		}

		if typeA.GetKind() == ListKind || typeB.GetKind() == ListKind {
			if typeA.GetKind() != ListKind || typeB.GetKind() != ListKind {
				return false
			}
			typeA = typeA.(*List).Unwrap()
			typeB = typeB.(*List).Unwrap()
			continue
		}
		break
	}

	if typeA.GetKind() == ScalarKind || typeB.GetKind() == ScalarKind || typeA.GetKind() == EnumKind || typeB.GetKind() == EnumKind {
		// TODO: Could try reflect.DeepEqual(typeA, typeB)
		if typeA.GetName() == typeB.GetName() {
			return true
		}
		return false
	}

	if !isCompositeType(typeA) || !isCompositeType(typeB) {
		return false
	}

	mergedSet := append(fa.SelectionSet, fb.SelectionSet...)
	fieldsForName := collectFields(ctx, typeA.(*Object), mergedSet, []string{})
	for _, fields := range fieldsForName {
		if len(fields) > 1 {
			for i := 1; i < len(fields); i++ {
				if !sameResponseShape(ctx, fields[0], fields[i], typeA) {
					return false
				}
			}
		}
	}
	return true
}

func isCompositeType(t Type) bool {
	if t.GetKind() == ObjectKind || t.GetKind() == InterfaceKind || t.GetKind() == UnionKind {
		return true
	}
	return false
}

func validateSelectionSet(ctx *eCtx, set []ast.Selection, t Type) {
	//types := ctx.Get(keyTypes).(map[string]Type)
	fieldsInSetCanMerge(ctx, set, t)
	for _, s := range set {
		if s.Kind() == ast.FieldSelectionKind {
			f := s.(*ast.Field)

			// check if it's a meta field then check if that meta field can be queried on the specific type or does it even exists
			if strings.HasPrefix(f.Name, "__") {
				validateMetaField(ctx, f, t)
			} else {
				// check if the type 't' is an Object
				if o, ok := t.(*Object); ok {
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
									ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrLeafFieldSelectionsSelectionMissing, selType.GetName())), nil, nil, nil})
								}
							} else if selType.GetKind() == ScalarKind || selType.GetKind() == EnumKind {
								if len(f.SelectionSet) != 0 {
									ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrLeafFieldSelectionsSelectionNotAllowed, selType.GetName())), nil, nil, nil})
								}
							}

							ok = true
						}
					}

					// 5.3.1 - Field Selections on Objects, Interfaces, and Unions Types
					// if field does NOT exist on type
					if !ok {
						ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrFieldDoesNotExist, f.Name, t.GetName())), nil, nil, nil})
					}
				} else if i, ok := t.(*Interface); ok {
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
									ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrLeafFieldSelectionsSelectionMissing, selType.GetName())), nil, nil, nil})
								}
							} else if selType.GetKind() == ScalarKind || selType.GetKind() == EnumKind {
								if len(f.SelectionSet) != 0 {
									ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrLeafFieldSelectionsSelectionNotAllowed, selType.GetName())), nil, nil, nil})
								}
							}

							ok = true
						}
					}

					// 5.3.1 - Field Selections on Objects, Interfaces, and Unions Types
					// if field does NOT exist on type
					if !ok {
						ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrFieldDoesNotExist, f.Name, t.GetName())), nil, nil, nil})
					}
				} else if _, ok := t.(*Union); ok {
					// TODO: add error, that field selection on Union type does not supported (only fragments)
				} else {
					// TODO: add error ..
				}
			}
		}
	}
}
