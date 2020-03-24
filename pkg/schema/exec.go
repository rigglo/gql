package schema

import (
	"context"
	"errors"
	"fmt"
	"log"
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
		keyTypes:         getTypes(s),
	})
	validate(ectx)
	if len(ectx.res.Errors) > 0 {
		return ectx.res
	}
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
	rmap := executeSelectionSet(ctx, []interface{}{}, op.SelectionSet, schema.GetRootQuery(), nil)
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
				value = argVal.Value
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
						Column: f.Location.Column,
						Line:   f.Location.Line,
					},
				},
			})
		} else if hasValue {
			if value == nil {
				coercedVals[argName] = value
			} else if argVal.Value.Kind() == ast.VariableValueKind {
				coercedVals[argName] = value
			} else {
				coercedVal, err := coerceValue(ctx, value, argType)
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

func coerceValue(ctx *eCtx, val interface{}, t Type) (interface{}, error) {
	switch {
	case t.GetKind() == NonNullKind:
		if _, ok := val.(ast.NullValue); ok {
			return nil, errors.New("Null value on NonNull type")
		}
		return coerceValue(ctx, val, t.(NonNull).Unwrap())
	case t.GetKind() == ListKind:
		wt := t.(List).Unwrap()
		lv := val.(ast.ListValue)
		res := make([]interface{}, len(lv.Values))
		for i := 0; i < len(res); i++ {
			r, err := coerceValue(ctx, lv.Values[i].GetValue(), wt)
			if err != nil {
				// TODO: add location info with custom error
				return nil, err
			}
			res[i] = r
		}
		return res, nil
	case t.GetKind() == ScalarKind:
		s := t.(ScalarType)
		if v, ok := val.(ast.Value); ok {
			// TODO: this should be nicer...
			return s.CoerceInput([]byte(v.GetValue().(string)))
		}
		return nil, fmt.Errorf("invalid value on a Scalar type")
	case t.GetKind() == InputObjectKind:
		log.Printf("%+v, %+v", val, t)
		res := map[string]interface{}{}
		ov, ok := val.(*ast.ObjectValue)
		if !ok {
			return nil, fmt.Errorf("Invalid input object value")
		}
		o := t.(InputObjectType)
		for name, field := range o.GetFields() {
			astf, ok := ov.Fields[name]
			if !ok && field.IsDefaultValueSet() {
				res[name] = field.GetDefaultValue()
			} else if !ok && field.GetType().GetKind() == NonNullKind {
				return nil, fmt.Errorf("No value provided for NonNull type")
			}
			if astf.Value.Kind() == ast.NullValueKind && field.GetType().GetKind() != NonNullKind {
				res[name] = nil
			}
			if astf.Value.Kind() != ast.VariableValueKind {
				if fv, err := coerceValue(ctx, astf.Value, field.GetType()); err == nil {
					res[name] = fv
				} else {
					return nil, err
				}
			}
			if astf.Value.Kind() == ast.VariableValueKind {
				varVal, ok := ctx.Get(keyVariables).(map[string]interface{})
				if ok && varVal == nil && field.GetType().GetKind() == NonNullKind {
					return nil, fmt.Errorf("null value on NonNull type")
				} else if ok {
					res[name] = varVal
				}
				if !ok {
					op := ctx.Get(keyOperation).(*ast.Operation)
					vv := astf.Value.(*ast.VariableValue)
					vDef := op.Variables[vv.Name]
					if vDef.DefaultValue != nil {
						defVal, err := coerceValue(ctx, vDef.DefaultValue, field.GetType())
						if err != nil {
							return nil, err
						}
						res[name] = defVal
					} else {
						if field.IsDefaultValueSet() {
							res[name] = field.GetDefaultValue()
						}
					}
				}
			}

			// TODO: resolve variable value
		}
		return res, nil
	}
	return nil, errors.New("invalid value to coerce")
}

func resolveFieldValue(ctx *eCtx, path []interface{}, fast *ast.Field, ot ObjectType, ov interface{}, fn string, args map[string]interface{}) interface{} {
	f := getFieldOfFields(fn, ot.GetFields())
	v, err := f.Resolve(ctx, ov, args)
	if err != nil {
		// TODO: add option to return custom error
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
	if ft.GetKind() == NonNullKind {
		// Step 1 - NonNull kinds
		if result == nil {
			ctx.res.addErr(&Error{Message: "null value on a NonNull field", Path: path})
			return nil
		}
		return completeValue(ctx, path, ft.(NonNull).Unwrap(), fs, result)
	} else if result == nil {
		// Step 2 - Return null if nil
		return nil
	} else if ft.GetKind() == ListKind {
		// Step 3 - go through the list and complete each value, then return result
		lt := ft.(List)
		v := reflect.ValueOf(result)
		res := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			res[i] = completeValue(ctx, append(path, i), lt.Unwrap(), fs, v.Index(i).Interface())
		}
		return res
	} else if ft.GetKind() == ScalarKind {
		// Step 4.1 - coerce scalar value
		res, err := ft.(ScalarType).CoerceResult(result)
		if err != nil {
			ctx.res.addErr(&Error{Message: err.Error(), Path: path})
		}
		return res
	} else if ft.GetKind() == EnumKind {
		// Step 4.2 - coerce Enum value
		for name, ev := range ft.(EnumType).GetValues() {
			if ev.GetValue() == result {
				return name
			}
		}
		ctx.res.addErr(&Error{Message: "invalid result", Path: path})
		return nil
	} else if ft.GetKind() == ObjectKind {
		ot := ft.(ObjectType)
		subSel := fs[0].SelectionSet
		return executeSelectionSet(ctx, path, subSel, ot, result)
	} else if ft.GetKind() == InterfaceKind {
		ot := ft.(InterfaceType).Resolve(ctx, result)
		subSel := fs[0].SelectionSet
		return executeSelectionSet(ctx, path, subSel, ot, result)
	}
	return nil
}
