package gql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rigglo/gql/pkg/language/ast"
)

func validate(ctx *gqlCtx) {
	// TODO: parse type system definitions and return error that those are not supported in a query

	ops := map[string]bool{}

	for _, f := range ctx.doc.Fragments {
		if _, ok := ctx.fragments[f.Name]; ok {
			ctx.addErr(&Error{fmt.Sprintf("Fragment name '%s' is not unique, it's already used", f.Name), nil, nil, nil})
			continue
		}
		if t, ok := ctx.types[f.TypeCondition]; !ok {
			ctx.addErr(&Error{fmt.Sprintf("Invalid fragment target '%s' for fragment '%s', type does not exist", f.TypeCondition, f.Name), nil, nil, nil})
			continue
		} else {
			if !isCompositeType(t) {
				ctx.addErr(&Error{fmt.Sprintf("Invalid fragment target '%s' for fragment '%s', type is not composite", f.TypeCondition, f.Name), nil, nil, nil})
				continue
			}
		}
		// validateDirectives(ctx, o, f.Directives, FragmentDefinitionLoc)
		ctx.fragments[f.Name] = f
		ctx.fragmentUsage[f.Name] = false
	}

	for _, o := range ctx.doc.Operations {
		// 5.2.1 - Named Operation Definitions
		if _, ok := ops[o.Name]; ok && o.Name != "" {
			ctx.addErr(&Error{fmt.Sprintf(errValidateOperationName, o.Name), nil, nil, nil})
		} else {
			ops[o.Name] = true
		}

		// 5.2.2 - Anonymous Operation Definitions
		if o.Name == "" && len(ctx.doc.Operations) > 1 {
			ctx.addErr(&Error{fmt.Sprintf(errAnonymousOperationDefinitions), nil, nil, nil})
		}

		// validate varibles
		ctx.variableDefs[o.Name] = validateVariables(ctx, o)
		ctx.variableUsages[o.Name] = map[string]struct{}{}

		// 5.3 - Fields
		switch o.OperationType {
		case ast.Query:
			if ctx.schema.Query == nil {
				ctx.addErr(&Error{fmt.Sprintf("No root query defined in schema"), nil, nil, nil})
				break
			}
			validateDirectives(ctx, o, o.Directives, QueryLoc)
			validateSelectionSet(ctx, o, o.SelectionSet, ctx.schema.Query, []string{})
		case ast.Mutation:
			if ctx.schema.Mutation == nil {
				ctx.addErr(&Error{fmt.Sprintf("No root mutation defined in schema"), nil, nil, nil})
				break
			}
			validateDirectives(ctx, o, o.Directives, MutationLoc)
			validateSelectionSet(ctx, o, o.SelectionSet, ctx.schema.Mutation, []string{})
		case ast.Subscription:
			if ctx.schema.Subscription == nil {
				ctx.addErr(&Error{fmt.Sprintf("No root subscription defined in schema"), nil, nil, nil})
				break
			}
			if len(o.SelectionSet) != 1 {
				ctx.addErr(&Error{fmt.Sprintf("Subscriptions must have only one root field in the selection set"), nil, nil, nil})
				break
			}
			validateDirectives(ctx, o, o.Directives, SubscriptionLoc)
			validateSelectionSet(ctx, o, o.SelectionSet, ctx.schema.Subscription, []string{})
		}
		for _, vDef := range ctx.variableDefs[o.Name] {
			if _, ok := ctx.variableUsages[o.Name][vDef.Name]; !ok {
				ctx.addErr(&Error{fmt.Sprintf("Variable defined but not used"), nil, nil, nil})
			}
		}
	}

	for fragName, used := range ctx.fragmentUsage {
		if !used {
			ctx.addErr(&Error{fmt.Sprintf("fragment '%s' is not used", fragName), nil, nil, nil})
		}
	}
}

func validateMetaField(ctx *gqlCtx, op *ast.Operation, f *ast.Field, t Type, visitedFrags []string) {
	switch f.Name {
	case "__typename":
		{
			// TODO: raise errors..
		}
	case "__schema":
		{
			if rq := ctx.schema.Query; rq != nil {
				if rq == t {
					validateSelectionSet(ctx, op, f.SelectionSet, schemaIntrospection, visitedFrags)
				} else {
					// TODO: raise errors
				}
			} else {
				// TODO: raise errors
			}
		}
	case "__type":
		{
			if rq := ctx.schema.Query; rq != nil {
				if rq == t {
					validateSelectionSet(ctx, op, f.SelectionSet, t, visitedFrags)
				} else {
					// TODO: raise errors
				}
			} else {
				// TODO: raise errors
			}
		}
	default:
		{
			ctx.addErr(&Error{fmt.Sprintf(errFieldDoesNotExist, f.Name, t.GetName()), nil, nil, nil})
		}
	}
}

