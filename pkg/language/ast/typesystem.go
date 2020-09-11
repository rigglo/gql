package ast

import (
	"encoding/json"
	"strings"
)

type DefinitionKind uint

const (
	SchemaKind DefinitionKind = iota
	ScalarKind
	ObjectKind
	InterfaceKind
	UnionKind
	EnumKind
	InputObjectKind
	DirectiveKind
)

type Definition interface {
	Kind() DefinitionKind
	String() string
}

type SchemaDefinition struct {
	Name           string
	Directives     []*Directive
	RootOperations map[OperationType]*NamedType
}

func (d *SchemaDefinition) Kind() DefinitionKind {
	return SchemaKind
}

func (d *SchemaDefinition) String() string {
	out := "schema "
	for _, dir := range d.Directives {
		out += dir.String() + " "
	}
	out += "{\n"
	if nt, ok := d.RootOperations[Query]; ok && nt != nil {
		out += "\tquery: " + nt.Name + "\n"
	}
	if nt, ok := d.RootOperations[Mutation]; ok && nt != nil {
		out += "\tmutation: " + nt.Name + "\n"
	}
	if nt, ok := d.RootOperations[Subscription]; ok && nt != nil {
		out += "\tsubscription: " + nt.Name + "\n"
	}
	out += "}\n"
	return out
}

type ScalarDefinition struct {
	Description string
	Name        string
	Directives  []*Directive
}

func (d *ScalarDefinition) Kind() DefinitionKind {
	return ScalarKind
}

func (d *ScalarDefinition) String() string {
	out := ""
	if d.Description != "" {
		out += `"""` + jsonEscape(d.Description) + "\"\"\"\n"
	}
	out += "scalar " + d.Name + " "
	for _, dir := range d.Directives {
		out += dir.String() + " "
	}
	out += "\n"
	return out
}

type ObjectDefinition struct {
	Description string
	Name        string
	Implements  []*NamedType
	Directives  []*Directive
	Fields      []*FieldDefinition
}

func (d *ObjectDefinition) Kind() DefinitionKind {
	return ObjectKind
}

func (d *ObjectDefinition) String() string {
	out := ""
	if d.Description != "" {
		out += `"""` + jsonEscape(d.Description) + "\"\"\"\n"
	}
	out += "type " + d.Name + " "
	for _, dir := range d.Directives {
		out += dir.String() + " "
	}
	out += "{\n"
	for i, f := range d.Fields {
		if i != 0 {
			out += "\n"
		}
		out += f.String()
	}
	out += "}\n"
	return out
}

type FieldDefinition struct {
	Description string
	Name        string
	Arguments   []*InputValueDefinition
	Type        Type
	Directives  []*Directive
}

func (d *FieldDefinition) String() string {
	out := ""
	if d.Description != "" {
		out += "\t\"\"\"" + jsonEscape(d.Description) + "\"\"\"\n"
	}
	out += "\t" + d.Name
	if d.Arguments != nil && len(d.Arguments) != 0 {
		out += "(\n\t\t"
		for i, a := range d.Arguments {
			out += a.String()
			if i+1 == len(d.Arguments) {
				out += "\n\t)"
			} else {
				out += "\n\t\t"
			}
		}
	}
	out += ": " + d.Type.String()
	for _, dir := range d.Directives {
		out += " " + dir.String()
	}
	out += "\n"
	return out
}

type InputValueDefinition struct {
	Description  string
	Name         string
	Type         Type
	DefaultValue Value
	Directives   []*Directive
}

func (d *InputValueDefinition) String() string {
	return d.Name + ": " + d.Type.String()
}

type InterfaceDefinition struct {
	Description string
	Name        string
	Directives  []*Directive
	Fields      []*FieldDefinition
}

func (d *InterfaceDefinition) Kind() DefinitionKind {
	return InterfaceKind
}

func (d *InterfaceDefinition) String() string {
	out := ""
	if d.Description != "" {
		out += `"""` + jsonEscape(d.Description) + "\"\"\"\n"
	}
	out += "interface " + d.Name + " "
	for _, dir := range d.Directives {
		out += dir.String() + " "
	}
	out += "{\n"
	for i, f := range d.Fields {
		if i != 0 {
			out += "\n"
		}
		out += f.String()
	}
	out += "}\n"
	return out
}

type UnionDefinition struct {
	Description string
	Name        string
	Directives  []*Directive
	Members     []*NamedType
}

func (d *UnionDefinition) Kind() DefinitionKind {
	return UnionKind
}

func (d *UnionDefinition) String() string {
	out := ""
	if d.Description != "" {
		out += `"""` + jsonEscape(d.Description) + "\"\"\"\n"
	}
	out += "union " + d.Name + " "
	for _, dir := range d.Directives {
		out += dir.String() + " "
	}
	out += "= "
	for i, m := range d.Members {
		if i != 0 {
			out += " | "
		}
		out += m.Name
	}
	return out + "\n"
}

type EnumDefinition struct {
	Description string
	Name        string
	Directives  []*Directive
	Values      []*EnumValueDefinition
}

func (d *EnumDefinition) Kind() DefinitionKind {
	return EnumKind
}

func (d *EnumDefinition) String() string {
	out := ""
	if d.Description != "" {
		out += `"""` + jsonEscape(d.Description) + "\"\"\"\n"
	}
	out += "enum " + d.Name + " "
	for _, dir := range d.Directives {
		out += dir.String() + " "
	}
	out += "{\n}\n"
	return out
}

type EnumValueDefinition struct {
	Description string
	Value       *EnumValue
	Directives  []*Directive
}

type InputObjectDefinition struct {
	Description string
	Name        string
	Directives  []*Directive
	Fields      []*InputValueDefinition
}

func (d *InputObjectDefinition) Kind() DefinitionKind {
	return InputObjectKind
}

func (d *InputObjectDefinition) String() string {
	out := ""
	if d.Description != "" {
		out += `"""` + jsonEscape(d.Description) + "\"\"\"\n"
	}
	out += "input " + d.Name + " "
	for _, dir := range d.Directives {
		out += dir.String() + " "
	}
	out += "{\n}\n"
	return out
}

type DirectiveDefinition struct {
	Description string
	Name        string
	Locations   []string
	Arguments   []*InputValueDefinition
}

func (d *DirectiveDefinition) Kind() DefinitionKind {
	return DirectiveKind
}

func (d *DirectiveDefinition) String() string {
	out := ""
	if d.Description != "" {
		out += `"""` + jsonEscape(d.Description) + "\"\"\"\n"
	}
	out += "directive @" + d.Name + " on " + strings.Join(d.Locations, " | ") + "\n"
	return out
}

func jsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s[1 : len(s)-1]
}
