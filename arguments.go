package gql

import (
	"sync"
)

// Argument is an argument 
type Argument struct {
	Name              string
	DefaultValue      interface{}
	Description       string
	Depricated        bool
	DepricationReason string
	Type              Type
}

// Arguments contains multiple Arguments
type Arguments struct {
	args  []*Argument
	cache map[string]*Argument
	mu    sync.Mutex
}

// NewArguments returns a new Arguments that containse the given arguments
func NewArguments(args ...*Argument) *Arguments {
	return &Arguments{
		args:  args,
		cache: make(map[string]*Argument),
		mu:    sync.Mutex{},
	}
}

// Add a new field to fields
func (as *Arguments) Add(a *Argument) {
	as.args = append(as.args, a)
	as.cache[a.Name] = a
}

// Slice returns a slice of arguments
func (as *Arguments) Slice() []*Argument {
	return as.args
}

// Map returns the map of arguments
func (as *Arguments) Map() map[string]*Argument {
	return as.cache
}

// Len returns the count of arguments in Arguments
func (as *Arguments) Len() int {
	return len(as.args)
}

func (as *Arguments) buildCache() {
	as.cache = make(map[string]*Argument)
	as.mu.Lock()
	for _, a := range as.args {
		as.cache[a.Name] = a
	}
	as.mu.Unlock()
}

// Get returns the argument with the given name (if exists)
func (as *Arguments) Get(name string) (*Argument, bool) {
	if as.cache == nil || len(as.cache) != len(as.args) {
		as.buildCache()
	}
	if a, ok := as.cache[name]; ok {
		return a, true
	}
	return nil, false
}