func fieldsInSetCanMerge(ctx *gqlCtx, set []ast.Selection, t Type) {
	fieldsForName := collectFieldsForValidation(ctx, t, set, []string{})
	for _, fields := range fieldsForName {
		if len(fields) > 1 {
			for i := 1; i < len(fields); i++ {
				var pa, pb Type
				pa = ctx.types[fields[0].ParentType]
				pb = ctx.types[fields[i].ParentType]

				if !sameResponseShape(ctx, fields[0], fields[i], pa, pb) {
					ctx.addErr(&Error{fmt.Sprintf(errResponseShapeMismatch, "response shape is not the same"), nil, nil, nil})
					continue
				}

				// this is bad, we should check the PARENT TYPE..
				if reflect.DeepEqual(pa, pb) || (pa.GetKind() != ObjectKind || pb.GetKind() != ObjectKind) {
					if fields[0].Name != fields[i].Name {
						ctx.addErr(&Error{fmt.Sprintf(errResponseShapeMismatch, "field names are not equal"), nil, nil, nil})
						continue
					}

					if !equalArguments(fields[0].Arguments, fields[i].Arguments) {
						ctx.addErr(&Error{fmt.Sprintf(errResponseShapeMismatch, "arguments don't match"), nil, nil, nil})
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

func equalType(a Type, b Type) bool {
	if a.GetName() != b.GetName() {
		return false
	}
	if wa, ok := a.(WrappingType); ok {
		return equalType(wa.Unwrap(), b.(WrappingType).Unwrap())
	}
	return true
}

func sameResponseShape(ctx *gqlCtx, fa *ast.Field, fb *ast.Field, pa Type, pb Type) bool {
	var typeA, typeB Type
	typeA = pa.(hasFields).GetFields()[fa.Name].Type
	typeB = pb.(hasFields).GetFields()[fb.Name].Type

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

func collectFieldsForValidation(ctx *gqlCtx, t Type, ss []ast.Selection, vFrags []string) map[string]ast.Fields {
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

				fragment, ok := ctx.fragments[fSpread.Name]
				if !ok {
					continue
				}

				fgfields := collectFieldsForValidation(ctx, ctx.types[fragment.TypeCondition], fragment.SelectionSet, vFrags)
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
					fragmentType = ctx.types[f.TypeCondition]
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

func validateSelectionSet(ctx *gqlCtx, op *ast.Operation, set []ast.Selection, t Type, visitedFrags []string) {
	fieldsInSetCanMerge(ctx, set, t)

	if len(set) == 0 {
		ctx.addErr(&Error{fmt.Sprintf(errLeafFieldSelectionsSelectionMissing, t.GetName()), nil, nil, nil})
		return
	}

	for _, s := range set {
		if s.Kind() == ast.FieldSelectionKind {
			f := s.(*ast.Field)
			t := t

			// check if it's a meta field then check if that meta field can be queried on the specific type or does it even exists
			if strings.HasPrefix(f.Name, "__") {
				if f.Name == "__typename" {
					validateMetaField(ctx, op, f, t, visitedFrags)
					continue
				} else if rq := ctx.schema.Query; rq != nil {
					if rq == t {
						t = introspectionQuery
					} else {
						// TODO: INVALID META FIELD, FIELD DOES NOT EXIST ON TYPE
					}
				}
			}

			// check if the type 't' is an Object
			if o, ok := t.(*Object); ok {
				if tf, ok := o.Fields[f.Name]; ok {
					// 5.3.3 - Leaf Field Selections
					selType := unwrapper(tf.GetType())

					validateArguments(ctx, op, f.Arguments, tf.Arguments)
					validateDirectives(ctx, op, f.Directives, FieldLoc)

					if isCompositeType(selType) {
						validateSelectionSet(ctx, op, f.SelectionSet, selType, visitedFrags)
					} else if !isCompositeType(selType) {
						if len(f.SelectionSet) != 0 {
							ctx.addErr(&Error{fmt.Sprintf(errLeafFieldSelectionsSelectionNotAllowed, selType.GetName()), nil, nil, nil})
						}
					}
				} else {
					// 5.3.1 - Field Selections on Objects, Interfaces, and Unions Types
					// if field does NOT exist on type
					ctx.addErr(&Error{fmt.Sprintf(errFieldDoesNotExist, f.Name, t.GetName()), nil, nil, nil})
					continue
				}
			} else if i, ok := t.(*Interface); ok {
				if tf, ok := i.Fields[f.Name]; ok {
					validateArguments(ctx, op, f.Arguments, tf.Arguments)
					validateDirectives(ctx, op, f.Directives, FieldLoc)

					// 5.3.3 - Leaf Field Selections
					selType := unwrapper(tf.GetType())
					if isCompositeType(selType) {
						validateSelectionSet(ctx, op, f.SelectionSet, selType, visitedFrags)
					} else if !isCompositeType(selType) {
						if len(f.SelectionSet) != 0 {
							ctx.addErr(&Error{fmt.Sprintf(errLeafFieldSelectionsSelectionNotAllowed, selType.GetName()), nil, nil, nil})
						}
					}

				} else {
					// 5.3.1 - Field Selections on Objects, Interfaces, and Unions Types
					// if field does NOT exist on type
					ctx.addErr(&Error{fmt.Sprintf(errFieldDoesNotExist, f.Name, t.GetName()), nil, nil, nil})
					continue
				}
			} else {
				ctx.addErr(&Error{fmt.Sprintf("Invalid field selection on type '%s'", t.GetName()), nil, nil, nil})
				continue
			}
		} else if s.Kind() == ast.FragmentSpreadSelectionKind {
			f := s.(*ast.FragmentSpread)
			cycle := false
			for _, v := range visitedFrags {
				if v == f.Name {
					ctx.addErr(&Error{fmt.Sprintf("fragment cycle detected for fragment '%s'", f.Name), nil, nil, nil})
					cycle = true
				}
			}
			if cycle {
				continue
			}

			fDef, ok := ctx.fragments[f.Name]
			if !ok {
				ctx.addErr(&Error{fmt.Sprintf("fragment '%s' is not defined in query", f.Name), nil, nil, nil})
				continue
			}
			tCond, ok := ctx.types[fDef.TypeCondition]
			if !ok {
				ctx.addErr(&Error{fmt.Sprintf("fragment's (%s) target type (%s) is not defined in query", f.Name, fDef.TypeCondition), nil, nil, nil})
				continue
			}
			if !isPossibleSpread(ctx, t, tCond) {
				ctx.addErr(&Error{fmt.Sprintf("cannot use '%s' spead on type '%s'", f.Name, t.GetName()), nil, nil, nil})
				continue
			}

			validateDirectives(ctx, op, f.Directives, FragmentSpreadLoc)
			ctx.fragmentUsage[f.Name] = true
			validateSelectionSet(ctx, op, fDef.SelectionSet, tCond, append(visitedFrags, f.Name))
		} else if s.Kind() == ast.InlineFragmentSelectionKind {
			fDef := s.(*ast.InlineFragment)
			tCond, ok := ctx.types[fDef.TypeCondition]
			if !ok && fDef.TypeCondition != "" {
				ctx.addErr(&Error{fmt.Sprintf("fragment's target type (%s) is not defined in query", fDef.TypeCondition), nil, nil, nil})
				continue
			} else if fDef.TypeCondition == "" {
				tCond = t
			}
			if !isPossibleSpread(ctx, t, tCond) {
				ctx.addErr(&Error{fmt.Sprintf("invalid use of inline fragment on type '%s': target does not match", t.GetName()), nil, nil, nil})
				continue
			}

			validateDirectives(ctx, op, fDef.Directives, InlineFragmentLoc)
			validateSelectionSet(ctx, op, fDef.SelectionSet, tCond, visitedFrags)
		}
	}
}

func isPossibleSpread(ctx *gqlCtx, parentType Type, fragType Type) bool {
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

func getPossibleTypes(ctx *gqlCtx, t Type) []Type {
	switch t.GetKind() {
	case ObjectKind:
		return []Type{t}
	case InterfaceKind:
		return ctx.implementors[t.GetName()]
	case UnionKind:
		return t.(*Union).GetMembers()
	}
	return []Type{}
}

func validateArguments(ctx *gqlCtx, op *ast.Operation, astArgs []*ast.Argument, args []*Argument) {
	visitesArgs := map[string]*ast.Argument{}
	for _, a := range astArgs {
		for _, ta := range args {
			if a.Name == ta.Name {
				if _, visited := visitesArgs[a.Name]; visited {
					ctx.addErr(&Error{fmt.Sprintf("argument '%s' is set multiple times", a.Name), nil, nil, nil})
				} else {
					if ta.IsDefaultValueSet() && ta.Type.GetKind() == NonNullKind {

					}
					// TODO: don't raise error when arg is NonNull, has default value and the provided value is null
					validateValue(ctx, op, ta.Type, a.Value)
					visitesArgs[a.Name] = a
				}
				break
			}
		}
		if _, ok := visitesArgs[a.Name]; !ok {
			ctx.addErr(&Error{fmt.Sprintf("argument '%s' is not defined", a.Name), nil, nil, nil})
		}
	}

	for _, a := range args {
		if a.Type.GetKind() == NonNullKind && !a.IsDefaultValueSet() {
			if astArg, ok := visitesArgs[a.Name]; ok {
				if astArg.Value.Kind() == ast.NullValueKind {
					ctx.addErr(&Error{fmt.Sprintf("argument '%s' is NonNull and the provided value is null", a.Name), nil, nil, nil})
				}
			} else {
				ctx.addErr(&Error{fmt.Sprintf("argument '%s' is required (NonNull) but not provided", a.Name), nil, nil, nil})
			}
		}
	}
}

func validateDirectives(ctx *gqlCtx, op *ast.Operation, ds []*ast.Directive, loc DirectiveLocation) {
	visited := map[string]bool{}
	for _, d := range ds {
		if def, ok := ctx.directives[d.Name]; ok {
			ok = false
			for _, l := range def.GetLocations() {
				if l == loc {
					ok = true
				}
			}
			if !ok {
				// DIRECTIVE IS ON INVALID LOCATION
				ctx.addErr(&Error{fmt.Sprintf("directive '%s' is on invalid location", d.Name), nil, nil, nil})
				continue
			}
			if _, ok = visited[d.Name]; ok {
				// DIRECTIVE IS NOT UNIQUE PER LOCATION
				ctx.addErr(&Error{fmt.Sprintf("directive '%s' is not unique per location", d.Name), nil, nil, nil})
				continue
			}
			validateArguments(ctx, op, d.Arguments, def.GetArguments())
			visited[d.Name] = true
		} else {
			// NOT DEFINED
			ctx.addErr(&Error{fmt.Sprintf("directive '%s' is not defined", d.Name), nil, nil, nil})
		}
	}
}

func validateValue(ctx *gqlCtx, op *ast.Operation, t Type, val ast.Value) {
	switch {
	case val.Kind() == ast.VariableValueKind:
		vv := val.(*ast.VariableValue)
		for _, vDef := range op.Variables {
			if vDef.Name == vv.Name {
				ctx.variableUsages[op.Name][vv.Name] = struct{}{}
				vdefType, err := resolveAstType(ctx.types, vDef.Type)
				if err != nil {
					return
				}

				if equalType(vdefType, t) {
					return
				} else if t.GetKind() == NonNullKind {
					if equalType(t.(*NonNull).Unwrap(), vdefType) && vDef.DefaultValue != nil && vDef.DefaultValue.Kind() != ast.NullValueKind {
						return
					}
				}
				ctx.addErr(&Error{fmt.Sprintf("variable '%s' is not allowed to use", vv.Name), []*ErrorLocation{{Line: vv.Location.Line, Column: vv.Location.Column}}, nil, nil})
				return
			}
		}
		ctx.addErr(&Error{fmt.Sprintf("variable '%s' is defined", vv.Name), []*ErrorLocation{{Line: vv.Location.Line, Column: vv.Location.Column}}, nil, nil})
		return
	case t.GetKind() == NonNullKind:
		if vv, ok := val.(*ast.NullValue); ok {
			ctx.addErr(&Error{fmt.Sprintf("null value provided for NonNull type"), []*ErrorLocation{{Line: vv.Location.Line, Column: vv.Location.Column}}, nil, nil})
			return
		}
		validateValue(ctx, op, t.(*NonNull).Unwrap(), val)
		return
	case val.Kind() == ast.NullValueKind:
		return
	case t.GetKind() == ListKind:
		lv := val.(*ast.ListValue)
		for i := 0; i < len(lv.Values); i++ {
			validateValue(ctx, op, t.(*List).Unwrap(), val)
		}
		return
	case t.GetKind() == ScalarKind:
		s := t.(*Scalar)
		if v, ok := val.GetValue().(string); ok {
			if _, err := s.CoerceInput([]byte(v)); err != nil {
				ctx.addErr(&Error{fmt.Sprintf("couldn't coerce value: %s", err.Error()), []*ErrorLocation{{Line: val.GetLocation().Line, Column: val.GetLocation().Column}}, nil, nil})
			}
			return
		}
		ctx.addErr(&Error{fmt.Sprintf("invalid value for Scalar"), []*ErrorLocation{{Line: val.GetLocation().Line, Column: val.GetLocation().Column}}, nil, nil})
		return
	case t.GetKind() == EnumKind:
		if val.Kind() != ast.EnumValueKind {
			ctx.addErr(&Error{fmt.Sprintf("invalid value for Enum"), []*ErrorLocation{{Line: val.GetLocation().Line, Column: val.GetLocation().Column}}, nil, nil})
		}
		e := t.(*Enum)
		if v, ok := val.GetValue().(string); ok {
			for _, ev := range e.Values {
				if ev.Name == v {
					return
				}
			}
			ctx.addErr(&Error{fmt.Sprintf("invalid enum value '%s'", v), []*ErrorLocation{{Line: val.GetLocation().Line, Column: val.GetLocation().Column}}, nil, nil})
			return
		}
		ctx.addErr(&Error{fmt.Sprintf("invalid value for Enum"), []*ErrorLocation{{Line: val.GetLocation().Line, Column: val.GetLocation().Column}}, nil, nil})
		return
	case t.GetKind() == InputObjectKind:
		ov, ok := val.(*ast.ObjectValue)
		if !ok {
			ctx.addErr(&Error{fmt.Sprintf("invalid value for InputObject"), []*ErrorLocation{{Line: val.GetLocation().Line, Column: val.GetLocation().Column}}, nil, nil})
			return
		}
		o := t.(*InputObject)
		visitedFields := map[string]struct{}{}
		for _, astf := range ov.Fields {
			field, ok := o.Fields[astf.Name]
			if !ok {
				ctx.addErr(&Error{fmt.Sprintf("field '%s' is not defined", astf.Name), []*ErrorLocation{{Line: astf.GetLocation().Line, Column: astf.GetLocation().Column}}, nil, nil})
				continue
			}

			if _, ok := visitedFields[astf.Name]; ok {
				ctx.addErr(&Error{fmt.Sprintf("field '%s' was set multiple times", astf.Name), []*ErrorLocation{{Line: val.GetLocation().Line, Column: val.GetLocation().Column}}, nil, nil})
				continue
			}
			visitedFields[astf.Name] = struct{}{}

			if astf.Value.Kind() == ast.NullValueKind && field.Type.GetKind() != NonNullKind {
				continue
			}
			if astf.Value.Kind() != ast.VariableValueKind {
				validateValue(ctx, op, field.Type, astf.Value)
				continue
			}
			if astf.Value.Kind() == ast.VariableValueKind {
				vv := astf.Value.(*ast.VariableValue)
				ok := false
				for _, vDef := range op.Variables {
					if vDef.Name == vv.Name {
						ctx.variableUsages[op.Name][vv.Name] = struct{}{}
						vdefType, err := resolveAstType(ctx.types, vDef.Type)
						if err != nil {
							ctx.addErr(err)
							ok = true
							break
						}
						if equalType(vdefType, field.Type) {
							ok = true
							break
						}
						ctx.addErr(&Error{fmt.Sprintf("invalid variable type for field '%s'", astf.Name), []*ErrorLocation{{Line: val.GetLocation().Line, Column: val.GetLocation().Column}}, nil, nil})
						ok = true
						break
					}
				}
				if !ok {
					ctx.addErr(&Error{fmt.Sprintf("variable '%s' is defined", vv.Name), []*ErrorLocation{{Line: vv.Location.Line, Column: vv.Location.Column}}, nil, nil})
				}
			}
		}

		for fn, field := range o.Fields {
			if _, ok := visitedFields[fn]; !ok {
				if !field.IsDefaultValueSet() && field.Type.GetKind() == NonNullKind {
					ctx.addErr(&Error{fmt.Sprintf("no value provided for field '%s' with NonNull type", fn), []*ErrorLocation{{Line: ov.Location.Line, Column: ov.Location.Column}}, nil, nil})
				}
			}
		}
	}
}

func validateVariables(ctx *gqlCtx, op *ast.Operation) map[string]*ast.Variable {
	visited := map[string]*ast.Variable{}
	for _, v := range op.Variables {
		// variable has to be unique
		if _, ok := visited[v.Name]; ok {
			ctx.addErr(&Error{fmt.Sprintf("variable '%s' is used multiple times", v.Name), nil, nil, nil})
			continue
		}

		// variable has to be input type
		if vt, err := resolveAstType(ctx.types, v.Type); err != nil {
			ctx.addErr(err)
			continue
		} else if !isInputType(vt) {
			ctx.addErr(&Error{fmt.Sprintf("variable '%s' is not an input type", v.Name), nil, nil, nil})
			continue
		} else if v.DefaultValue != nil {
			// default value has to be validated
			validateValue(ctx, op, vt, v.DefaultValue)
		}
		visited[v.Name] = v
	}
	return visited
}
