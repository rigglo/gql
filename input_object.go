package gql

import (
	"fmt"
	"sync"
)

// InputObject ...
type InputObject struct {
	Name        string
	Description string
	Fields      *InputFields
}

// InputFields ...
type InputFields struct {
	fields []*InputField
	cache  map[string]*InputField
	mu     sync.Mutex
}

func NewInputFields(fs ...*InputField) *InputFields {
	return &InputFields{
		fields: fs,
		cache:  make(map[string]*InputField),
		mu:     sync.Mutex{},
	}
}

// Add a new field to fields
func (fs *InputFields) Add(f *InputField) {
	fs.fields = append(fs.fields, f)
	fs.cache[f.Name] = f
}

func (fs *InputFields) Slice() []*InputField {
	return fs.fields
}

func (fs *InputFields) Len() int {
	return len(fs.fields)
}

func (fs *InputFields) buildCache() {
	fs.cache = make(map[string]*InputField)
	fs.mu.Lock()
	for _, f := range fs.fields {
		fs.cache[f.Name] = f
	}
	fs.mu.Unlock()
}

// Get returns the field with the given name (if exists)
func (fs *InputFields) Get(name string) (*InputField, error) {
	if fs.cache == nil || len(fs.cache) != len(fs.fields) {
		fs.buildCache()
	}
	if f, ok := fs.cache[name]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("there's no field named '%s'", name)
}

// InputField describes a field of an InputObject
type InputField struct {
	Name         string
	Type         Type
	Description  string
	DefaultValue interface{}
}

// Unwrap is defined to implement the Type interface
func (o *InputObject) Unwrap() Type {
	return nil
}

// Kind returns the kind of the type, which is InputObjectTypeDefinition
func (o *InputObject) Kind() TypeDefinition {
	return InputObjectTypeDefinition
}

// Validate runs a check if everything is okay with the Type
func (o *InputObject) Validate(ctx *ValidationContext) error {
	return nil
}
