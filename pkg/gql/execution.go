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
			GoroutineLimit:   100,
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
	_, doc, err := parser.Parse(p.Query)
	if err != nil {
		return &Result{
			Errors: Errors{&Error{err.Error(), nil, nil, nil}},
		}
	}
	types, directives, implementors := getTypes(e.config.Schema)
	ectx := newCtx(
		ctx,
		map[string]interface{}{
			keyQuery:         doc,
			keyRawQuery:      p.Query,
			keyOperationName: p.OperationName,
			keyRawVariables:  p.Variables,
			keySchema:        e.config.Schema,
			keyTypes:         types,
			keyDirectives:    directives,
			keyImplementors:  implementors,
		},
		e.config.GoroutineLimit,
		e.config.EnableGoroutines,
	)
	validate(ectx)
	if len(ectx.res.Errors) > 0 {
		return ectx.res
	}

	getOperation(ectx)
	if len(ectx.res.Errors) > 0 {
		return ectx.res
	}

	coerceVariableValues(ectx)
	if len(ectx.res.Errors) > 0 {
		return ectx.res
	}

	return resolveOperation(ectx)
}

type Params struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

func Execute(ctx context.Context, s *Schema, p Params) *Result {
	return DefaultExecutor(s).Execute(ctx, p)
}

func getOperation(ctx *eCtx) {
	oname := ctx.Get(keyOperationName).(string)
	doc := ctx.Get(keyQuery).(*ast.Document)

	var op *ast.Operation
	if oname == "" {
		if len(doc.Operations) == 1 {
			op = doc.Operations[0]
		} else {
			ctx.addErr(&Error{
				Message:   "missing operationName",
				Path:      []interface{}{},
				Locations: []*ErrorLocation{},
			})
		}
	} else {
		for _, o := range doc.Operations {
			if oname == o.Name {
				op = o
			}
		}
	}
	if op != nil {
		ctx.Set(keyOperation, op)
		return
	}
	ctx.addErr(&Error{
		Message:   "operation not found",
		Path:      []interface{}{},
		Locations: []*ErrorLocation{},
	})
	return
}

