package ordered

import (
	"encoding/json"
	"fmt"
)

// NewMap creates a new Map
func NewMap() *Map {
	return &Map{
		Values: map[string]interface{}{},
		Orders: []string{},
	}
}

// Map ...
type Map struct {
	Values map[string]interface{}
	Orders []string
}

// MarshalJSON implements the json Marshaler interface
func (om *Map) MarshalJSON() ([]byte, error) {
	out := "{"
	var err error
	for i, key := range om.Orders {
		val, ok := om.Values[key].([]byte)
		if !ok {
			val, err = json.Marshal(om.Values[key])
			if err != nil {
				return nil, err
			}
		}
		out += fmt.Sprintf(`"%s": %s`, key, string(val))
		if i < len(om.Orders)-1 {
			out += ","
		}
	}
	out += "}"
	return []byte(out), nil
}

// Append appends a key value pair to the map
func (om *Map) Append(key string, value interface{}) {
	om.Orders = append(om.Orders, key)
	om.Values[key] = value
}

// ByName return a value ByName
func (om *Map) ByName(name string) interface{} {
	return om.Values[name]
}

// ByOrder returns a value ByOrder
func (om *Map) ByOrder(i int) interface{} {
	return om.Values[om.Orders[i]]
}

// Len returns the size of the map
func (om *Map) Len() int {
	return len(om.Orders)
}

// Iter returns an iterator for the map
func (om *Map) Iter() *MapIterator {
	return newMapIterator(om)
}

func newMapIterator(om *Map) *MapIterator {
	return &MapIterator{
		i:  -1,
		om: om,
	}
}

// MapIterator is an iterator for an Map
type MapIterator struct {
	i  int
	om *Map
}

// Next returns true if there's a next element in the Map
func (iter *MapIterator) Next() bool {
	if iter.om.Len()-1 > iter.i {
		iter.i++
		return true
	}
	return false
}

// Value returns the current element of the iterator
func (iter *MapIterator) Value() (string, interface{}) {
	return iter.om.Orders[iter.i], iter.om.ByOrder(iter.i)
}
