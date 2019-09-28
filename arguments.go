package gql

import (
	"errors"

	"github.com/rigglo/gql/utils/parser"
)

type Argument struct {
	Name         string
	Type         Type
	DefaultValue interface{}
}

type Arguments map[string]*Argument

type ResolverArgs interface {
	Get(string) (interface{}, bool)
}

type argStore struct {
	args  map[string]interface{}
	isSet map[string]bool
}

func (as *argStore) Get(key string) (interface{}, bool) {
	return as.args[key], as.isSet[key]
}

func fromParsedArguments(args Arguments, lexArgs []*parser.Argument) (ResolverArgs, error) {
	as := new(argStore)
	as.args = make(map[string]interface{})
	as.isSet = make(map[string]bool)
	for _, lexArg := range lexArgs {
		if _, ok := args[lexArg.Key]; !ok {
			return nil, errors.New(ErrArgNotExists)
		}
		// TODO: Resolve object argument type
		scalar := args[lexArg.Key].Type.Value().(*Scalar)

		if !scalar.IsNullable() && (lexArg.Value == "" || lexArg.Value == "null") {
			return nil, errors.New(ErrNullValue)
		}

		val, err := scalar.Decode([]byte(lexArg.Value))
		if err != nil {
			return nil, err
		}
		as.args[lexArg.Key] = val
		as.isSet[lexArg.Key] = true
	}
	for _, arg := range args {
		if _, ok := as.args[arg.Name]; !ok {
			as.args[arg.Name] = arg.DefaultValue
		}
	}
	return as, nil
}
