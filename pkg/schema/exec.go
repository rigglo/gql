package schema

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/rigglo/gql/pkg/language/ast"
	"github.com/rigglo/gql/pkg/language/parser"
)

type Result struct {
	ctx    context.Context
	Data   map[string]interface{} `json:"data"`
	Errors []*Error               `json:"errors,omitempty"`
	errMu  sync.Mutex
}

func (r *Result) addErr(err *Error) {
	r.errMu.Lock()
	defer r.errMu.Unlock()
	r.Errors = append(r.Errors, err)
}

type ExecuteParams struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"OperationName"`
}

func Execute(ctx context.Context, s Schema, p ExecuteParams) *Result {
	_, doc, err := parser.Parse(p.Query)
	if err != nil {
		return &Result{
			ctx:    ctx,
			Errors: Errors{NewError(ctx, err.Error(), nil)},
		}
	}
	ectx := newCtx(ctx, map[string]interface{}{
		keyQuery:         doc,
		keyRawQuery:      p.Query,
		keyOperationName: p.OperationName,
		keyVariables:     p.Variables,
		keySchema:        s,
	})
	return resolveOperation(ectx)
}

func resolveOperation(ctx *eCtx) *Result {
	oname := ctx.Get(keyOperationName).(string)
	doc := ctx.Get(keyQuery).(*ast.Document)
	if oname == "" && len(doc.Operations) != 1 {
		return &Result{
			ctx:    ctx,
			Errors: Errors{NewError(ctx, "operationName is requeired", nil)},
		}
	}
	for _, op := range doc.Operations {
		if op.Name == oname {
			ctx.Set(keyOperation, op)
			switch op.OperationType {
			case ast.Query:
				return executeQuery(ctx, op)
			case ast.Mutation:
				return executeMutation(ctx)
			case ast.Subscription:
				// TODO: implement ExecuteSubscription
				break
			}
		}
	}
	return &Result{
		ctx:    ctx,
		Errors: Errors{NewError(ctx, "execution failed", nil)},
	}
}

func executeQuery(ctx *eCtx, op *ast.Operation) *Result {
	schema := ctx.Get(keySchema).(Schema)
	rmap := executeSelectionSet(ctx, []interface{}{}, op.SelectionSet, schema.GetRootQuery().(ObjectType), nil)
	ctx.res.Data = rmap
	return ctx.res
}

func executeMutation(ctx *eCtx) *Result {
	return nil
}

func executeSelectionSet(ctx *eCtx, path []interface{}, ss []ast.Selection, ot ObjectType, ov interface{}) map[string]interface{} {
	gfields := collectFields(ctx, ss, nil)
	resMap := map[string]interface{}{}
	for rkey, fields := range gfields {
		fieldName := fields[0].Name
		fieldType := getFieldOfFields(fieldName, ot.GetFields()).GetType()
		rval := executeField(ctx, append(path, fields[0].Alias), ot, ov, fieldType, fields)
		resMap[rkey] = rval
	}
	return resMap
}

func collectFields(ctx *eCtx, ss []ast.Selection, vFrags []string) map[string]ast.Fields {
	if vFrags == nil {
		vFrags = []string{}
	}
	gfields := map[string]ast.Fields{}

	for _, sel := range ss {
		// TODO: directive checks (skip, include)
		switch sel.Kind() {
		case ast.FieldSelectionKind:
			{
				f := sel.(*ast.Field)
				if _, ok := gfields[f.Alias]; ok {
					gfields[f.Alias] = append(gfields[f.Alias], f)
				} else {
					gfields[f.Alias] = ast.Fields{f}
				}
			}
			// TODO: FragmentSpread
			// TODO: InlineFragment
		}
	}
	return gfields
}

func getFieldOfFields(fn string, fs []Field) Field {
	for _, f := range fs {
		if f.GetName() == fn {
			return f
		}
	}
	return nil
}

func getArgOfArgs(an string, as []*ast.Argument) (*ast.Argument, bool) {
	for _, a := range as {
		if a.Name == an {
			return a, true
		}
	}
	return nil, false
}

func executeField(ctx *eCtx, path []interface{}, ot ObjectType, ov interface{}, ft Type, fs ast.Fields) interface{} {
	f := fs[0]
	fn := f.Name
	args := coerceArgumentValues(ctx, path, ot, f)
	resVal := resolveFieldValue(ctx, path, f, ot, ov, fn, args)
	return completeValue(ctx, path, getFieldOfFields(fn, ot.GetFields()).GetType(), fs, resVal)
}

func coerceArgumentValues(ctx *eCtx, path []interface{}, ot ObjectType, f *ast.Field) map[string]interface{} {
	vars := ctx.Get(keyVariables).(map[string]interface{})
	coercedVals := map[string]interface{}{}
	fieldName := f.Name
	argDefs := getFieldOfFields(fieldName, ot.GetFields()).GetArguments()
	for _, argDef := range argDefs {
		argName := argDef.GetName()
		argType := argDef.GetType()
		defaultValue := argDef.GetDefaultValue()
		argVal, hasValue := getArgOfArgs(argName, f.Arguments)
		var value interface{}
		if argVal != nil {
			if argVal.Value.Kind() == ast.VariableValueKind {
				varName := argVal.Value.(*ast.VariableValue).Name
				value, hasValue = vars[varName]
			} else {
				value = argVal.Value.GetValue()
			}
		}
		if !hasValue && argDef.IsDefaultValueSet() {
			coercedVals[argName] = defaultValue
		} else if argType.GetKind() == NonNullKind && (!hasValue || value == nil) {
			ctx.res.addErr(&Error{
				Message: fmt.Sprintf("Argument '%s' is a Non-Null field, but got null value", argName),
				Path:    path,
				Locations: []*ErrorLocation{
					&ErrorLocation{
						Column: argVal.Location.Column,
						Line:   argVal.Location.Line,
					},
				},
			})
		} else if hasValue {
			if value == nil {
				coercedVals[argName] = value
			} else if argVal.Value.Kind() == ast.VariableValueKind {
				coercedVals[argName] = value
			} else {
				coercedVal, err := coerceValue(value, argType)
				if err != nil {
					ctx.res.addErr(&Error{
						Message: err.Error(),
						Path:    path,
						Locations: []*ErrorLocation{
							&ErrorLocation{
								Column: argVal.Location.Column,
								Line:   argVal.Location.Line,
							},
						},
					})
				} else {
					coercedVals[argName] = coercedVal
				}
			}
		}

	}
	return coercedVals
}

func coerceValue(val interface{}, t Type) (interface{}, error) {
	switch {
	case t.GetKind() == NonNullKind:
		return coerceValue(val, t.(NonNull).Unwrap())
	case t.GetKind() == ScalarKind:
		s := t.(ScalarType)
		if reflect.ValueOf(val).Kind() == reflect.String {
			return s.CoerceInput([]byte(val.(string)))
		} else if v, ok := val.(ast.Value); ok {
			return s.CoerceInput([]byte(v.GetValue().(string)))
		}
		return nil, fmt.Errorf("invalid value on a Scalar type")
	}
	return nil, errors.New("invalid value to coerce")
}

func resolveFieldValue(ctx *eCtx, path []interface{}, fast *ast.Field, ot ObjectType, ov interface{}, fn string, args map[string]interface{}) interface{} {
	f := getFieldOfFields(fn, ot.GetFields())
	v, err := f.Resolve(ctx, ov, args)
	if err != nil {
		ctx.res.addErr(&Error{
			Message: err.Error(),
			Path:    path,
			Locations: []*ErrorLocation{
				&ErrorLocation{
					Column: fast.Location.Column,
					Line:   fast.Location.Line,
				},
			},
		})
	}
	return v
}

func completeValue(ctx *eCtx, path []interface{}, ft Type, fs ast.Fields, result interface{}) interface{} {
	// TODO: complete value
	switch ft.GetKind() {
	case ScalarKind:
		res, err := ft.(ScalarType).CoerceResult(result)
		if err != nil {
			ctx.res.addErr(&Error{Message: err.Error()})
		}
		return res
	case ObjectKind:
		ot := ft.(ObjectType)
		subSel := fs[0].SelectionSet
		return executeSelectionSet(ctx, path, subSel, ot, result)
	case NonNullKind:
		return completeValue(ctx, path, ft.(NonNull).Unwrap(), fs, result)
	case ListKind:
		lt := ft.(List)
		v := reflect.ValueOf(result)
		res := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			res[i] = completeValue(ctx, append(path, i), lt.Unwrap(), fs, v.Index(i).Interface())
		}
		return res
	}
	return result
}
