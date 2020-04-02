package gql

import (
	"context"
	"encoding/json"
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

type ExecuteParams struct {
	Query         string `json:"query"`
	Variables     string `json:"variables"`
	OperationName string `json:"operationName"`
}

func Execute(ctx context.Context, s *Schema, p ExecuteParams) *Result {
	_, doc, err := parser.Parse(p.Query)
	if err != nil {
		return &Result{
			Errors: Errors{&Error{err.Error(), nil, nil, nil}},
		}
	}
	types, directives := getTypes(s)
	ectx := newCtx(
		ctx,
		map[string]interface{}{
			keyQuery:         doc,
			keyRawQuery:      p.Query,
			keyOperationName: p.OperationName,
			keyRawVariables:  p.Variables,
			keySchema:        s,
			keyTypes:         types,
			keyDirectives:    directives,
		},
		100,
		true,
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
	/*
		coercedValues := map[string]interface{}{}

		types := ctx.Get(keyTypes).(map[string]Type)
		op := ctx.Get(keyOperation).(*ast.Operation)
	*/
	raw := ctx.Get(keyRawVariables).(string)

	if raw == "" {
		ctx.Set(keyVariables, map[string]interface{}{})
		return
	}

	rawVars := map[string]interface{}{}

	err := json.Unmarshal([]byte(raw), &rawVars)
	if err != nil {
		ctx.addErr(&Error{
			Message:   "invalid variables format",
			Path:      []interface{}{},
			Locations: []*ErrorLocation{},
		})
	}
	ctx.Set(keyVariables, rawVars)
	/*
		for _, varDef := range op.Variables {
			varType, err := resolveAstType(ctx, types, varDef.Type)
			if err != nil {
				ctx.res.addErr(&Error{
					Message: "operation not found",
					Path:    []interface{}{},
					Locations: []*ErrorLocation{
						&ErrorLocation{
							Column: varDef.Location.Column,
							Line:   varDef.Location.Line,
						},
					},
				})
				return
			}
			hasValue := false
			if val, ok := rawVars[varDef.Name]; ok {

			}
			defVal := varDef.DefaultValue

		}
	*/
}

func resolveAstType(ctx *eCtx, types map[string]Type, t ast.Type) (Type, *Error) {

	return nil, nil
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
	rmap := executeSelectionSet(ctx, []interface{}{}, op.SelectionSet, schema.GetRootQuery(), nil)
	ctx.res.Data = rmap
	return ctx.res
}

func executeMutation(ctx *eCtx, op *ast.Operation) *Result {
	schema := ctx.Get(keySchema).(*Schema)
	rmap := executeSelectionSet(ctx, []interface{}{}, op.SelectionSet, schema.GetRootMutation(), nil)
	ctx.res.Data = rmap
	return ctx.res
}

func executeSelectionSet(ctx *eCtx, path []interface{}, ss []ast.Selection, ot *Object, ov interface{}) map[string]interface{} {
	gfields := collectFields(ctx, ot, ss, nil)
	resMap := map[string]interface{}{}
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
				var rval interface{}
				if strings.HasPrefix(fieldName, "__") {
					rval = resolveMetaFields(ctx, fs, ot)
				} else {
					fieldType := getFieldOfFields(fieldName, ot.GetFields()).GetType()
					rval = executeField(ctx, append(path, fs[0].Alias), ot, ov, fieldType, fs)
				}
				mu.Lock()
				resMap[rkey] = rval
				mu.Unlock()
				<-ctx.sem
				wg.Done()
			}()
		default:
			fieldName := fields[0].Name
			var rval interface{}
			if strings.HasPrefix(fieldName, "__") {
				rval = resolveMetaFields(ctx, fs, ot)
			} else {
				fieldType := getFieldOfFields(fieldName, ot.GetFields()).GetType()
				rval = executeField(ctx, append(path, fs[0].Alias), ot, ov, fieldType, fs)
			}
			mu.Lock()
			resMap[rkey] = rval
			mu.Unlock()
			wg.Done()
		}
	}
	wg.Wait()
	return resMap
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

				fragment, ok := ctx.Get(keyQuery).(*ast.Document).Fragments[fSpread.Name]
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

func executeField(ctx *eCtx, path []interface{}, ot *Object, ov interface{}, ft Type, fs ast.Fields) interface{} {
	f := fs[0]
	fn := f.Name
	args := coerceArgumentValues(ctx, path, ot, f)
	resVal := resolveFieldValue(ctx, path, f, ot, ov, fn, args)
	return completeValue(ctx, path, getFieldOfFields(fn, ot.GetFields()).GetType(), fs, resVal)
}

