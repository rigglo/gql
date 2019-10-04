package gql

import (
	"context"
	"fmt"
	"github.com/rigglo/gql/utils/ast"
	"github.com/rigglo/gql/utils/parser"
	"reflect"
)

type Result struct {
	Data   interface{} `json:"data"`
	Errors []error     `json:"errors"`
}

type execCtx struct {
	query         string
	Context       context.Context
	doc           *ast.Document
	operationName string
	vars          map[string]interface{}
}

func (schema *Schema) ExecuteQuery(ctx *execCtx, query *ast.Operation, initVal interface{}) *Result {
	data, err := schema.ExecuteSelectionSet(ctx, schema.Query, query.SelectionSet, initVal)
	return &Result{
		Data:   data,
		Errors: err,
	}
}

func (schema *Schema) ExecuteSelectionSet(ctx *execCtx, o *Object, set []ast.Selection, val interface{}) (*OrderedMap, []error) {
	ofg := schema.CollectFields(ctx, o, set, map[string]*ast.FragmentSpread{})
	res := NewOrderedMap()
	errs := []error{}
	iter := ofg.Iter()
	for iter.Next() {
		key, fields := iter.Value()
		fName := fields[0].Name
		f, err := o.Fields.Get(fName)
		if err != nil {
			errs = append(errs, err)
		}
		fVal, fErrs := schema.ExecuteField(ctx, o, val, fields, f.Type)
		errs = append(errs, fErrs...)
		res.Append(key, fVal)
	}

	return res, nil
}

func (schema *Schema) ExecuteField(ctx *execCtx, o *Object, val interface{}, fields ast.Fields, ftype Type) (interface{}, []error) {
	field := fields[0]
	errs := []error{}
	fieldName := field.Name
	args, cErrs := schema.CoerceArgumentValues(ctx, o, field)
	errs = append(errs, cErrs...)
	resVal, err := schema.ResolveFieldValue(o, val, fieldName, args)
	errs = append(errs, err)
	fT, err := o.Fields.Get(fieldName)
	errs = append(errs, err)
	cVal, cErrs := schema.CompleteValue(ctx, fT.Type, fields, resVal)
	errs = append(errs, cErrs...)
	return cVal, errs
}

func (schema *Schema) CoerceArgumentValues(ctx *execCtx, o *Object, field *ast.Field) (map[string]interface{}, []error) {
	return nil, nil
}

func (schema *Schema) ResolveFieldValue(o *Object, val interface{}, fieldName string, args map[string]interface{}) (interface{}, error) {
	field, _ := o.Fields.Get(fieldName)
	// TODO: run resolver in VM
	return field.Resolver(context.Background(), args, val)
}

func (schema *Schema) CompleteValue(ctx *execCtx, fT Type, fields ast.Fields, cVal interface{}) (interface{}, []error) {
	errs := []error{}
	switch {
	case fT.Kind() == NonNullTypeDefinition:
		innerType := fT.Unwrap()
		cRes, cErrs := schema.CompleteValue(ctx, innerType, fields, cVal)
		errs = append(errs, cErrs...)
		if cRes == nil {
			errs = append(errs, fmt.Errorf("null value error on NonNull field"))
		}
		return cRes, errs
	case cVal == nil:
		return nil, errs
	case fT.Kind() == ListTypeDefinition:
		out := []interface{}{}
		if reflect.TypeOf(cVal).Kind() == reflect.Slice {
			for i := 0; i < reflect.ValueOf(cVal).Len(); i++ {
				cRes, cErrs := schema.CompleteValue(ctx, fT.Unwrap(), fields, reflect.ValueOf(cVal).Index(i).Interface())
				errs = append(errs, cErrs...)
				out = append(out, cRes)
			}
		}
		return out, errs
	case fT.Kind() == ScalarTypeDefinition: // TODO: also check Enum
		cRes, err := fT.(*Scalar).OutputCoercion(cVal)
		if err != nil {
			errs = append(errs, err)
			return nil, errs
		}
		return cRes, errs
	case fT.Kind() == ObjectTypeDefinition: // TODO: check Interface and Union too
		out, errs := schema.ExecuteSelectionSet(ctx, fT.(*Object), schema.MergeSelectionSets(ctx, fields), cVal)
		return out, errs
	}
	return nil, []error{fmt.Errorf("Schema.CompleteValue - END - you should not get here.. ")}
}

