package gql

import (
	"context"
	"fmt"
	"sync"
)

// Directive ..
type Directive struct {
	Name      string
	Arguments []*Argument
}

// Field ...
type Field struct {
	Name              string
	Description       string
	Depricated        bool
	DepricationReason string
	Type              Type
	Arguments         []*Argument
	Resolver          ResolverFunc
}

// ResolverFunc ..
type ResolverFunc func(context.Context, map[string]interface{}, interface{}) (interface{}, error)

// Fields ...
type Fields struct {
	fields []*Field
	cache  map[string]*Field
	mu     sync.Mutex
}

func NewFields(fs ...*Field) *Fields {
	return &Fields{
		fields: fs,
		cache:  make(map[string]*Field),
		mu:     sync.Mutex{},
	}
}

// Add a new field to fields
func (fs *Fields) Add(f *Field) {
	fs.fields = append(fs.fields, f)
	fs.cache[f.Name] = f
}

func (fs *Fields) Slice() []*Field {
	return fs.fields
}

func (fs *Fields) Len() int {
	return len(fs.fields)
}

func (fs *Fields) buildCache() {
	fs.cache = make(map[string]*Field)
	fs.mu.Lock()
	for _, f := range fs.fields {
		fs.cache[f.Name] = f
	}
	fs.mu.Unlock()
}

// Get returns the field with the given name (if exists)
func (fs *Fields) Get(name string) (*Field, error) {
	if fs.cache == nil || len(fs.cache) != len(fs.fields) {
		fs.buildCache()
	}
	if f, ok := fs.cache[name]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("there's no field named '%s'", name)
}

func (f *Field) Validate(ctx *ValidationContext) error {
	if err := f.Type.Validate(ctx); err != nil {
		return err
	}
	return nil
}
