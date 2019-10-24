package gql

import (
	"context"
	"fmt"
	"github.com/rigglo/gql/internal"
	"github.com/rigglo/gql/language/ast"
	"github.com/rigglo/gql/pkg/ordered"
	"github.com/rigglo/gql/pkg/vm"
	"reflect"
)

type Result struct {
	Data   interface{} `json:"data"`
	Errors []error     `json:"errors"`
}

type gqlCtx struct {
	query         string
	Context       context.Context
	doc           *ast.Document
	operationName string
	types         map[string]Type
	vars          map[string]interface{}
}

type executor struct {
	schema *Schema
	types  map[string]Type
}

func (e *executor) ExecuteQuery(ctx *gqlCtx, query *ast.Operation, initVal interface{}) *Result {
	data, err := e.ExecuteSelectionSet(ctx, e.schema.Query, query.SelectionSet, initVal)
	return &Result{
		Data:   data,
		Errors: err,
	}
}

func (e *executor) ExecuteMutation(ctx *gqlCtx, mut *ast.Operation, initVal interface{}) *Result {
	data, err := e.ExecuteSelectionSet(ctx, e.schema.Mutation, mut.SelectionSet, initVal)
	return &Result{
		Data:   data,
		Errors: err,
	}
}

func (e *executor) ExecuteSelectionSet(ctx *gqlCtx, o *Object, set []ast.Selection, val interface{}) (*ordered.Map, []error) {
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

	return res, errs
}

func (e *executor) ExecuteField(ctx *gqlCtx, o *Object, val interface{}, fields ast.Fields, ftype Type) (interface{}, []error) {
	field := fields[0]
	errs := []error{}
	fieldName := field.Name
	args, cErrs := e.CoerceArgumentValues(ctx, o, field)
	if cErrs != nil {
		errs = append(errs, cErrs...)
	}
	errs = append(errs, cErrs...)
	resVal, err := e.ResolveFieldValue(ctx, o, val, fieldName, args)
	if err != nil {
		errs = append(errs, err)
	}
	fT, _ := o.Fields.Get(fieldName)
	cVal, cErrs := e.CompleteValue(ctx, fT.Type, fields, resVal)
	if cErrs != nil {
		errs = append(errs, cErrs...)
	}
	return cVal, errs
}

func (e *executor) CoerceArgumentValues(ctx *gqlCtx, o *Object, field *ast.Field) (map[string]interface{}, []error) {
	fieldDef, _ := o.Fields.Get(field.Name)
	coercedVals := map[string]interface{}{}
	if fieldDef.Arguments == nil {
		return coercedVals, nil
	}

	for _, argDef := range fieldDef.Arguments.Slice() {
		defaultVal := argDef.DefaultValue
		hasVal := false
		var value interface{}
		var argVal ast.Value
		for _, arg := range field.Arguments {
			if arg.Name == argDef.Name {
				hasVal = true
				argVal = arg.Value
				break
			}
		}
		if argVal.Kind() == ast.VariableValueKind {
			varName := argVal.GetValue().(string)
			value, hasVal = ctx.vars[varName]
		} else {
			value = argVal.GetValue()
		}
		if !hasVal {
			coercedVals[argDef.Name] = defaultVal
		} else if argDef.Type.Kind() == NonNullTypeDefinition && (!hasVal || argVal.Kind() == ast.NullValueKind) {
			return nil, []error{fmt.Errorf("null value on Non-Null type")}
		} else if hasVal {
			if value == nil {
				coercedVals[argDef.Name] = value
			} else if argVal.Kind() == ast.VariableValueKind {
				coercedVals[argDef.Name] = value
			} else {
				value, err := coerceValue(value, argDef.Type)
				if err != nil {
					return nil, []error{err}
				}
				coercedVals[argDef.Name] = value
			}
		}
	}
	return coercedVals, nil
}

func coerceValue(val interface{}, t Type) (interface{}, error) {
	switch {
	case t.Kind() == NonNullTypeDefinition:
		res, err := coerceValue(val, t.Unwrap())
		return res, err
	case t.Kind() == ScalarTypeDefinition:
		s := t.(*Scalar)
		if reflect.ValueOf(val).Kind() == reflect.String {
			return s.InputCoercion(val.(string))
		} else if v, ok := val.(ast.Value); ok {
			return s.InputCoercion(v.GetValue().(string))
		}
		return nil, fmt.Errorf("invalid value on a Scalar type")
	case t.Kind() == EnumTypeDefinition:
		e := t.(*Enum)
		if reflect.ValueOf(val).Kind() == reflect.String {
			return e.InputCoercion(val.(string))
		} else if v, ok := val.(ast.Value); ok {
			return e.InputCoercion(v.GetValue().(string))
		}
		return nil, fmt.Errorf("invalid value on an Emum type")
	case t.Kind() == ListTypeDefinition:
		vals := []interface{}{}
		if reflect.TypeOf(val).Kind() == reflect.Slice {
			for i := 0; i < reflect.ValueOf(val).Len(); i++ {
				cVal, err := coerceValue(reflect.ValueOf(val).Index(i).Interface().(ast.Value).GetValue(), t.Unwrap())
				if err != nil {
					return nil, err
				}
				vals = append(vals, cVal)
			}
			return vals, nil
		} else if val == nil {
			return nil, nil
		}
		return nil, fmt.Errorf("value is not coercible to List")
	case t.Kind() == InputObjectTypeDefinition:
		vals := map[string]interface{}{}
		o := t.(*InputObject)
		if reflect.TypeOf(val).Kind() == reflect.Slice {
			for i := 0; i < reflect.ValueOf(val).Len(); i++ {
				field, ok := reflect.ValueOf(val).Index(i).Interface().(*ast.ObjectFieldValue)
				if !ok {
					return nil, fmt.Errorf("value is not valid for InputObject type")
				}
				ft, _ := o.Fields.Get(field.Name)
				fieldVal, err := coerceValue(field.GetValue(), ft.Type)
				if err != nil {
					return nil, err
				}
				vals[field.Name] = fieldVal
			}
			return vals, nil
		} else if val == nil {
			return nil, nil
		}
		return nil, fmt.Errorf("value is not coercible to InputObject")
	}
	return nil, nil
}