func (s *Schema) MergeSelectionSets(ctx *execCtx, fields ast.Fields) []ast.Selection {
	set := []ast.Selection{}
	for _, f := range fields {
		set = append(set, f.SelectionSet...)
	}
	return set
}

func (schema *Schema) CollectFields(ctx *execCtx, o *Object, set []ast.Selection, vf map[string]*ast.FragmentSpread) *orderedFieldGroups {
	groupedFields := newOrderedFieldGroups()
	for s := range set {
		// TODO: Check skip and include directives
		if f, ok := set[s].(*ast.Field); ok {
			if _, err := o.Fields.Get(f.Name); err != nil {
				continue
			} else {
				groupedFields.Append(f.Alias, f)
			}
		} else if f, ok := set[s].(*ast.FragmentSpread); ok {
			if _, ok := vf[f.Name]; ok {
				continue
			}
			vf[f.Name] = f

			fragment, ok := ctx.doc.Fragments[f.Name]
			if ok {
				fragmentType, ok := schema.types[fragment.TypeCondition]
				if !ok {
					// RAISE FIELD ERROR
					continue
				}
				if !schema.DoesFragmentTypeApply(o, fragmentType) {
					// RAISE FIELD ERROR
					continue
				}
			} else {
				// RAISE FIELD ERROR
				continue
			}
			res := schema.CollectFields(ctx, o, fragment.SelectionSet, vf)
			rIter := res.Iter()
			for rIter.Next() {
				alias, rfs := rIter.Value()
				for rf := range rfs {
					groupedFields.Append(alias, rfs[rf])
				}
			}
		} else if f, ok := set[s].(*ast.InlineFragment); ok {
			fragment, ok := ctx.doc.Fragments[f.TypeCondition]
			if ok {
				fragmentType, ok := schema.types[fragment.TypeCondition]
				if !ok {
					// RAISE FIELD ERROR
					continue
				}
				if !schema.DoesFragmentTypeApply(o, fragmentType) {
					// RAISE FIELD ERROR
					continue
				}
			} else {
				// RAISE FIELD ERROR
				continue
			}
		}
	}
	return groupedFields
}

func (s *Schema) DoesFragmentTypeApply(o *Object, fragmentType Type) bool {
	return true
}

func (s *Schema) Execute(ctx context.Context, query string, operationName string, vars map[string]interface{}) *Result {
	_, doc, err := parser.Parse(query)
	if err != nil {
		return &Result{
			Errors: []error{err},
		}
	}

	eCtx := &execCtx{
		Context:       ctx,
		doc:           doc,
		operationName: operationName,
		query:         query,
		vars:          vars,
	}

	if errs := s.ValidateRequest(eCtx); len(errs) > 0 {
		return &Result{
			Errors: errs,
		}
	}

	op, err := s.GetOperation(doc, operationName)
	if err != nil {
		return &Result{
			Errors: []error{err},
		}
	}

	switch op.OperationType {
	case ast.Query:
		return s.ExecuteQuery(eCtx, op, nil)
	case ast.Mutation:
		// TODO: implement ExecuteMutation
		break
	case ast.Subscription:
		// TODO: implement ExecuteSubscription
		break
	}
	return &Result{
		Errors: []error{fmt.Errorf("invalid operation type")},
	}
}

func (s *Schema) GetOperation(doc *ast.Document, operationName string) (*ast.Operation, error) {
	if operationName == "" && len(doc.Operations) != 1 {
		return nil, fmt.Errorf("operationName is required")
	}
	for i := range doc.Operations {
		if doc.Operations[i].Name == operationName {
			return doc.Operations[i], nil
		}
	}
	return nil, fmt.Errorf("invalid operationName: operation '%s' not found", operationName)
}