func coerceVariableValues(ctx *eCtx) {
	coercedValues := map[string]interface{}{}

	types := ctx.Get(keyTypes).(map[string]Type)
	op := ctx.Get(keyOperation).(*ast.Operation)

	raw := ctx.Get(keyRawVariables).(map[string]interface{})

	for _, varDef := range op.Variables {
		varType, Err := resolveAstType(types, varDef.Type)
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
		var defaultValue interface{} = nil
		value, hasValue := raw[varDef.Name]
		if !hasValue && varDef.DefaultValue != nil {
			var err error
			defaultValue, err = coerceAstValue(ctx, varDef.DefaultValue, varType)
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
	ctx.Set(keyVariables, coercedValues)
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

func resolveOperation(ctx *eCtx) *Result {
	op := ctx.Get(keyOperation).(*ast.Operation)
	switch op.OperationType {
	case ast.Query:
		return executeQuery(ctx, op)
	case ast.Mutation:
		return executeMutation(ctx, op)
	case ast.Subscription:
		// TODO: implement ExecuteSubscription
		break
	}
	return &Result{
		Errors: Errors{&Error{"invalid operation", nil, nil, nil}},
	}
}

func executeQuery(ctx *eCtx, op *ast.Operation) *Result {
	schema := ctx.Get(keySchema).(*Schema)
	rmap, hasNullErrs := executeSelectionSet(ctx, []interface{}{}, op.SelectionSet, schema.GetRootQuery(), nil)
	if hasNullErrs {
		ctx.res.Data = nil
	} else {
		ctx.res.Data = rmap
	}
	return ctx.res
}

func executeMutation(ctx *eCtx, op *ast.Operation) *Result {
	schema := ctx.Get(keySchema).(*Schema)
	rmap, hasNullErrs := executeSelectionSet(ctx, []interface{}{}, op.SelectionSet, schema.GetRootMutation(), nil)
	if hasNullErrs {
		ctx.res.Data = nil
	} else {
		ctx.res.Data = rmap
	}
	return ctx.res
}

func executeSelectionSet(ctx *eCtx, path []interface{}, ss []ast.Selection, ot *Object, ov interface{}) (map[string]interface{}, bool) {
	gfields := collectFields(ctx, ot, ss, nil)
	resMap := map[string]interface{}{}
	hasNullErrs := false
	if ctx.enableGoroutines && !reflect.DeepEqual(ot, ctx.Get(keySchema).(*Schema).GetRootMutation()) {
		wg := sync.WaitGroup{}
		wg.Add(len(gfields))
		mu := sync.Mutex{}
		for rkey, fields := range gfields {
			fs := fields
			rkey := rkey
			select {
			case ctx.sem <- struct{}{}:
				go func() {
					fieldName := fs[0].Name
					var (
						rval   interface{}
						hasErr bool
					)
					if strings.HasPrefix(fieldName, "__") {
						rval, hasErr = resolveMetaFields(ctx, fs, ot)
					} else {
						fieldType := getFieldOfFields(fieldName, ot.GetFields()).GetType()
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
				fieldName := fields[0].Name
				var (
					rval   interface{}
					hasErr bool
				)
				if strings.HasPrefix(fieldName, "__") {
					rval, hasErr = resolveMetaFields(ctx, fs, ot)
				} else {
					fieldType := getFieldOfFields(fieldName, ot.GetFields()).GetType()
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
				fieldType := getFieldOfFields(fieldName, ot.GetFields()).GetType()
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

func collectFields(ctx *eCtx, t *Object, ss []ast.Selection, vFrags []string) map[string]ast.Fields {
	types := ctx.Get(keyTypes).(map[string]Type)
	// log.Printf("types: %+v", types)

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
				f.ParentType = t.Name
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

				fragment, ok := ctx.Get(keyFragments).(map[string]*ast.Fragment)[fSpread.Name]
				if !ok {
					continue
				}

				if !doesFragmentTypeApply(ctx, t, types[fragment.TypeCondition]) {
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

				if f.TypeCondition != "" && !doesFragmentTypeApply(ctx, t, types[f.TypeCondition]) {
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

func doesFragmentTypeApply(ctx *eCtx, ot *Object, ft Type) bool {
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

func getFragmentSpread(ctx *eCtx, fragName string) (*ast.FragmentSpread, bool) {
	return nil, false
}

func getFieldOfFields(fn string, fs []*Field) *Field {
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

func executeField(ctx *eCtx, path []interface{}, ot *Object, ov interface{}, ft Type, fs ast.Fields) (interface{}, bool) {
	f := fs[0]
	fn := f.Name
	args := coerceArgumentValues(ctx, path, ot.GetFields(), f)
	resVal := resolveFieldValue(ctx, path, f, ot, ov, fn, args)
	return completeValue(ctx, path, getFieldOfFields(fn, ot.GetFields()).GetType(), fs, resVal)
}

func coerceArgumentValues(ctx *eCtx, path []interface{}, fields Fields, f *ast.Field) map[string]interface{} {
	vars := ctx.Get(keyVariables).(map[string]interface{})
	coercedVals := map[string]interface{}{}
	fieldName := f.Name
	argDefs := getFieldOfFields(fieldName, fields).GetArguments()
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
				coercedVal, err := coerceAstValue(ctx, value, argType)
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

func coerceJsonValue(ctx *eCtx, val interface{}, t Type) (interface{}, error) {
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
		for _, field := range o.GetFields() {
			fv, ok := ov[field.Name]
			if !ok && field.IsDefaultValueSet() {
				res[field.Name] = field.GetDefaultValue()
			} else if !ok && field.GetType().GetKind() == NonNullKind {
				return nil, fmt.Errorf("No value provided for NonNull type")
			}
			if fv == nil && field.GetType().GetKind() != NonNullKind {
				res[field.Name] = nil
			}
			if fv != nil {
				if cv, err := coerceJsonValue(ctx, fv, field.GetType()); err == nil {
					res[field.Name] = cv
				} else {
					return nil, err
				}
			}
		}
		return res, nil
	}
	return nil, errors.New("invalid value to coerce")
}

func coerceAstValue(ctx *eCtx, val interface{}, t Type) (interface{}, error) {
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
		for _, field := range o.GetFields() {
			astf, ok := ov.Fields[field.Name]
			if !ok && field.IsDefaultValueSet() {
				res[field.Name] = field.GetDefaultValue()
			} else if !ok && field.GetType().GetKind() == NonNullKind {
				return nil, fmt.Errorf("No value provided for NonNull type")
			}
			if astf.Value.Kind() == ast.NullValueKind && field.GetType().GetKind() != NonNullKind {
				res[field.Name] = nil
			}
			if astf.Value.Kind() != ast.VariableValueKind {
				if fv, err := coerceAstValue(ctx, astf.Value, field.GetType()); err == nil {
					res[field.Name] = fv
				} else {
					return nil, err
				}
			}
			if astf.Value.Kind() == ast.VariableValueKind {
				varVal, ok := ctx.Get(keyVariables).(map[string]interface{})
				if ok && varVal == nil && field.GetType().GetKind() == NonNullKind {
					return nil, fmt.Errorf("null value on NonNull type")
				} else if ok {
					res[field.Name] = varVal[field.Name]
				}
				if !ok {
					op := ctx.Get(keyOperation).(*ast.Operation)
					vv := astf.Value.(*ast.VariableValue)
					vDef := op.Variables[vv.Name]
					if vDef.DefaultValue != nil {
						defVal, err := coerceAstValue(ctx, vDef.DefaultValue, field.GetType())
						if err != nil {
							return nil, err
						}
						res[field.Name] = defVal
					} else {
						if field.IsDefaultValueSet() {
							res[field.Name] = field.GetDefaultValue()
						}
					}
				}
			}
		}
		return res, nil
	}
	return nil, errors.New("invalid value to coerce")
}

func resolveMetaFields(ctx *eCtx, fs []*ast.Field, t Type) (interface{}, bool) {
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

func resolveFieldValue(ctx *eCtx, path []interface{}, fast *ast.Field, ot *Object, ov interface{}, fn string, args map[string]interface{}) interface{} {
	f := getFieldOfFields(fn, ot.GetFields())
	rCtx := &resolveContext{
		ctx:    ctx.ctx, // this is the original context
		eCtx:   ctx,     // execution context
		args:   args,
		parent: ov, // parent's value
		path:   path,
		fields: []string{}, // currently it's not implemented
	}

	v, err := f.Resolve(rCtx)
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
	} else if f.GetType().GetKind() == NonNullKind && v == nil && err == nil {
		ctx.addErr(&Error{Message: "null value on a NonNull field", Path: path})
	}
	return v
}

// returns the completed value and a bool value if there is a NonNull error
func completeValue(ctx *eCtx, path []interface{}, ft Type, fs ast.Fields, result interface{}) (interface{}, bool) {
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
		if ctx.enableGoroutines {
			wg := sync.WaitGroup{}
			wg.Add(v.Len())
			mu := sync.Mutex{}
			for i := 0; i < v.Len(); i++ {
				select {
				case ctx.sem <- struct{}{}:
					i := i
					go func() {
						rval, hasErr := completeValue(ctx, append(path, i), lt.Unwrap(), fs, v.Index(i).Interface())
						if hasErr {
							mu.Lock()
							res[i] = nil
							mu.Unlock()
						} else {
							mu.Lock()
							res[i] = rval
							mu.Unlock()
						}
						<-ctx.sem
						wg.Done()
					}()
				default:
					rval, hasErr := completeValue(ctx, append(path, i), lt.Unwrap(), fs, v.Index(i).Interface())
					if hasErr {
						mu.Lock()
						res[i] = nil
						mu.Unlock()
					} else {
						mu.Lock()
						res[i] = rval
						mu.Unlock()
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
		ot := ft.(*Object)
		subSel := fs[0].SelectionSet
		return executeSelectionSet(ctx, path, subSel, ot, result)
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
	typeWalker(types, directives, implementors, s.GetRootQuery())
	if s.GetRootMutation() != nil {
		typeWalker(types, directives, implementors, s.GetRootMutation())
	}
	return types, directives, implementors
}

type hasFields interface {
	GetFields() []*Field
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
			for _, d := range v.GetDirectives() {
				if _, ok := directives[d.GetName()]; !ok {
					directives[d.GetName()] = d
				}
			}
		}
	}
}
