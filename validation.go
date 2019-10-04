package gql

import (
	"fmt"
	"github.com/rigglo/gql/utils/ast"
)

func (s *Schema) ValidateRequest(ctx *execCtx) []error {
	errs := s.ValidateOperations(ctx)
	return errs
}

func (s *Schema) ValidateOperations(ctx *execCtx) []error {
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
			ofg := s.CollectFields(ctx, s.Subsciption, ctx.doc.Operations[i].SelectionSet, map[string]*ast.FragmentSpread{})
			if ofg.Len() != 1 {
				errs = append(errs, fmt.Errorf("subscription operation must have exactly one entry"))
			}
		}

		switch ctx.doc.Operations[i].OperationType {
		case ast.Query:
			vErrs := s.ValidateSelectionSet(ctx, s.Query, ctx.doc.Operations[i].SelectionSet)
			errs = append(errs, vErrs...)
		case ast.Mutation:
			vErrs := s.ValidateSelectionSet(ctx, s.Mutation, ctx.doc.Operations[i].SelectionSet)
			errs = append(errs, vErrs...)
		case ast.Subscription:
			vErrs := s.ValidateSelectionSet(ctx, s.Mutation, ctx.doc.Operations[i].SelectionSet)
			errs = append(errs, vErrs...)
		}

		operations[ctx.doc.Operations[i].Name] = ctx.doc.Operations[i]
	}
	if len(operations) > 1 && len(anonymous) != 0 {
		errs = append(errs, fmt.Errorf("only one operation is allowed if there's an anonymous operation"))
	}
	return errs
}

func (s *Schema) ValidateSelectionSet(ctx *execCtx, o *Object, set []ast.Selection) []error {
	for _, selection := range set {
		switch selection.Kind() {
		case ast.FieldSelectionKind:
			astField := selection.(*ast.Field)
			fieldType, err := o.Fields.Get(astField.Name)
			if err != nil {
				return []error{err}
			}
			// TODO: ValidateArguments
			// TODO: ValidateDirectives
			s.ValidateLeafFieldSelections(ctx, fieldType)
		}
	}
	return nil
}

func (s *Schema) ValidateLeafFieldSelections(ctx *execCtx, f *Field) {

}
