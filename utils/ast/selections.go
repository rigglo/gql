package ast

/*
type SelectionKind int

const (
	FieldSelectionKind SelectionKind = iota
	FragmentSpreadSelectionKind
	InlineFragmentSelectionKind
)

type Selection struct {
	Kind           SelectionKind
	Field          *Field
	FragmentSpread *FragmentSpread
	InlineFragment *InlineFragment
}

func (s *Selection) GetKind() SelectionKind {
	return s.Kind
}

func (s *Selection) GetValue() interface{} {
	if s.Kind == FieldSelectionKind {
		return s.Field
	} else if s.Kind == FragmentSpreadSelectionKind {
		return s.FragmentSpread
	} else if s.Kind == InlineFragmentSelectionKind {
		return s.InlineFragment
	}
	return nil
}

type Field struct {
	Alias        string
	Name         string
	Arguments    []*Argument
	Directives   []*Directive
	SelectionSet []*Selection
}

type FragmentSpread struct {
	Name       string
	Directives []*Directive
}

type InlineFragment struct {
	TypeCondition string
	Directives    []*Directive
	SelectionSet  []*Selection
}
*/
