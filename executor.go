package gql

import (
	"context"
	"fmt"
	"github.com/rigglo/gql/internal"
	"github.com/rigglo/gql/language/ast"
	"github.com/rigglo/gql/language/parser"
	"github.com/rigglo/gql/pkg/ordered"
	"github.com/rigglo/gql/schema"
	"log"
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
	types         map[string]schema.Type
	vars          map[string]interface{}
}

type executor struct {
	schema *schema.Schema
	types  map[string]schema.Type
}

func (e *executor) ExecuteQuery(ctx *execCtx, query *ast.Operation, initVal interface{}) *Result {
	data, err := e.ExecuteSelectionSet(ctx, e.schema.Query, query.SelectionSet, initVal)
	return &Result{
		Data:   data,
		Errors: err,
	}
}

func (e *executor) ExecuteSelectionSet(ctx *execCtx, o *schema.Object, set []ast.Selection, val interface{}) (*ordered.Map, []error) {
	ofg := e.CollectFields(ctx, o, set, map[string]*ast.FragmentSpread{})
	res := ordered.NewMap()
	errs := []error{}
	iter := ofg.Iter()
	for iter.Next() {
		// TODO: FieldsInSetCanMerge
		key, fields := iter.Value()
		fName := fields[0].Name
		f, err := o.Fields.Get(fName)
		if err != nil {
			errs = append(errs, err)
		}
		fVal, fErrs := e.ExecuteField(ctx, o, val, fields, f.Type)
		errs = append(errs, fErrs...)
		res.Append(key, fVal)
	}

	return res, nil
}

func (e *executor) ExecuteField(ctx *execCtx, o *schema.Object, val interface{}, fields ast.Fields, ftype schema.Type) (interface{}, []error) {
	field := fields[0]
	errs := []error{}
	fieldName := field.Name
	args, cErrs := e.CoerceArgumentValues(ctx, o, field)
	errs = append(errs, cErrs...)
	resVal, err := e.ResolveFieldValue(o, val, fieldName, args)
	errs = append(errs, err)
	fT, err := o.Fields.Get(fieldName)
	errs = append(errs, err)
	cVal, cErrs := e.CompleteValue(ctx, fT.Type, fields, resVal)
	errs = append(errs, cErrs...)
	return cVal, errs
}

func (e *executor) CoerceArgumentValues(ctx *execCtx, o *schema.Object, field *ast.Field) (map[string]interface{}, []error) {
	vals := map[string]interface{}{}
	/* fieldDef, _ := o.Fields.Get(field.Name)
	argVals := field.Arguments
	argDefs := fieldDef.Arguments
	for _, argDef := range argDefs {
		argName := argDef.Name
		argType := argDef.Type
		defaultVal := argDef.DefaultValue
		hasVal := false
		for _, arg := range field.Arguments {
			if arg.Name == argName {
				hasVal = true
				break
			}
		}

	} */
	return vals, nil
}

func (e *executor) ResolveFieldValue(o *schema.Object, val interface{}, fieldName string, args map[string]interface{}) (interface{}, error) {
	field, _ := o.Fields.Get(fieldName)
	// TODO: run resolver in VM
	return field.Resolver(context.Background(), args, val)
}

func (e *executor) CompleteValue(ctx *execCtx, fT schema.Type, fields ast.Fields, cVal interface{}) (interface{}, []error) {
	errs := []error{}
	switch {
	case fT.Kind() == schema.NonNullTypeDefinition:
		innerType := fT.Unwrap()
		cRes, cErrs := e.CompleteValue(ctx, innerType, fields, cVal)
		errs = append(errs, cErrs...)
		if cRes == nil {
			errs = append(errs, fmt.Errorf("null value error on NonNull field"))
		}
		return cRes, errs
	case cVal == nil:
		return nil, errs
	case fT.Kind() == schema.ListTypeDefinition:
		out := []interface{}{}
		if reflect.TypeOf(cVal).Kind() == reflect.Slice {
			for i := 0; i < reflect.ValueOf(cVal).Len(); i++ {
				cRes, cErrs := e.CompleteValue(ctx, fT.Unwrap(), fields, reflect.ValueOf(cVal).Index(i).Interface())
				errs = append(errs, cErrs...)
				out = append(out, cRes)
			}
		}
		return out, errs
	case fT.Kind() == schema.ScalarTypeDefinition:
		cRes, err := fT.(*schema.Scalar).OutputCoercion(cVal)
		if err != nil {
			errs = append(errs, err)
			return nil, errs
		}
		return cRes, errs
	case fT.Kind() == schema.EnumTypeDefinition:
		log.Println(cVal)
		cRes, err := fT.(*schema.Enum).OutputCoercion(cVal)
		log.Println(string(cRes))
		if err != nil {
			errs = append(errs, err)
			return nil, errs
		}
		return cRes, errs
	case fT.Kind() == schema.ObjectTypeDefinition: // TODO: check Interface and Union too
		out, errs := e.ExecuteSelectionSet(ctx, fT.(*schema.Object), e.MergeSelectionSets(ctx, fields), cVal)
		return out, errs
	}
	return nil, []error{fmt.Errorf("Schema.CompleteValue - END - you should not get here.. ")}
}

func (e *executor) MergeSelectionSets(ctx *execCtx, fields ast.Fields) []ast.Selection {
	set := []ast.Selection{}
	for _, f := range fields {
		set = append(set, f.SelectionSet...)
	}
	return set
}

func (e *executor) CollectFields(ctx *execCtx, o *schema.Object, set []ast.Selection, vf map[string]*ast.FragmentSpread) *internal.OrderedFieldGroups {
	groupedFields := internal.NewOrderedFieldGroups()
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
				fragmentType, ok := e.types[fragment.TypeCondition]
				if !ok {
					// RAISE FIELD ERROR
					continue
				}
				if !e.DoesFragmentTypeApply(o, fragmentType) {
					// RAISE FIELD ERROR
					continue
				}
			} else {
				// RAISE FIELD ERROR
				continue
			}
			res := e.CollectFields(ctx, o, fragment.SelectionSet, vf)
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
				fragmentType, ok := e.types[fragment.TypeCondition]
				if !ok {
					// RAISE FIELD ERROR
					continue
				}
				if !e.DoesFragmentTypeApply(o, fragmentType) {
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

func (e *executor) DoesFragmentTypeApply(o *schema.Object, fragmentType schema.Type) bool {
	return true
}

func (e *executor) Execute(ctx context.Context, query string, operationName string, vars map[string]interface{}) *Result {
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
		types:         map[string]schema.Type{},
	}

	/*
		if errs := e.ValidateRequest(eCtx); len(errs) > 0 {
			return &Result{
				Errors: append(errs, fmt.Errorf("validation error")),
			}
		}*/

	op, err := e.GetOperation(doc, operationName)
	if err != nil {
		return &Result{
			Errors: []error{err},
		}
	}

	switch op.OperationType {
	case ast.Query:
		return e.ExecuteQuery(eCtx, op, nil)
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

func (s *executor) GetOperation(doc *ast.Document, operationName string) (*ast.Operation, error) {
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
