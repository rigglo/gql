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

func (schema *Schema) ExecuteQuery(doc *ast.Document, query *ast.Operation, vars map[string]interface{}, initVal interface{}) *Result {
	data, err := schema.ExecuteSelectionSet(doc, schema.Query, query.SelectionSet, vars, initVal)
	return &Result{
		Data:   data,
		Errors: err,
	}
}

func (schema *Schema) ExecuteSelectionSet(doc *ast.Document, o *Object, set []ast.Selection, vars map[string]interface{}, val interface{}) (*OrderedMap, []error) {
	ofg := schema.CollectFields(doc, o, set, vars, map[string]*ast.FragmentSpread{})
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
		fVal, fErrs := schema.ExecuteField(doc, o, val, fields, f.Type, vars)
		errs = append(errs, fErrs...)
		res.Append(key, fVal)
	}

	return res, nil
}

func (schema *Schema) ExecuteField(doc *ast.Document, o *Object, val interface{}, fields ast.Fields, ftype Type, vars map[string]interface{}) (interface{}, []error) {
	field := fields[0]
	errs := []error{}
	fieldName := field.Name
	args, cErrs := schema.CoerceArgumentValues(o, field, vars)
	errs = append(errs, cErrs...)
	resVal, err := schema.ResolveFieldValue(o, val, fieldName, args)
	errs = append(errs, err)
	fT, err := o.Fields.Get(fieldName)
	errs = append(errs, err)
	cVal, cErrs := schema.CompleteValue(doc, fT.Type, fields, resVal, vars)
	errs = append(errs, cErrs...)
	return cVal, errs
}

func (schema *Schema) CoerceArgumentValues(o *Object, field *ast.Field, vars map[string]interface{}) (map[string]interface{}, []error) {
	return nil, nil
}

func (schema *Schema) ResolveFieldValue(o *Object, val interface{}, fieldName string, args map[string]interface{}) (interface{}, error) {
	field, _ := o.Fields.Get(fieldName)
	return field.Resolver(context.Background(), args, val)
}

func (schema *Schema) CompleteValue(doc *ast.Document, fT Type, fields ast.Fields, cVal interface{}, vars map[string]interface{}) (interface{}, []error) {
	errs := []error{}
	switch {
	case fT.Kind() == NonNullTypeDefinition:
		innerType := fT.Unwrap()
		cRes, cErrs := schema.CompleteValue(doc, innerType, fields, cVal, vars)
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
				cRes, cErrs := schema.CompleteValue(doc, fT.Unwrap(), fields, reflect.ValueOf(cVal).Index(i).Interface(), vars)
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
		out, errs := schema.ExecuteSelectionSet(doc, fT.(*Object), fields[0].SelectionSet, vars, cVal)
		return out, errs
	}
	return cVal, nil
}

func (schema *Schema) CollectFields(doc *ast.Document, o *Object, set []ast.Selection, vars map[string]interface{}, vf map[string]*ast.FragmentSpread) *orderedFieldGroups {
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

			fragment, ok := doc.Fragments[f.Name]
			if ok {
				if fragment.TypeCondition != o.Name {
					continue
				}
				// TODO: Check if fragment on type matches the object type
			} else {
				continue
			}
			res := schema.CollectFields(doc, o, fragment.SelectionSet, vars, vf)
			rIter := res.Iter()
			for rIter.Next() {
				alias, rfs := rIter.Value()
				for rf := range rfs {
					groupedFields.Append(alias, rfs[rf])
				}
			}
		} else if _, ok := set[s].(*ast.InlineFragment); ok {
			// log.Println("InlineFragment:", f)
			// TODO: add inline fragments
		}
	}
	return groupedFields
}

func (s *Schema) Resolve(ctx context.Context, query string) (*Result, error) {
	_, doc, err := parser.Parse(query)
	if err != nil {
		return nil, err
	}

	for _, op := range doc.Operations {
		if op.OperationType == ast.Query {
			// TODO: check if s.Query is nil
			return s.ExecuteQuery(doc, op, map[string]interface{}{}, nil), nil
		}
	}
	return nil, nil
}