func coerceArgumentValues(ctx *eCtx, path []interface{}, ot *Object, f *ast.Field) map[string]interface{} {
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
				coercedVal, err := coerceValue(ctx, value, argType)
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

func coerceValue(ctx *eCtx, val interface{}, t Type) (interface{}, error) {
	switch {
	case t.GetKind() == NonNullKind:
		if _, ok := val.(ast.NullValue); ok {
			return nil, errors.New("Null value on NonNull type")
		}
		return coerceValue(ctx, val, t.(*NonNull).Unwrap())
	case t.GetKind() == ListKind:
		wt := t.(*List).Unwrap()
		lv := val.(ast.ListValue)
		res := make([]interface{}, len(lv.Values))
		for i := 0; i < len(res); i++ {
			r, err := coerceValue(ctx, lv.Values[i].GetValue(), wt)
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
				if fv, err := coerceValue(ctx, astf.Value, field.GetType()); err == nil {
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
					res[field.Name] = varVal
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
						res[field.Name] = defVal
					} else {
						if field.IsDefaultValueSet() {
							res[field.Name] = field.GetDefaultValue()
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

func resolveMetaFields(ctx *eCtx, fs []*ast.Field, t Type) interface{} {
	switch fs[0].Name {
	case "__typename":
		return t.GetName()
	case "__schema":
		return completeValue(ctx, []interface{}{}, schemaIntrospection, fs, true)
	case "__type":
		return executeField(ctx, []interface{}{}, introspectionQuery, nil, typeIntrospection, fs)
	}
	return nil
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
	}
	return v
}

func completeValue(ctx *eCtx, path []interface{}, ft Type, fs ast.Fields, result interface{}) interface{} {
	if ft.GetKind() == NonNullKind {
		// Step 1 - NonNull kinds
		if result == nil {
			ctx.addErr(&Error{Message: "null value on a NonNull field", Path: path})
			return nil
		}
		return completeValue(ctx, path, ft.(*NonNull).Unwrap(), fs, result)
	} else if result == nil {
		// Step 2 - Return null if nil
		return nil
	} else if ft.GetKind() == ListKind {
		// Step 3 - go through the list and complete each value, then return result
		lt := ft.(*List)
		v := reflect.ValueOf(result)
		res := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			res[i] = completeValue(ctx, append(path, i), lt.Unwrap(), fs, v.Index(i).Interface())
		}
		return res
	} else if ft.GetKind() == ScalarKind {
		// Step 4.1 - coerce scalar value
		res, err := ft.(*Scalar).CoerceResult(result)
		if err != nil {
			ctx.addErr(&Error{Message: err.Error(), Path: path})
		}
		return res
	} else if ft.GetKind() == EnumKind {
		// Step 4.2 - coerce Enum value
		for _, ev := range ft.(*Enum).GetValues() {
			if ev.GetValue() == result {
				return ev.Name
			}
		}
		ctx.addErr(&Error{Message: "invalid result", Path: path})
		return nil
	} else if ft.GetKind() == ObjectKind {
		ot := ft.(*Object)
		subSel := fs[0].SelectionSet
		return executeSelectionSet(ctx, path, subSel, ot, result)
	} else if ft.GetKind() == InterfaceKind {
		ot := ft.(*Interface).Resolve(ctx, result)
		subSel := fs[0].SelectionSet
		return executeSelectionSet(ctx, path, subSel, ot, result)
	} else if ft.GetKind() == UnionKind {
		ot := ft.(*Union).Resolve(ctx, result)
		subSel := fs[0].SelectionSet
		return executeSelectionSet(ctx, path, subSel, ot, result)
	}
	return nil
}

func getTypes(s *Schema) (map[string]Type, map[string]Directive) {
	types := map[string]Type{}
	directives := map[string]Directive{
		"skip":       skipDirective,
		"include":    includeDirective,
		"deprecated": deprecatedDirective,
	}
	addIntrospectionTypes(types)
	typeWalker(types, directives, s.GetRootQuery())
	if s.GetRootMutation() != nil {
		typeWalker(types, directives, s.GetRootMutation())
	}
	return types, directives
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

func typeWalker(types map[string]Type, directives map[string]Directive, t Type) {
	wt := unwrapper(t)
	if _, ok := types[wt.GetName()]; !ok {
		types[wt.GetName()] = wt
	}
	// TODO: directives are not checked and "walked" through
	if hf, ok := t.(hasFields); ok {
		for _, f := range hf.GetFields() {
			for _, arg := range f.GetArguments() {
				typeWalker(types, directives, arg.Type)
			}
			typeWalker(types, directives, f.GetType())
		}
	} else if u, ok := wt.(*Union); ok {
		for _, m := range u.GetMembers() {
			typeWalker(types, directives, m)
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
