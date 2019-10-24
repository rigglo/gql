package gql

import (
	"fmt"
	"github.com/rigglo/gql/language/ast"
)

func newValidator(s *Schema) *validator {
	return &validator{
		schema: s,
		e: &executor{
			schema: s,
		},
	}
}

type validator struct {
	schema *Schema
	e      *executor
}

func (v *validator) Validate(ctx *gqlCtx) (*executor, []error) {
	return v.e, v.ValidateRequest(ctx)
}

func (v *validator) ValidateRequest(ctx *gqlCtx) []error {
	errs := v.ValidateOperations(ctx)
	return errs
}

func (v *validator) ValidateOperations(ctx *gqlCtx) []error {
	errs := []error{}
	anonymous := []*ast.Operation{}
	operations := map[string]*ast.Operation{}
	for i := range ctx.doc.Operations {
		if ctx.doc.Operations[i].Name == "" {
			anonymous = append(anonymous, ctx.doc.Operations[i])
		}
		if _, ok := operations[ctx.doc.Operations[i].Name]; ok {
			errs = append(errs, fmt.Errorf("operation name must be unique"))
			continue
		}

		if ctx.doc.Operations[i].OperationType == ast.Subscription {
			// FIXME: provide CollectFields function
			ofg := v.e.CollectFields(ctx, v.schema.Subsciption, ctx.doc.Operations[i].SelectionSet, map[string]*ast.FragmentSpread{})
			if ofg.Len() != 1 {
				errs = append(errs, fmt.Errorf("subscription operation must have exactly one entry"))
			}
		}

		switch ctx.doc.Operations[i].OperationType {
		case ast.Query:
			vErrs := v.ValidateSelectionSet(ctx, v.schema.Query, ctx.doc.Operations[i].SelectionSet)
			errs = append(errs, vErrs...)
		case ast.Mutation:
			vErrs := v.ValidateSelectionSet(ctx, v.schema.Mutation, ctx.doc.Operations[i].SelectionSet)
			errs = append(errs, vErrs...)
		case ast.Subscription:
			vErrs := v.ValidateSelectionSet(ctx, v.schema.Mutation, ctx.doc.Operations[i].SelectionSet)
			errs = append(errs, vErrs...)
		}

		operations[ctx.doc.Operations[i].Name] = ctx.doc.Operations[i]
	}
	if len(operations) > 1 && len(anonymous) != 0 {
		errs = append(errs, fmt.Errorf("only one operation is allowed if there's an anonymous operation"))
	}
	return errs
}

func (v *validator) ValidateSelectionSet(ctx *gqlCtx, o *Object, set []ast.Selection) []error {
	for _, selection := range set {
		switch selection.Kind() {
		case ast.FieldSelectionKind:
			astField := selection.(*ast.Field)
			fieldType, err := o.Fields.Get(astField.Name)
			if err != nil {
				return []error{fmt.Errorf("field '%s' is not defined on type '%s'", astField.Name, fieldType.Name)}
			}
			errs := v.ValidateField(ctx, fieldType, astField)
			if len(errs) > 0 {
				return errs
			}
		}
	}
	return v.ValidateSelectionSetMerges(ctx, o, set)
}

func (v *validator) ValidateField(ctx *gqlCtx, f *Field, fAst *ast.Field) []error {
	if errs := v.ValidateArguments(ctx, f.Arguments, fAst.Arguments); len(errs) > 0 {
		return errs
	}
	// TODO: ValidateDirectives
	if ft := getFinalType(f.Type); ft.Kind() == ObjectTypeDefinition {
		// TODO: validate selectionSets on Interfaces and Unions
		errs := v.ValidateLeafFieldSelections(ctx, f)
		if len(errs) > 0 {
			return errs
		}

		errs = v.ValidateSelectionSet(ctx, ft.(*Object), fAst.SelectionSet)
		if len(errs) > 0 {
			return errs
		}
		errs = v.ValidateSelectionSetMerges(ctx, ft.(*Object), fAst.SelectionSet)
		if len(errs) > 0 {
			return errs
		}
	}
	return nil
}

func (v *validator) ValidateArguments(ctx *gqlCtx, args *Arguments, argsAst []*ast.Argument) []error {
	argNames := map[string]bool{}
	for _, argAst := range argsAst {
		if arg, ok := args.Get(argAst.Name); ok {
			if _, ok := argNames[argAst.Name]; ok {
				return []error{fmt.Errorf("argument '%s' is defined multiple times", argAst.Name)}
			}
			argNames[argAst.Name] = true

			if err := v.ValidateValue(argAst.Value, arg.Type, false); err != nil {
				return []error{err}
			}
		} else {
			return []error{fmt.Errorf("field '%s' does not esits", argAst.Name)}
		}
	}

	for _, arg := range args.Slice() {
		if _, ok := argNames[arg.Name]; !ok && arg.Type.Kind() == NonNullTypeDefinition {
			return []error{fmt.Errorf("required field '%s' is missing", arg.Name)}
		}
	}
	return nil
}

