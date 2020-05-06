package ast

type Document struct {
	Operations  []*Operation
	Fragments   []*Fragment
	Definitions []Definition
}

func NewDocument() *Document {
	return &Document{
		Operations:  []*Operation{},
		Fragments:   []*Fragment{},
		Definitions: []Definition{},
	}
}
