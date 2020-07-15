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
	Data       map[string]interface{} `json:"data,omitempty"`
	Errors     []*Error               `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

type Executor struct {
	config *ExecutorConfig
}

type ExecutorConfig struct {
	GoroutineLimit   int
	EnableGoroutines bool
	Schema           *Schema
	Extensions       []Extension
}

func DefaultExecutor(s *Schema) *Executor {
	return &Executor{
		config: &ExecutorConfig{
			EnableGoroutines: false,
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
	for _, exts := range e.config.Extensions {
		ctx = exts.Init(ctx, p)
	}

	callExtensions(ctx, e.config.Extensions, EventParseStart, nil)
	doc, err := parser.Parse([]byte(p.Query))
	callExtensions(ctx, e.config.Extensions, EventParseFinish, err)
	if err != nil {
		return &Result{
			Errors: Errors{
				&Error{
					err.Error(),
					[]*ErrorLocation{
						{
							Column: err.(*parser.ParserError).Column,
							Line:   err.(*parser.ParserError).Line,
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
	gqlctx.extensions = e.config.Extensions

	callExtensions(ctx, e.config.Extensions, EventValidationStart, nil)
	validate(gqlctx)
	callExtensions(ctx, e.config.Extensions, EventValidationFinish, gqlctx.res.Errors)
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

	callExtensions(ctx, e.config.Extensions, EventExecutionStart, gqlctx.operation)
	resolveOperation(gqlctx)
	callExtensions(ctx, e.config.Extensions, EventExecutionFinish, gqlctx.res)

	for _, ext := range e.config.Extensions {
		if r := ext.Result(ctx); r != nil {
			if gqlctx.res.Extensions == nil {
				gqlctx.res.Extensions = map[string]interface{}{}
			}
			gqlctx.res.Extensions[ext.GetName()] = r
		}
	}
	return gqlctx.res
}

func (e *Executor) Subscribe(ctx context.Context, query string, operationName string, variables map[string]interface{}) (<-chan interface{}, error) {
	if e.config.Schema.Subscription == nil {
		return nil, errors.New("Schema does not provide subscriptions")
	}
	p := Params{
		Query:         query,
		OperationName: operationName,
		Variables:     variables,
	}

	doc, err := parser.Parse([]byte(p.Query))
	if err != nil {
		return nil, err
	}
	types, directives, implementors := getTypes(e.config.Schema)

	gqlctx := newContext(ctx, e.config.Schema, doc, &p, e.config.GoroutineLimit, e.config.EnableGoroutines)
	gqlctx.types = types
	gqlctx.directives = directives
	gqlctx.implementors = implementors

	validate(gqlctx)
	if len(gqlctx.res.Errors) > 0 {
		return nil, errors.New("validation error: invalid document")
	}

	getOperation(gqlctx)
	if len(gqlctx.res.Errors) > 0 {
		return nil, errors.New("invalid operation")
	}

	coerceVariableValues(gqlctx)
	if len(gqlctx.res.Errors) > 0 {
		return nil, errors.New("invalid variables")
	}

	return subscribe(gqlctx)
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
			defaultValue, err := coerceValue(ctx, varDef.DefaultValue, varType)
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
			} else if cv, err := coerceValue(ctx, value, varType); err == nil {
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

func resolveOperation(ctx *gqlCtx) {
	switch ctx.operation.OperationType {
	case ast.Query:
		ctx.res = executeQuery(ctx, ctx.operation)
	case ast.Mutation:
		ctx.res = executeMutation(ctx, ctx.operation)
	default:
		ctx.res = &Result{
			Errors: Errors{&Error{"invalid operation", nil, nil, nil}},
		}
	}
}

func executeQuery(ctx *gqlCtx, op *ast.Operation) *Result {
	rmap, hasNullErrs := executeSelectionSet(ctx, []interface{}{}, op.SelectionSet, ctx.schema.Query, ctx.schema.RootValue)
	if hasNullErrs {
		ctx.res.Data = nil
	} else {
		ctx.res.Data = rmap
	}
	return ctx.res
}

func executeMutation(ctx *gqlCtx, op *ast.Operation) *Result {
	rmap, hasNullErrs := executeSelectionSet(ctx, []interface{}{}, op.SelectionSet, ctx.schema.Mutation, ctx.schema.RootValue)
	if hasNullErrs {
		ctx.res.Data = nil
	} else {
		ctx.res.Data = rmap
	}
	return ctx.res
}

func subscribe(ctx *gqlCtx) (<-chan interface{}, error) {
	gfields := collectFields(ctx, ctx.schema.Subscription, ctx.operation.SelectionSet, nil)

	out := make(chan interface{})

	// Since the subscription operations must have ONE selection ONLY
	// it's not a problem to run the field.Subscribe function serially
	for rkey, fs := range gfields {
		fieldName := fs[0].Name
		if !strings.HasPrefix(fieldName, "__") {
			res, err := ctx.schema.Subscription.Fields[fieldName].Resolver(
				&resolveContext{
					ctx:    ctx.ctx, // this is the original context
					gqlCtx: ctx,     // execution context
					args:   coerceArgumentValues(ctx, []interface{}{rkey}, ctx.schema.Subscription, fs[0]),
					parent: ctx.schema.RootValue, // root value
					path:   []interface{}{rkey},
				},
			)
			if err != nil {
				return nil, err
			}

			if ch, ok := res.(chan interface{}); ok {
				go func() {
					for v := range ch {
						res, hasErr := completeValue(ctx, nil, ctx.schema.Subscription.Fields[fieldName].Type, fs, v)
						if hasErr {
							out <- &Result{
								Data:   nil,
								Errors: ctx.res.Errors,
							}
							ctx.res.Errors = []*Error{}
						}
						out <- &Result{
							Data: map[string]interface{}{
								rkey: res,
							},
						}
					}
					close(out)
				}()
			}
			return out, nil
		}
	}
	return nil, errors.New("invalid subscription")
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
	return completeValue(ctx, path, ot.Fields[f.Name].GetType(), fs, resolveFieldValue(ctx, path, f, ot, ov, f.Name, coerceArgumentValues(ctx, path, ot, f)))
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
				value, hasValue = ctx.variables[varName]
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
				coercedVal, err := coerceValue(ctx, value, argDef.Type)
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

func coerceValue(ctx *gqlCtx, val interface{}, t Type) (interface{}, error) {
	switch {
	case t.GetKind() == NonNullKind:
		if _, ok := val.(ast.NullValue); ok || val == nil {
			return nil, errors.New("Null value on NonNull type")
		}
		return coerceValue(ctx, val, t.(*NonNull).Unwrap())
	case t.GetKind() == ListKind:
		wt := t.(*List).Unwrap()
		switch val := val.(type) {
		case *ast.ListValue:
			res := make([]interface{}, len(val.Values))
			for i := 0; i < len(res); i++ {
				r, err := coerceValue(ctx, val.Values[i].GetValue(), wt)
				if err != nil {
					return nil, err
				}
				res[i] = r
			}
			return res, nil
		case []interface{}:
			res := make([]interface{}, len(val))
			for i := 0; i < len(res); i++ {
				r, err := coerceValue(ctx, val[i], wt)
				if err != nil {
					return nil, err
				}
				res[i] = r
			}
			return res, nil
		}
		return nil, fmt.Errorf("invalid list value")
	case t.GetKind() == ScalarKind:
		return t.(*Scalar).CoerceInput(val)
	case t.GetKind() == EnumKind:
		e := t.(*Enum)
		switch val := val.(type) {
		case ast.Value:
			for _, ev := range e.Values {
				if ev.Name == val.GetValue().(string) {
					return ev.Value, nil
				}
			}
		case string:
			for _, ev := range e.Values {
				if ev.Name == val {
					return ev.Value, nil
				}
			}
		}
		return nil, fmt.Errorf("invalid enum value")
	case t.GetKind() == InputObjectKind:
		res := map[string]interface{}{}
		o := t.(*InputObject)
		switch val := val.(type) {
		case *ast.ObjectValue:
			for _, astf := range val.Fields {
				field, ok := o.Fields[astf.Name]
				if !ok {
					return nil, fmt.Errorf("field '%s' is not defined on '%s'", astf.Name, o.Name)
				}

				if astf.Value.Kind() == ast.NullValueKind && field.Type.GetKind() != NonNullKind {
					res[astf.Name] = nil
				}
				if astf.Value.Kind() != ast.VariableValueKind {
					if fv, err := coerceValue(ctx, astf.Value, field.Type); err == nil {
						res[astf.Name] = fv
					} else {
						return nil, err
					}
				}
				if astf.Value.Kind() == ast.VariableValueKind {
					vv := astf.Value.(*ast.VariableValue)

					ctx.mu.Lock()
					varVal, ok := ctx.variables[vv.Name]
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
							defVal, err := coerceValue(ctx, vDef.DefaultValue, field.Type)
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
		case map[string]interface{}:
			for fn, field := range o.GetFields() {
				fv, ok := val[fn]
				if !ok && field.IsDefaultValueSet() {
					res[fn] = field.DefaultValue
				} else if !ok && field.Type.GetKind() == NonNullKind {
					return nil, fmt.Errorf("No value provided for NonNull type")
				}
				if fv == nil && field.Type.GetKind() != NonNullKind {
					res[fn] = nil
				}
				if fv != nil {
					if cv, err := coerceValue(ctx, fv, field.Type); err == nil {
						res[fn] = cv
					} else {
						return nil, err
					}
				}
			}
			return res, nil
		}
	}
	return nil, errors.New("invalid object value")
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
		v := reflect.ValueOf(ctx.Parent())
		if v.IsValid() && v.Type().Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if !v.IsValid() {
			return nil, nil
		}

		if v.Type().Kind() != reflect.Struct {
			return nil, nil
		}

		for i := 0; i < v.Type().NumField(); i++ {
			// Get the field, returns https://golang.org/pkg/reflect/#StructField
			field := v.Type().Field(i)
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

	// call directives for field definition
	for _, d := range ot.Fields[fn].Directives {
		if di, ok := d.(FieldDefinitionDirective); ok {
			r = di.VisitFieldDefinition(ctx.ctx, *ot.Fields[fn], r)
		}
	}

	resCtx := &resolveContext{
		ctx:    ctx.ctx, // this is the original context
		gqlCtx: ctx,     // execution context
		args:   args,
		parent: ov, // parent's value
		path:   path,
	}

	callExtensions(ctx.ctx, ctx.extensions, EventFieldResolverStart, resCtx)
	v, err := r(resCtx)
	callExtensions(ctx.ctx, ctx.extensions, EventFieldResolverFinish, v)
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

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
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
	} else if isNil(result) {
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
		"String":   String,
		"Boolean":  Boolean,
		"Int":      Int,
		"ID":       ID,
		"Float":    Float,
		"DateTime": DateTime,
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
	if s.Subscription != nil {
		typeWalker(types, directives, implementors, s.Subscription)
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
	} else {
		return
	}
	// TODO: directives are not checked and "walked" through
	if hf, ok := wt.(hasFields); ok {
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
			for _, arg := range f.Arguments {
				typeWalker(types, directives, implementors, arg.Type)
			}
			typeWalker(types, directives, implementors, f.Type)
		}
	} else if u, ok := wt.(*Union); ok {
		for _, m := range u.Members {
			typeWalker(types, directives, implementors, m)
		}
	} else if io, ok := wt.(*InputObject); ok {
		for _, f := range io.Fields {
			typeWalker(types, directives, implementors, f.Type)
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
