package gql

import (
	"sync"
)

// Argument ...
type Argument struct {
	Name              string
	DefaultValue      interface{}
	Description       string
	Depricated        bool
	DepricationReason string
	Type              Type
}

type Arguments struct {
	args  []*Argument
	cache map[string]*Argument
	mu    sync.Mutex
}

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

func (as *Arguments) Slice() []*Argument {
	return as.args
}

func (as *Arguments) Map() map[string]*Argument {
	return as.cache
}

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
