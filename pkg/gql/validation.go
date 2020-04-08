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

	fieldsForName := collectFieldsForValidation(ctx, t, set, []string{})
	for _, fields := range fieldsForName {
		if len(fields) > 1 {
			for i := 1; i < len(fields); i++ {
				var pa, pb Type
				pa = types[fields[0].ParentType]
				pb = types[fields[i].ParentType]

				if !sameResponseShape(ctx, fields[0], fields[i], pa, pb) {
					ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrResponseShapeMismatch, "response shape is not the same")), nil, nil, nil})
					continue
				}

				// this is bad, we should check the PARENT TYPE..
				if reflect.DeepEqual(pa, pb) || (pa.GetKind() != ObjectKind || pb.GetKind() != ObjectKind) {
					if fields[0].Name != fields[i].Name {
						ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrResponseShapeMismatch, "field names are not equal")), nil, nil, nil})
						continue
					}

					if !equalArguments(fields[0].Arguments, fields[i].Arguments) {
						ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrResponseShapeMismatch, "arguments don't match")), nil, nil, nil})
						continue
					}
					mergedSet := append(fields[0].SelectionSet, fields[i].SelectionSet...)
					fieldsInSetCanMerge(ctx, mergedSet, t)
				}
			}
		}
	}
}

func equalArguments(a []*ast.Argument, b []*ast.Argument) bool {
	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}
	for _, fa := range a {
		found := false
		for _, fb := range b {
			if fa.Name == fb.Name && equalValue(fa.Value, fb.Value) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func equalValue(a ast.Value, b ast.Value) bool {
	if a.Kind() == b.Kind() {
		switch a.Kind() {
		case ast.VariableValueKind:
			if a.GetValue() != b.GetValue() {
				return false
			}
		default:
			if a.GetValue() != b.GetValue() {
				return false
			}
		}
		return true
	}
	return false
}

func sameResponseShape(ctx *eCtx, fa *ast.Field, fb *ast.Field, pa Type, pb Type) bool {
	var typeA, typeB Type
	for _, f := range pa.(hasFields).GetFields() {
		if f.GetName() == fa.Name {
			typeA = f.GetType()
		}
	}
	for _, f := range pb.(hasFields).GetFields() {
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
				if !sameResponseShape(ctx, fields[0], fields[i], pa, pb) {
					return false
				}
			}
		}
	}
	return true
}

func collectFieldsForValidation(ctx *eCtx, t Type, ss []ast.Selection, vFrags []string) map[string]ast.Fields {
	types := ctx.Get(keyTypes).(map[string]Type)

	if vFrags == nil {
		vFrags = []string{}
	}
	gfields := map[string]ast.Fields{}

	for _, sel := range ss {
		skip := false
		for _, d := range sel.GetDirectives() {
			if d.Name == "skip" {
				skip = skipDirective.Skip(d.Arguments)
			} else if d.Name == "include" {
				skip = !includeDirective.Include(d.Arguments)
			}
		}
		if skip {
			continue
		}

		switch sel.Kind() {
		case ast.FieldSelectionKind:
			{
				f := sel.(*ast.Field)
				f.ParentType = t.GetName()
				if _, ok := gfields[f.Alias]; ok {
					gfields[f.Alias] = append(gfields[f.Alias], f)
				} else {
					gfields[f.Alias] = ast.Fields{f}
				}
			}
		case ast.FragmentSpreadSelectionKind:
			{
				fSpread := sel.(*ast.FragmentSpread)
				skip := false
				for _, fragName := range vFrags {
					if fSpread.Name == fragName {
						skip = true
					}
				}
				if skip {
					continue
				}

				vFrags = append(vFrags, fSpread.Name)

				fragment, ok := ctx.Get(keyQuery).(*ast.Document).Fragments[fSpread.Name]
				if !ok {
					continue
				}

				fgfields := collectFieldsForValidation(ctx, types[fragment.TypeCondition], fragment.SelectionSet, vFrags)
				for rkey, fg := range fgfields {
					if _, ok := gfields[rkey]; ok {
						gfields[rkey] = append(gfields[rkey], fg...)
					} else {
						gfields[rkey] = fg
					}
				}
			}
		case ast.InlineFragmentSelectionKind:
			{
				f := sel.(*ast.InlineFragment)

				fragmentType := t
				if f.TypeCondition != "" {
					fragmentType = types[f.TypeCondition]
				}

				fgfields := collectFieldsForValidation(ctx, fragmentType, f.SelectionSet, vFrags)
				for rkey, fg := range fgfields {
					if _, ok := gfields[rkey]; ok {
						gfields[rkey] = append(gfields[rkey], fg...)
					} else {
						gfields[rkey] = fg
					}
				}
			}
		}
	}
	return gfields
}

