package internal

import (
	"github.com/rigglo/gql/language/ast"
)

func NewOrderedFieldGroups() *OrderedFieldGroups {
	return &OrderedFieldGroups{
		Fields: map[string]ast.Fields{},
		Orders: []string{},
	}
}

type OrderedFieldGroups struct {
	Fields map[string]ast.Fields
	Orders []string
}

func (ofg *OrderedFieldGroups) Append(alias string, f *ast.Field) {
	if _, ok := ofg.Fields[alias]; ok {
		ofg.Fields[alias] = append(ofg.Fields[alias], f)
		return
	}
	ofg.Orders = append(ofg.Orders, alias)
	ofg.Fields[alias] = ast.Fields{f}
}

func (ofg *OrderedFieldGroups) ByName(name string) ast.Fields {
	return ofg.Fields[name]
}

func (ofg *OrderedFieldGroups) ByOrder(i int) ast.Fields {
	return ofg.Fields[ofg.Orders[i]]
}

func (ofg *OrderedFieldGroups) Len() int {
	return len(ofg.Orders)
}

func (ofg *OrderedFieldGroups) Iter() *ofgIterator {
	return newOfgIterator(ofg)
}

func newOfgIterator(ofg *OrderedFieldGroups) *ofgIterator {
	return &ofgIterator{
		i:   -1,
		ofg: ofg,
	}
}

type ofgIterator struct {
	i   int
	ofg *OrderedFieldGroups
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