func (e *executor) ResolveFieldValue(ctx *gqlCtx, o *Object, val interface{}, fieldName string, args map[string]interface{}) (interface{}, error) {
	field, _ := o.Fields.Get(fieldName)
	return vm.Run(func() (interface{}, error) {
		return field.Resolver(ctx.Context, args, val)
	})
}

func (e *executor) CompleteValue(ctx *gqlCtx, fT Type, fields ast.Fields, cVal interface{}) (interface{}, []error) {
	errs := []error{}
	switch {
	case fT.Kind() == NonNullTypeDefinition:
		innerType := fT.Unwrap()
		cRes, cErrs := e.CompleteValue(ctx, innerType, fields, cVal)
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
				cRes, cErrs := e.CompleteValue(ctx, fT.Unwrap(), fields, reflect.ValueOf(cVal).Index(i).Interface())
				errs = append(errs, cErrs...)
				out = append(out, cRes)
			}
		}
		return out, errs
	case fT.Kind() == ScalarTypeDefinition:
		cRes, err := fT.(*Scalar).OutputCoercion(cVal)
		if err != nil {
			errs = append(errs, err)
			return nil, errs
		}
		return cRes, errs
	case fT.Kind() == EnumTypeDefinition:
		cRes, err := fT.(*Enum).OutputCoercion(cVal)
		if err != nil {
			errs = append(errs, err)
			return nil, errs
		}
		return cRes, errs
	case fT.Kind() == ObjectTypeDefinition: // TODO: check Union
		out, errs := e.ExecuteSelectionSet(ctx, fT.(*Object), e.MergeSelectionSets(ctx, fields), cVal)
		return out, errs
	case fT.Kind() == InterfaceTypeDefinition:
		i := fT.(*Interface)
		out, errs := e.ExecuteSelectionSet(ctx, i.ResolveType(ctx.Context, cVal), e.MergeSelectionSets(ctx, fields), cVal)
		return out, errs
	}
	return nil, []error{fmt.Errorf("Schema.CompleteValue - END - you should not get here.. ")}
}

func (e *executor) MergeSelectionSets(ctx *gqlCtx, fields ast.Fields) []ast.Selection {
	set := []ast.Selection{}
	for _, f := range fields {
		set = append(set, f.SelectionSet...)
	}
	return set
}

func (e *executor) CollectFields(ctx *gqlCtx, o *Object, set []ast.Selection, vf map[string]*ast.FragmentSpread) *internal.OrderedFieldGroups {
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
				/*fragmentType, ok := e.types[fragment.TypeCondition]
				if !ok {
					// RAISE FIELD ERROR
					continue
				}
				*/
				if !e.DoesFragmentTypeApply(o, fragment.TypeCondition) {
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
			if !e.DoesFragmentTypeApply(o, f.TypeCondition) {
				// RAISE FIELD ERROR
				continue
			}

			res := e.CollectFields(ctx, o, f.SelectionSet, vf)
			rIter := res.Iter()
			for rIter.Next() {
				alias, rfs := rIter.Value()
				for rf := range rfs {
					groupedFields.Append(alias, rfs[rf])
				}
			}
		}
	}
	return groupedFields
}

func (e *executor) DoesFragmentTypeApply(o *Object, fragmentType string) bool {
	// FIXME: check if this correct
	if o.Name == fragmentType {
		return true
	}
	for _, i := range o.Implements {
		if i.Name == fragmentType {
			return true
		}
	}
	return false
}

func (e *executor) Execute(eCtx *gqlCtx) *Result {
	op, err := e.GetOperation(eCtx.doc, eCtx.operationName)
	if err != nil {
		return &Result{
			Errors: []error{err},
		}
	}

	switch op.OperationType {
	case ast.Query:
		return e.ExecuteQuery(eCtx, op, nil)
	case ast.Mutation:
		return e.ExecuteMutation(eCtx, op, nil)
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