func isCompositeType(t Type) bool {
	if t.GetKind() == ObjectKind || t.GetKind() == InterfaceKind || t.GetKind() == UnionKind {
		return true
	}
	return false
}

func validateSelectionSet(ctx *eCtx, set []ast.Selection, t Type) {
	fieldsInSetCanMerge(ctx, set, t)

	if len(set) == 0 {
		ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf(ErrLeafFieldSelectionsSelectionMissing, t.GetName())), nil, nil, nil})
		return
	}

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

							if isCompositeType(selType) {
								validateSelectionSet(ctx, f.SelectionSet, selType)
							} else if !isCompositeType(selType) {
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
							if isCompositeType(selType) {
								validateSelectionSet(ctx, f.SelectionSet, selType)
							} else if !isCompositeType(selType) {
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
				} else {
					ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf("Invalid selection set on type '%s'", t.GetName())), nil, nil, nil})
				}
			}
		} else if s.Kind() == ast.FragmentSpreadSelectionKind {
			f := s.(*ast.FragmentSpread)
			doc := ctx.Get(keyQuery).(*ast.Document)
			types := ctx.Get(keyTypes).(map[string]Type)
			fDef, ok := doc.Fragments[f.Name]
			if !ok {
				ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf("fragment '%s' is not defined in query", f.Name)), nil, nil, nil})
				continue
			}
			tCond, ok := types[fDef.TypeCondition]
			if !ok {
				ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf("fragment's (%s) target type (%s) is not defined in query", f.Name, fDef.TypeCondition)), nil, nil, nil})
				continue
			}
			if !isPossibleSpread(ctx, t, tCond) {
				ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf("cannot use '%s' spead on type '%s'", f.Name, t.GetName())), nil, nil, nil})
				continue
			}

			validateSelectionSet(ctx, fDef.SelectionSet, tCond)
		} else if s.Kind() == ast.InlineFragmentSelectionKind {
			fDef := s.(*ast.InlineFragment)
			types := ctx.Get(keyTypes).(map[string]Type)
			tCond, ok := types[fDef.TypeCondition]
			if !ok && fDef.TypeCondition != "" {
				ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf("fragment's target type (%s) is not defined in query", fDef.TypeCondition)), nil, nil, nil})
				continue
			} else if fDef.TypeCondition == "" {
				tCond = t
			}
			if !isPossibleSpread(ctx, t, tCond) {
				ctx.addErr(&Error{fmt.Sprintf(fmt.Sprintf("invalid use of inline fragment on type '%s': target does not match", t.GetName())), nil, nil, nil})
				continue
			}

			validateSelectionSet(ctx, fDef.SelectionSet, tCond)
		}
	}
}

func isPossibleSpread(ctx *eCtx, parentType Type, fragType Type) bool {
	pts := getPossibleTypes(ctx, parentType)
	fts := getPossibleTypes(ctx, fragType)
	for _, t := range pts {
		for _, ft := range fts {
			if t == ft {
				return true
			}
		}
	}
	return false
}

func getPossibleTypes(ctx *eCtx, t Type) []Type {
	switch t.GetKind() {
	case ObjectKind:
		return []Type{t}
	case InterfaceKind:
		implementors := ctx.Get(keyImplementors).(map[string][]Type)
		return implementors[t.GetName()]
	case UnionKind:
		return t.(*Union).GetMembers()
	}
	return []Type{}
}