func (v *validator) ValidateValue(val ast.Value, t Type, nnParent bool) error {
	switch {
	case t.Kind() == NonNullTypeDefinition && val.Kind() != ast.NullValueKind:
		return v.ValidateValue(val.GetValue().(ast.Value), t.Unwrap(), true)
	case t.Kind() == ScalarTypeDefinition:
		s := t.(*Scalar)
		if val.Kind() == ast.NullValueKind && !nnParent {
			return nil
		} else if val.Kind() == ast.FloatValueKind || val.Kind() == ast.FloatValueKind || val.Kind() == ast.BooleanValueKind || val.Kind() == ast.StringValueKind {
			if _, err := s.InputCoercion(val.GetValue().(string)); err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("value is not valid")
	case t.Kind() == EnumTypeDefinition:
		e := t.(*Enum)
		if _, err := e.InputCoercion(val.GetValue().(string)); err != nil {
			return err
		}
		return nil
	case t.Kind() == ListTypeDefinition && val.Kind() == ast.ListValueKind:
		for _, sv := range val.GetValue().([]ast.Value) {
			if err := v.ValidateValue(sv, t.Unwrap(), false); err != nil {
				return err
			}
		}
		return nil
	case t.Kind() == InputObjectTypeDefinition && val.Kind() == ast.ObjectValueKind:
		obj := t.(*InputObject)
		fields := map[string]bool{}
		for _, fval := range val.GetValue().([]*ast.ObjectFieldValue) {
			if _, ok := fields[fval.Name]; ok {
				return fmt.Errorf("InputField uniqueness failed")
			}
			fields[fval.Name] = true

			inpField, err := obj.Fields.Get(fval.Name)
			if err != nil {
				return err
			}
			if err = v.ValidateValue(fval.Value, inpField.Type, false); err != nil {
				return err
			}
		}

		for _, f := range obj.Fields.Slice() {
			if _, ok := fields[f.Name]; !ok && f.Type.Kind() == NonNullTypeDefinition {
				return fmt.Errorf("required field is missing")
			}
		}
		return nil
	default:
		return fmt.Errorf("failed")
	}
}

func (v *validator) ValidateLeafFieldSelections(ctx *gqlCtx, f *Field) []error {
	// TODO: implement ValidateLeafFieldSelections
	return nil
}

func (v *validator) ValidateSelectionSetMerges(ctx *gqlCtx, o *Object, set []ast.Selection) []error {
	// TODO: implement ValidateSelectionSetMerges
	return nil
}

func (v *validator) ValidateFieldsInSetCanMerge(ctx *gqlCtx, o *Object, set []ast.Selection) []error {
	ofg := v.e.CollectFields(ctx, o, set, map[string]*ast.FragmentSpread{})
	for fieldName, fieldsForName := range ofg.Fields {
		fieldA := fieldsForName[0]
		for i := 1; i < len(fieldsForName); i++ {
			if !v.ValidateSameResponseShape(ctx, o, fieldA, fieldsForName[i]) {
				return []error{fmt.Errorf("fields '%s' can NOT be merged", fieldName)}
			}
		}
	}
	return nil
}

func (v *validator) ValidateSameResponseShape(ctx *gqlCtx, o *Object, fieldA *ast.Field, fieldB *ast.Field) bool {
	fieldDefA, err := o.Fields.Get(fieldA.Name)
	if err != nil {
		return false
	}
	fieldDefB, err := o.Fields.Get(fieldB.Name)
	if err != nil {
		return false
	}

	typeA, typeB, res, ok := validateNonNullAndList(fieldDefA.Type, fieldDefB.Type)
	if ok {
		return res
	}

	if typeA.Kind() == ScalarTypeDefinition || typeB.Kind() == ScalarTypeDefinition || typeA.Kind() == EnumTypeDefinition || typeB.Kind() == EnumTypeDefinition {
		if typeA == typeB {
			return true
		}
		return false
	}

	if isCompositeType(typeA) || isCompositeType(typeB) {
		return false
	}

	return true
}

func validateNonNullAndList(typeA Type, typeB Type) (Type, Type, bool, bool) {
	if typeA.Kind() == NonNullTypeDefinition || typeB.Kind() == NonNullTypeDefinition {
		if typeA.Kind() != NonNullTypeDefinition || typeB.Kind() != NonNullTypeDefinition {
			return typeA, typeB, false, true
		}
		typeA = typeA.Unwrap()
		typeB = typeB.Unwrap()

	}
	if typeA.Kind() == ListTypeDefinition || typeB.Kind() == ListTypeDefinition {
		if typeA.Kind() != ListTypeDefinition || typeB.Kind() != ListTypeDefinition {
			return typeA, typeB, false, true
		}
		typeA = typeA.Unwrap()
		typeB = typeB.Unwrap()
		return validateNonNullAndList(typeA, typeB)
	}
	return typeA, typeB, false, false
}

func isCompositeType(t Type) bool {
	if t.Kind() == UnionTypeDefinition || t.Kind() == InterfaceTypeDefinition || t.Kind() == ObjectTypeDefinition {
		return true
	}
	return false
}
