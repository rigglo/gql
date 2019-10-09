package gql

import (
	"fmt"
	"github.com/rigglo/gql/language/ast"
)

type Validator struct {
	schema *Schema
}

func (v *Validator) ValidateRequest(ctx *execCtx) []error {
	errs := v.ValidateOperations(ctx)
	return errs
}

func (v *Validator) ValidateOperations(ctx *execCtx) []error {
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
			/*
				ofg := v.CollectFields(ctx, s.Subsciption, ctx.doc.Operations[i].SelectionSet, map[string]*ast.FragmentSpread{})
				if ofg.Len() != 1 {
					errs = append(errs, fmt.Errorf("subscription operation must have exactly one entry"))
				}
			*/
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

func (v *Validator) ValidateSelectionSet(ctx *execCtx, o *Object, set []ast.Selection) []error {
	for _, selection := range set {
		switch selection.Kind() {
		case ast.FieldSelectionKind:
			astField := selection.(*ast.Field)
			fieldType, err := o.Fields.Get(astField.Name)
			if err != nil {
				return []error{fmt.Errorf("field '%s' is not defined on type '%s'", astField.Name, fieldType.Name)}
			}
			// TODO: ValidateArguments
			// TODO: ValidateDirectives
			v.ValidateLeafFieldSelections(ctx, fieldType)
		}
	}
	return v.ValidateSelectionSetMerges(ctx, o, set)
}

func (v *Validator) ValidateLeafFieldSelections(ctx *execCtx, f *Field) {

}

func (v *Validator) ValidateSelectionSetMerges(ctx *execCtx, o *Object, set []ast.Selection) []error {

	return nil
}

func (v *Validator) ValidateFieldsInSetCanMerge(ctx *execCtx, o *Object, set []ast.Selection) []error {
	/*
		ofg := v.CollectFields(ctx, o, set, map[string]*ast.FragmentSpread{})
		for fieldName, fieldsForName := range ofg.Fields {
			fieldA := fieldsForName[0]
			for i := 1; i < len(fieldsForName); i++ {
				if !v.ValidateSameResponseShape(ctx, o, fieldA, fieldsForName[i]) {
					return []error{fmt.Errorf("fields '%s' can NOT be merged", fieldName)}
				}
			}
		}
	*/
	return nil
}

func (v *Validator) ValidateSameResponseShape(ctx *execCtx, o *Object, fieldA *ast.Field, fieldB *ast.Field) bool {
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

/*
func (v *Validator) mergeSets(ctx *execCtx, o *Object, fieldA *ast.Field, fieldB *ast.Field) []ast.Selection {
	setA := s.CollectFields(ctx, o, fieldA.SelectionSet, map[string]*ast.FragmentSpread{})
	for _, sA := range setA {
		ofg.A
	}
}
*/
