package gql

import (
	"encoding/json"
	"fmt"
	"github.com/rigglo/gql/utils/ast"
)

func newOrderedFieldGroups() *orderedFieldGroups {
	return &orderedFieldGroups{
		Fields: map[string]ast.Fields{},
		Orders: []string{},
	}
}

type orderedFieldGroups struct {
	Fields map[string]ast.Fields
	Orders []string
}

func (ofg *orderedFieldGroups) Append(alias string, f *ast.Field) {
	if _, ok := ofg.Fields[alias]; ok {
		ofg.Fields[alias] = append(ofg.Fields[alias], f)
		return
	}
	ofg.Orders = append(ofg.Orders, alias)
	ofg.Fields[alias] = ast.Fields{f}
}

func (ofg *orderedFieldGroups) ByName(name string) ast.Fields {
	return ofg.Fields[name]
}

func (ofg *orderedFieldGroups) ByOrder(i int) ast.Fields {
	return ofg.Fields[ofg.Orders[i]]
}

func (ofg *orderedFieldGroups) Len() int {
	return len(ofg.Orders)
}

func (ofg *orderedFieldGroups) Iter() *ofgIterator {
	return newOfgIterator(ofg)
}

func newOfgIterator(ofg *orderedFieldGroups) *ofgIterator {
	return &ofgIterator{
		i:   -1,
		ofg: ofg,
	}
}

type ofgIterator struct {
	i   int
	ofg *orderedFieldGroups
}

func (iter *ofgIterator) Next() bool {
	if iter.ofg.Len()-1 > iter.i {
		iter.i++
		return true
	}
	return false
}

func (iter *ofgIterator) Value() (string, ast.Fields) {
	return iter.ofg.Orders[iter.i], iter.ofg.ByOrder(iter.i)
}

// NewOrderedMap creates a new OrderedMap
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		Values: map[string]interface{}{},
		Orders: []string{},
	}
}

// OrderedMap ...
type OrderedMap struct {
	Values map[string]interface{}
	Orders []string
}

// MarshalJSON implements the json Marshaler interface
func (om *OrderedMap) MarshalJSON() ([]byte, error) {
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
func (om *OrderedMap) Append(key string, value interface{}) {
	om.Orders = append(om.Orders, key)
	om.Values[key] = value
}

// ByName return a value ByName
func (om *OrderedMap) ByName(name string) interface{} {
	return om.Values[name]
}

// ByOrder returns a value ByOrder
func (om *OrderedMap) ByOrder(i int) interface{} {
	return om.Values[om.Orders[i]]
}

// Len returns the size of the map
func (om *OrderedMap) Len() int {
	return len(om.Orders)
}

// Iter returns an iterator for the map
func (om *OrderedMap) Iter() *MapIterator {
	return newMapIterator(om)
}

func newMapIterator(om *OrderedMap) *MapIterator {
	return &MapIterator{
		i:  -1,
		om: om,
	}
}

// MapIterator is an iterator for an OrderedMap
type MapIterator struct {
	i  int
	om *OrderedMap
}

// Next returns true if there's a next element in the OrderedMap
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