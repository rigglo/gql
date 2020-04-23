package gql

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/rigglo/gql/pkg/language/ast"
	"github.com/rigglo/gql/pkg/language/parser"
)

type Result struct {
	Data   map[string]interface{} `json:"data"`
	Errors []*Error               `json:"errors,omitempty"`
}

type Executor struct {
	config *ExecutorConfig
}

type ExecutorConfig struct {
	GoroutineLimit   int
	EnableGoroutines bool
	Schema           *Schema
}

func DefaultExecutor(s *Schema) *Executor {
	return &Executor{
		config: &ExecutorConfig{
			EnableGoroutines: true,
			GoroutineLimit:   10,
			Schema:           s,
		},
	}
}

func NewExecutor(c ExecutorConfig) *Executor {
	return &Executor{
		config: &c,
	}
}

func (e *Executor) Execute(ctx context.Context, p Params) *Result {
	t, doc, err := parser.Parse([]byte(p.Query))
	if err != nil {
		return &Result{
			Errors: Errors{
				&Error{
					err.Error(),
					[]*ErrorLocation{
						{
							Column: t.Col,
							Line:   t.Line,
						},
					},
					nil,
					nil,
				},
			},
		}
	}
	types, directives, implementors := getTypes(e.config.Schema)

	gqlctx := newContext(ctx, e.config.Schema, doc, &p, e.config.GoroutineLimit, e.config.EnableGoroutines)
	gqlctx.types = types
	gqlctx.directives = directives
	gqlctx.implementors = implementors

	validate(gqlctx)
	if len(gqlctx.res.Errors) > 0 {
		return gqlctx.res
	}

	getOperation(gqlctx)
	if len(gqlctx.res.Errors) > 0 {
		return gqlctx.res
	}

	coerceVariableValues(gqlctx)
	if len(gqlctx.res.Errors) > 0 {
		return gqlctx.res
	}

	return resolveOperation(gqlctx)
}

type Params struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

func Execute(ctx context.Context, s *Schema, p Params) *Result {
	return DefaultExecutor(s).Execute(ctx, p)
}

func getOperation(ctx *gqlCtx) {

	var op *ast.Operation
	if ctx.params.OperationName == "" {
		if len(ctx.doc.Operations) == 1 {
			op = ctx.doc.Operations[0]
		} else {
			ctx.addErr(&Error{
				Message:   "missing operationName",
				Path:      []interface{}{},
				Locations: []*ErrorLocation{},
			})
		}
	} else {
		for _, o := range ctx.doc.Operations {
			if ctx.params.OperationName == o.Name {
				op = o
			}
		}
	}
	if op != nil {
		ctx.operation = op
		return
	}
	ctx.addErr(&Error{
		Message:   "operation not found",
		Path:      []interface{}{},
		Locations: []*ErrorLocation{},
	})
	return
}

func coerceVariableValues(ctx *gqlCtx) {
	coercedValues := map[string]interface{}{}

	for _, varDef := range ctx.operation.Variables {
		varType, Err := resolveAstType(ctx.types, varDef.Type)
		if Err != nil {
			ctx.addErr(&Error{
				Message: "invalid type for variable",
				Path:    []interface{}{},
				Locations: []*ErrorLocation{
					&ErrorLocation{
						Column: varDef.Location.Column,
						Line:   varDef.Location.Line,
					},
				},
			})
			continue
		}
		if !isInputType(varType) {
			ctx.addErr(&Error{
				Message: "variable type is not input type",
				Path:    []interface{}{},
				Locations: []*ErrorLocation{
					&ErrorLocation{
						Column: varDef.Location.Column,
						Line:   varDef.Location.Line,
					},
				},
			})
			continue
		}
		value, hasValue := ctx.params.Variables[varDef.Name]
		if !hasValue && varDef.DefaultValue != nil {
			defaultValue, err := coerceAstValue(ctx, varDef.DefaultValue, varType)
			if err != nil {
				ctx.addErr(&Error{
					Message: "couldn't coerece default value of variable",
					Path:    []interface{}{},
					Locations: []*ErrorLocation{
						&ErrorLocation{
							Column: varDef.Location.Column,
							Line:   varDef.Location.Line,
						},
					},
				})
				continue
			}
			coercedValues[varDef.Name] = defaultValue
		} else if varType.GetKind() == NonNullKind && (!hasValue || value == nil) {
			ctx.addErr(&Error{
				Message: "null value or missing value for non null type",
				Path:    []interface{}{},
				Locations: []*ErrorLocation{
					&ErrorLocation{
						Column: varDef.Location.Column,
						Line:   varDef.Location.Line,
					},
				},
			})
			continue
		} else if hasValue {
			if value == nil {
				coercedValues[varDef.Name] = nil
			} else if cv, err := coerceJsonValue(ctx, value, varType); err == nil {
				coercedValues[varDef.Name] = cv
			} else {
				ctx.addErr(&Error{
					Message: err.Error(),
					Path:    []interface{}{},
					Locations: []*ErrorLocation{
						&ErrorLocation{
							Column: varDef.Location.Column,
							Line:   varDef.Location.Line,
						},
					},
				})
				continue
			}
		}
	}
	ctx.variables = coercedValues
}

func resolveAstType(types map[string]Type, t ast.Type) (Type, *Error) {
	switch t.Kind() {
	case ast.List:
		t, err := resolveAstType(types, t.GetValue().(ast.Type))
		if err != nil {
			return nil, err
		}
		return NewList(t), nil
	case ast.NonNull:
		t, err := resolveAstType(types, t.GetValue().(ast.Type))
		if err != nil {
			return nil, err
		}
		return NewNonNull(t), nil
	case ast.Named:
		if t, ok := types[t.GetValue().(string)]; ok {
			return t, nil
		}
		return nil, &Error{
			Message: "invalid type",
			Locations: []*ErrorLocation{
				{
					Column: t.(*ast.NamedType).Location.Column,
					Line:   t.(*ast.NamedType).Location.Line,
				},
			},
		}
	}
	return nil, &Error{
		Message: "couldn't resolve ast type",
	}
}

func resolveOperation(ctx *gqlCtx) *Result {
	switch ctx.operation.OperationType {
	case ast.Query:
		return executeQuery(ctx, ctx.operation)
	case ast.Mutation:
		return executeMutation(ctx, ctx.operation)
	case ast.Subscription:
		// TODO: implement ExecuteSubscription
		break
	}
	return &Result{
		Errors: Errors{&Error{"invalid operation", nil, nil, nil}},
	}
}

func executeQuery(ctx *gqlCtx, op *ast.Operation) *Result {
	rmap, hasNullErrs := executeSelectionSet(ctx, []interface{}{}, op.SelectionSet, ctx.schema.Query, nil)
	if hasNullErrs {
		ctx.res.Data = nil
	} else {
		ctx.res.Data = rmap
	}
	return ctx.res
}

func executeMutation(ctx *gqlCtx, op *ast.Operation) *Result {
	rmap, hasNullErrs := executeSelectionSet(ctx, []interface{}{}, op.SelectionSet, ctx.schema.Mutation, nil)
	if hasNullErrs {
		ctx.res.Data = nil
	} else {
		ctx.res.Data = rmap
	}
	return ctx.res
}

func executeSelectionSet(ctx *gqlCtx, path []interface{}, ss []ast.Selection, ot *Object, ov interface{}) (map[string]interface{}, bool) {
	gfields := collectFields(ctx, ot, ss, nil)
	resMap := map[string]interface{}{}
	hasNullErrs := false
	conc := ctx.concurrency
	if ctx.schema.Mutation != nil {
		conc = ot.Name == ctx.schema.Mutation.Name
	}
	if conc {
		wg := sync.WaitGroup{}
		wg.Add(len(gfields))
		mu := sync.Mutex{}

		for rkey, fields := range gfields {
			fs := fields
			rkey := rkey
			select {
			case ctx.sem <- struct{}{}:
				go func() {
					var (
						rval   interface{}
						hasErr bool
					)
					if strings.HasPrefix(fs[0].Name, "__") {
						rval, hasErr = resolveMetaFields(ctx, fs, ot)
					} else {
						fieldType := ot.Fields[fs[0].Name].GetType()
						rval, hasErr = executeField(ctx, append(path, fs[0].Alias), ot, ov, fieldType, fs)
					}
					if hasErr {
						hasNullErrs = true
					}
					mu.Lock()
					resMap[rkey] = rval
					mu.Unlock()
					<-ctx.sem
					wg.Done()
				}()
			default:
				var (
					rval   interface{}
					hasErr bool
				)
				if strings.HasPrefix(fs[0].Name, "__") {
					rval, hasErr = resolveMetaFields(ctx, fs, ot)
				} else {
					fieldType := ot.Fields[fs[0].Name].GetType()
					rval, hasErr = executeField(ctx, append(path, fs[0].Alias), ot, ov, fieldType, fs)
				}
				if hasErr {
					hasNullErrs = true
				}
				mu.Lock()
				resMap[rkey] = rval
				mu.Unlock()
				wg.Done()
			}
		}
		wg.Wait()
	} else {
		for rkey, fs := range gfields {
			fieldName := fs[0].Name
			var (
				rval   interface{}
				hasErr bool
			)
			if strings.HasPrefix(fieldName, "__") {
				rval, hasErr = resolveMetaFields(ctx, fs, ot)
			} else {
				fieldType := ot.Fields[fieldName].GetType()
				rval, hasErr = executeField(ctx, append(path, fs[0].Alias), ot, ov, fieldType, fs)
			}
			if hasErr {
				hasNullErrs = true
			}
			resMap[rkey] = rval
		}
	}
	if hasNullErrs {
		return nil, true
	}
	return resMap, false
}

func collectFields(ctx *gqlCtx, t *Object, ss []ast.Selection, vFrags []string) map[string]ast.Fields {
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
			// TODO: resolve directives
		}
		if skip {
			continue
		}

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

				ctx.mu.Lock()
				fragment, ok := ctx.fragments[fSpread.Name]
				ctx.mu.Unlock()
				if !ok {
					continue
				}

				ctx.mu.Lock()
				tCond := ctx.types[fragment.TypeCondition]
				ctx.mu.Unlock()

				if !doesFragmentTypeApply(ctx, t, tCond) {
					continue
				}

				fgfields := collectFields(ctx, t, fragment.SelectionSet, vFrags)
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

				ctx.mu.Lock()
				tCond := ctx.types[f.TypeCondition]
				ctx.mu.Unlock()

				if f.TypeCondition != "" && !doesFragmentTypeApply(ctx, t, tCond) {
					continue
				}

				fgfields := collectFields(ctx, t, f.SelectionSet, vFrags)
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

func doesFragmentTypeApply(ctx *gqlCtx, ot *Object, ft Type) bool {
	if ft.GetKind() == ObjectKind && reflect.DeepEqual(ot, ft) {
		return true
	} else if ft.GetKind() == InterfaceKind {
		for _, i := range ot.GetInterfaces() {
			if i.GetName() == ft.GetName() {
				if reflect.DeepEqual(ft, i) {
					return true
				}
			}
		}
	} else if ft.GetKind() == UnionKind {
		for _, m := range ft.(*Union).GetMembers() {
			if m.GetName() == ot.GetName() {
				if reflect.DeepEqual(ot, m) {
					return true
				}
			}
		}
	}
	return false
}

func getArgOfArgs(an string, as []*ast.Argument) (*ast.Argument, bool) {
	for _, a := range as {
		if a.Name == an {
			return a, true
		}
	}
	return nil, false
}

func executeField(ctx *gqlCtx, path []interface{}, ot *Object, ov interface{}, ft Type, fs ast.Fields) (interface{}, bool) {
	f := fs[0]
	fn := f.Name
	args := coerceArgumentValues(ctx, path, ot, f)
	resVal := resolveFieldValue(ctx, path, f, ot, ov, fn, args)
	return completeValue(ctx, path, ot.Fields[fn].GetType(), fs, resVal)
}

func coerceArgumentValues(ctx *gqlCtx, path []interface{}, ot *Object, f *ast.Field) map[string]interface{} {
	coercedVals := map[string]interface{}{}
	argDefs := ot.Fields[f.Name].Arguments
	for argName, argDef := range argDefs {
		defaultValue := argDef.DefaultValue
		argVal, hasValue := getArgOfArgs(argName, f.Arguments)
		var value interface{}
		if argVal != nil {
			if argVal.Value.Kind() == ast.VariableValueKind {
				varName := argVal.Value.(*ast.VariableValue).Name

				//ctx.mu.Lock()
				value, hasValue = ctx.params.Variables[varName]
				//ctx.mu.Unlock()
			} else {
				value = argVal.Value
			}
		}
		if !hasValue && argDef.IsDefaultValueSet() {
			coercedVals[argName] = defaultValue
		} else if argDef.Type.GetKind() == NonNullKind && (!hasValue || value == nil) {
			ctx.addErr(&Error{
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
				coercedVal, err := coerceAstValue(ctx, value, argDef.Type)
				if err != nil {
					ctx.addErr(&Error{
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

func coerceJsonValue(ctx *gqlCtx, val interface{}, t Type) (interface{}, error) {
	switch {
	case t.GetKind() == NonNullKind:
		if val == nil {
			return nil, errors.New("Null value on NonNull type")
		}
		return coerceJsonValue(ctx, val, t.(*NonNull).Unwrap())
	case t.GetKind() == ListKind:
		wt := t.(*List).Unwrap()
		lv, ok := val.([]interface{})
		if !ok {
			return nil, errors.New("invalid list value")
		}
		res := make([]interface{}, len(lv))
		for i := 0; i < len(res); i++ {
			r, err := coerceJsonValue(ctx, lv[i], wt)
			if err != nil {
				return nil, err
			}
			res[i] = r
		}
		return res, nil
	case t.GetKind() == ScalarKind:
		s := t.(*Scalar)
		return s.CoerceInput(val)
	case t.GetKind() == EnumKind:
		e := t.(*Enum)
		if v, ok := val.(string); ok {
			for _, ev := range e.Values {
				if ev.Name == v {
					return ev.Value, nil
				}
			}
			return nil, fmt.Errorf("invalid value for an Enum type")
		}
		return nil, fmt.Errorf("invalid value for an Enum type")
	case t.GetKind() == InputObjectKind:
		res := map[string]interface{}{}
		ov, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Invalid input object value")
		}
		o := t.(*InputObject)
		for fn, field := range o.GetFields() {
			fv, ok := ov[fn]
			if !ok && field.IsDefaultValueSet() {
				res[fn] = field.DefaultValue
			} else if !ok && field.Type.GetKind() == NonNullKind {
				return nil, fmt.Errorf("No value provided for NonNull type")
			}
			if fv == nil && field.Type.GetKind() != NonNullKind {
				res[fn] = nil
			}
			if fv != nil {
				if cv, err := coerceJsonValue(ctx, fv, field.Type); err == nil {
					res[fn] = cv
				} else {
					return nil, err
				}
			}
		}
		return res, nil
	}
	return nil, errors.New("invalid value to coerce")
}

func coerceAstValue(ctx *gqlCtx, val interface{}, t Type) (interface{}, error) {
	switch {
	case t.GetKind() == NonNullKind:
		if _, ok := val.(ast.NullValue); ok {
			return nil, errors.New("Null value on NonNull type")
		}
		return coerceAstValue(ctx, val, t.(*NonNull).Unwrap())
	case t.GetKind() == ListKind:
		wt := t.(*List).Unwrap()
		lv := val.(ast.ListValue)
		res := make([]interface{}, len(lv.Values))
		for i := 0; i < len(res); i++ {
			r, err := coerceAstValue(ctx, lv.Values[i].GetValue(), wt)
			if err != nil {
				return nil, err
			}
			res[i] = r
		}
		return res, nil
	case t.GetKind() == ScalarKind:
		s := t.(*Scalar)
		if v, ok := val.(ast.Value); ok {
			// TODO: this should be nicer...
			return s.CoerceInput([]byte(v.GetValue().(string)))
		}
		return nil, fmt.Errorf("invalid value on a Scalar type")
	case t.GetKind() == EnumKind:
		e := t.(*Enum)
		if v, ok := val.(ast.Value); ok {
			for _, ev := range e.Values {
				if ev.Name == v.GetValue().(string) {
					return ev.Value, nil
				}
			}
			return nil, fmt.Errorf("invalid value for an Enum type")
		}
		return nil, fmt.Errorf("invalid value for an Enum type")
	case t.GetKind() == InputObjectKind:
		res := map[string]interface{}{}
		ov, ok := val.(*ast.ObjectValue)
		if !ok {
			return nil, fmt.Errorf("Invalid input object value")
		}
		o := t.(*InputObject)
		for _, astf := range ov.Fields {
			field, ok := o.Fields[astf.Name]
			if !ok {
				return nil, fmt.Errorf("field '%s' is not defined on '%s'", astf.Name, o.Name)
			}

			if astf.Value.Kind() == ast.NullValueKind && field.Type.GetKind() != NonNullKind {
				res[astf.Name] = nil
			}
			if astf.Value.Kind() != ast.VariableValueKind {
				if fv, err := coerceAstValue(ctx, astf.Value, field.Type); err == nil {
					res[astf.Name] = fv
				} else {
					return nil, err
				}
			}
			if astf.Value.Kind() == ast.VariableValueKind {
				vv := astf.Value.(*ast.VariableValue)

				ctx.mu.Lock()
				varVal, ok := ctx.params.Variables[vv.Name]
				ctx.mu.Unlock()

				if ok && varVal == nil && field.Type.GetKind() == NonNullKind {
					return nil, fmt.Errorf("null value on NonNull type")
				} else if ok {
					res[astf.Name] = varVal
				} else {
					ctx.mu.Lock()
					vDef := ctx.variableDefs[ctx.operation.Name][vv.Name]
					ctx.mu.Unlock()

					if vDef.DefaultValue != nil {
						defVal, err := coerceAstValue(ctx, vDef.DefaultValue, field.Type)
						if err != nil {
							return nil, err
						}
						res[astf.Name] = defVal
					} else {
						if field.IsDefaultValueSet() {
							res[astf.Name] = field.DefaultValue
						}
					}
				}
			}
		}
		return res, nil
	}
	return nil, errors.New("invalid value to coerce")
}

func resolveMetaFields(ctx *gqlCtx, fs []*ast.Field, t Type) (interface{}, bool) {
	switch fs[0].Name {
	case "__typename":
		return t.GetName(), false
	case "__schema":
		return completeValue(ctx, []interface{}{}, schemaIntrospection, fs, true)
	case "__type":
		return executeField(ctx, []interface{}{}, introspectionQuery, nil, typeIntrospection, fs)
	}
	return nil, true
}

func defaultResolver(fname string) Resolver {
	return func(ctx Context) (interface{}, error) {
		t := reflect.TypeOf(ctx.Parent())
		v := reflect.ValueOf(ctx.Parent())
		for i := 0; i < t.NumField(); i++ {
			// Get the field, returns https://golang.org/pkg/reflect/#StructField
			field := t.Field(i)
			// Get the field tag value
			// TODO: check 'gql' tag first and if that does not exist, check 'json'
			tag := field.Tag.Get("json")
			if strings.Split(tag, ",")[0] == fname {
				return v.FieldByName(field.Name).Interface(), nil
			}
		}
		return nil, nil
	}
}

func resolveFieldValue(ctx *gqlCtx, path []interface{}, fast *ast.Field, ot *Object, ov interface{}, fn string, args map[string]interface{}) interface{} {
	var r Resolver
	if r = ot.Fields[fn].Resolver; r == nil {
		r = defaultResolver(fn)
	}
	v, err := r(&resolveContext{
		ctx:    ctx.ctx, // this is the original context
		gqlCtx: ctx,     // execution context
		args:   args,
		parent: ov, // parent's value
		path:   path,
	})
	if err != nil {
		if e, ok := err.(CustomError); ok {
			ctx.addErr(&Error{
				Message: e.GetMessage(),
				Path:    path,
				Locations: []*ErrorLocation{
					&ErrorLocation{
						Column: fast.Location.Column,
						Line:   fast.Location.Line,
					},
				},
				Extensions: e.GetExtensions(),
			})
		} else {
			ctx.addErr(&Error{
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
		v = nil
	} else if ot.Fields[fn].Type.GetKind() == NonNullKind && v == nil && err == nil {
		ctx.addErr(&Error{Message: "null value on a NonNull field", Path: path})
	}
	return v
}

// returns the completed value and a bool value if there is a NonNull error
func completeValue(ctx *gqlCtx, path []interface{}, ft Type, fs ast.Fields, result interface{}) (interface{}, bool) {
	if ft.GetKind() == NonNullKind {
		// Step 1 - NonNull kinds
		rval, hasErr := completeValue(ctx, path, ft.(*NonNull).Unwrap(), fs, result)
		if hasErr || rval == nil {
			return nil, true
		} else if ft.(*NonNull).Unwrap().GetKind() == ListKind {
			v := reflect.ValueOf(rval)
			for i := 0; i < v.Len(); i++ {
				if v.Index(i).Interface() == nil {
					return nil, true
				}
			}
		}
		return rval, false
	} else if result == nil {
		// Step 2 - Return null if nil
		return nil, false
	} else if ft.GetKind() == ListKind {
		// Step 3 - go through the list and complete each value, then return result
		lt := ft.(*List)
		v := reflect.ValueOf(result)
		res := make([]interface{}, v.Len())
		if ctx.concurrency {
			wg := sync.WaitGroup{}
			wg.Add(v.Len())
			//mu := sync.Mutex{}
			for i := 0; i < v.Len(); i++ {
				select {
				case ctx.sem <- struct{}{}:
					i := i
					go func() {
						rval, hasErr := completeValue(ctx, append(path, i), lt.Unwrap(), fs, v.Index(i).Interface())
						if hasErr {
							//mu.Lock()
							res[i] = nil
							//mu.Unlock()
						} else {
							//mu.Lock()
							res[i] = rval
							//mu.Unlock()
						}
						<-ctx.sem
						wg.Done()
					}()
				default:
					rval, hasErr := completeValue(ctx, append(path, i), lt.Unwrap(), fs, v.Index(i).Interface())
					if hasErr {
						//mu.Lock()
						res[i] = nil
						//mu.Unlock()
					} else {
						//mu.Lock()
						res[i] = rval
						//mu.Unlock()
					}
					wg.Done()
				}
			}
			wg.Wait()
		} else {
			for i := 0; i < v.Len(); i++ {
				rval, hasErr := completeValue(ctx, append(path, i), lt.Unwrap(), fs, v.Index(i).Interface())
				if hasErr {
					res[i] = nil
				} else {
					res[i] = rval
				}
			}
		}
		return res, false
	} else if ft.GetKind() == ScalarKind {
		// Step 4.1 - coerce scalar value
		res, err := ft.(*Scalar).CoerceResult(result)
		if err != nil {
			ctx.addErr(&Error{Message: err.Error(), Path: path})
		}
		return res, false
	} else if ft.GetKind() == EnumKind {
		// Step 4.2 - coerce Enum value
		for _, ev := range ft.(*Enum).GetValues() {
			if ev.GetValue() == result {
				return ev.Name, false
			}
		}
		ctx.addErr(&Error{Message: "invalid result", Path: path})
		return nil, false
	} else if ft.GetKind() == ObjectKind {
		return executeSelectionSet(ctx, path, fs[0].SelectionSet, ft.(*Object), result)
	} else if ft.GetKind() == InterfaceKind {
		ot := ft.(*Interface).Resolve(ctx.ctx, result)
		subSel := fs[0].SelectionSet
		return executeSelectionSet(ctx, path, subSel, ot, result)
	} else if ft.GetKind() == UnionKind {
		ot := ft.(*Union).Resolve(ctx.ctx, result)
		subSel := fs[0].SelectionSet
		return executeSelectionSet(ctx, path, subSel, ot, result)
	}
	return nil, true
}

func getTypes(s *Schema) (map[string]Type, map[string]Directive, map[string][]Type) {
	types := map[string]Type{
		"String":  String,
		"Boolean": Boolean,
		"Int":     Int,
		"ID":      ID,
		"Float":   Float,
	}
	directives := map[string]Directive{
		"skip":       skipDirective,
		"include":    includeDirective,
		"deprecated": deprecatedDirective,
	}
	implementors := map[string][]Type{}
	addIntrospectionTypes(types)
	typeWalker(types, directives, implementors, s.Query)
	if s.Mutation != nil {
		typeWalker(types, directives, implementors, s.Mutation)
	}
	return types, directives, implementors
}

type hasFields interface {
	GetFields() map[string]*Field
}

type WrappingType interface {
	Unwrap() Type
}

// unwrapper unwraps Type t, until it results in a real type, one of (scalar, enum, interface, union or an object)
func unwrapper(t Type) Type {
	if w, ok := t.(WrappingType); ok {
		return unwrapper(w.Unwrap())
	}
	return t
}

func typeWalker(types map[string]Type, directives map[string]Directive, implementors map[string][]Type, t Type) {
	wt := unwrapper(t)
	if _, ok := types[wt.GetName()]; !ok {
		types[wt.GetName()] = wt
	}
	// TODO: directives are not checked and "walked" through
	if hf, ok := t.(hasFields); ok {
		if o, ok := t.(*Object); ok {
			for _, i := range o.Implements {
				if os, ok := implementors[i.Name]; ok {
					os = append(os, o)
					implementors[i.Name] = os
				} else {
					implementors[i.Name] = []Type{o}
				}
				typeWalker(types, directives, implementors, i)
			}
		}
		for _, f := range hf.GetFields() {
			for _, arg := range f.GetArguments() {
				typeWalker(types, directives, implementors, arg.Type)
			}
			typeWalker(types, directives, implementors, f.GetType())
		}
	} else if u, ok := wt.(*Union); ok {
		for _, m := range u.GetMembers() {
			typeWalker(types, directives, implementors, m)
		}
	}
}

func gatherDirectives(directives map[string]Directive, t Type) {
	ds := []Directive{}
	switch t.GetKind() {
	case ScalarKind:
		ds = t.(*Scalar).GetDirectives()
		for _, d := range ds {
			if _, ok := directives[d.GetName()]; !ok {
				directives[d.GetName()] = d
			}
		}
	case ObjectKind:
		ds = t.(*Object).GetDirectives()
		for _, d := range ds {
			if _, ok := directives[d.GetName()]; !ok {
				directives[d.GetName()] = d
			}
		}
		for _, f := range t.(*Object).GetFields() {
			gatherDirectives(directives, unwrapper(f.GetType()))
		}
	case InterfaceKind:
		ds = t.(*Interface).GetDirectives()
		for _, d := range ds {
			if _, ok := directives[d.GetName()]; !ok {
				directives[d.GetName()] = d
			}
		}
		for _, f := range t.(*Interface).GetFields() {
			gatherDirectives(directives, unwrapper(f.GetType()))
		}
	case UnionKind:
		ds = t.(*Union).GetDirectives()
		for _, d := range ds {
			if _, ok := directives[d.GetName()]; !ok {
				directives[d.GetName()] = d
			}
		}
	case EnumKind:
		ds = t.(*Enum).GetDirectives()
		for _, d := range ds {
			if _, ok := directives[d.GetName()]; !ok {
				directives[d.GetName()] = d
			}
		}
		for _, v := range t.(*Enum).GetValues() {
			for _, d := range v.GetDirectives() {
				if _, ok := directives[d.GetName()]; !ok {
					directives[d.GetName()] = d
				}
			}
		}
	case InputObjectKind:
		ds = t.(*InputObject).GetDirectives()
		for _, d := range ds {
			if _, ok := directives[d.GetName()]; !ok {
				directives[d.GetName()] = d
			}
		}
		for _, v := range t.(*InputObject).GetFields() {
			for _, d := range v.Directives {
				if _, ok := directives[d.GetName()]; !ok {
					directives[d.GetName()] = d
				}
			}
		}
	}
}
