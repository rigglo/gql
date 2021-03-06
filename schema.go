package gql

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/rigglo/gql/pkg/language/ast"
)

/*
Schema is a graphql schema, a root for the mutations, queries and subscriptions.

A GraphQL service’s collective type system capabilities are referred to as that
service’s “schema”. A schema is defined in terms of the types and directives it
supports as well as the root operation types for each kind of operation: query,
mutation, and subscription; this determines the place in the type system where those operations begin.
*/
type Schema struct {
	Query           *Object
	Mutation        *Object
	Subscription    *Object
	Directives      TypeSystemDirectives
	AdditionalTypes []Type
	RootValue       interface{}
}

// SDL generates an SDL string from your schema
func (s Schema) SDL() string {
	b := newSDLBuilder(&s)
	return b.Build()
}

// TypeKind shows the kind of a Type
type TypeKind uint

const (
	// ScalarKind is for the Scalars in the GraphQL type system
	ScalarKind TypeKind = iota
	// ObjectKind is for the Object in the GraphQL type system
	ObjectKind
	// InterfaceKind is for the Interface in the GraphQL type system
	InterfaceKind
	// UnionKind is for the Union in the GraphQL type system
	UnionKind
	// EnumKind is for the Enum in the GraphQL type system
	EnumKind
	// InputObjectKind is for the InputObject in the GraphQL type system
	InputObjectKind
	// NonNullKind is for the NonNull in the GraphQL type system
	NonNullKind
	// ListKind is for the List in the GraphQL type system
	ListKind
)

// Type represents a Type in the GraphQL Type System
type Type interface {
	GetName() string
	GetDescription() string
	GetKind() TypeKind
	String() string
}

func isInputType(t Type) bool {
	if t.GetKind() == ListKind || t.GetKind() == NonNullKind {
		return isInputType(t.(WrappingType).Unwrap())
	}
	if t.GetKind() == ScalarKind || t.GetKind() == EnumKind || t.GetKind() == InputObjectKind {
		return true
	}
	return false
}

func isOutputType(t Type) bool {
	if t.GetKind() == ListKind || t.GetKind() == NonNullKind {
		return isOutputType(t.(WrappingType).Unwrap())
	}
	if t.GetKind() == ScalarKind || t.GetKind() == ObjectKind || t.GetKind() == InterfaceKind || t.GetKind() == UnionKind || t.GetKind() == EnumKind {
		return true
	}
	return false
}

/*
 _     ___ ____ _____
| |   |_ _/ ___|_   _|
| |    | |\___ \ | |
| |___ | | ___) || |
|_____|___|____/ |_|
*/

/*
List in the GraphQL Type System
For creating a new List type, always use the `NewList` function

	NewList(SomeType)
*/
type List struct {
	Name        string
	Description string
	Wrapped     Type
}

/*
NewList returns a new List

In every case, you should use NewList function to create a new List, instead of creating on your own.
*/
func NewList(t Type) Type {
	return &List{
		Name:        "List",
		Description: "Built-in 'List' type",
		Wrapped:     t,
	}
}

// GetName returns the name of the type, "List"
func (l *List) GetName() string {
	return l.Name
}

// GetDescription shows the description of the type
func (l *List) GetDescription() string {
	return l.Description
}

// GetKind returns the kind of the type, ListKind
func (l *List) GetKind() TypeKind {
	return ListKind
}

// Unwrap the inner type from the List
func (l *List) Unwrap() Type {
	return l.Wrapped
}

// String implements the fmt.Stringer
func (l *List) String() string {
	return fmt.Sprintf("[%s]", l.Wrapped.String())
}

/*
 _   _  ___  _   _       _   _ _   _ _     _
| \ | |/ _ \| \ | |     | \ | | | | | |   | |
|  \| | | | |  \| |_____|  \| | | | | |   | |
| |\  | |_| | |\  |_____| |\  | |_| | |___| |___
|_| \_|\___/|_| \_|     |_| \_|\___/|_____|_____|
*/

// NonNull type from the GraphQL type system
type NonNull struct {
	Name        string
	Description string
	Wrapped     Type
}

/*
NewNonNull function helps create a new Non-Null type
It must be used, instead of creating a non null type "manually"
*/
func NewNonNull(t Type) Type {
	return &NonNull{
		Name:        "NonNull",
		Description: "Built-in 'NonNull' type",
		Wrapped:     t,
	}
}

// GetName returns the name of the type, "NonNull"
func (l *NonNull) GetName() string {
	return l.Name
}

// GetDescription returns the description of the NonNull type
func (l *NonNull) GetDescription() string {
	return l.Description
}

// GetKind returns the kind of the type, NonNullKind
func (l *NonNull) GetKind() TypeKind {
	return NonNullKind
}

// Unwrap the inner type of the NonNull type
func (l *NonNull) Unwrap() Type {
	return l.Wrapped
}

// String implements the fmt.Stringer
func (l *NonNull) String() string {
	return fmt.Sprintf("%s!", l.Wrapped.String())
}

/*
 ____   ____    _    _        _    ____  ____
/ ___| / ___|  / \  | |      / \  |  _ \/ ___|
\___ \| |     / _ \ | |     / _ \ | |_) \___ \
 ___) | |___ / ___ \| |___ / ___ \|  _ < ___) |
|____/ \____/_/   \_\_____/_/   \_\_| \_\____/
*/

// CoerceResultFunc coerces the result of a field resolve to the final format
type CoerceResultFunc func(interface{}) (interface{}, error)

// CoerceInputFunc coerces the input value to a type which will be used during field resolve
type CoerceInputFunc func(interface{}) (interface{}, error)

// ScalarAstValueValidator validates if the ast value is right
type ScalarAstValueValidator func(ast.Value) error

/*
Scalar types represent primitive leaf values in a GraphQL type system. GraphQL responses
take the form of a hierarchical tree; the leaves of this tree are typically
GraphQL Scalar types (but may also be Enum types or null values)
*/
type Scalar struct {
	Name             string
	Description      string
	Directives       TypeSystemDirectives
	CoerceResultFunc CoerceResultFunc
	CoerceInputFunc  CoerceInputFunc
	AstValidator     ScalarAstValueValidator
}

// GetName returns the name of the scalar
func (s *Scalar) GetName() string {
	return s.Name
}

// GetDescription shows the description of the scalar
func (s *Scalar) GetDescription() string {
	return s.Description
}

// GetKind returns the kind of the type, ScalarKind
func (s *Scalar) GetKind() TypeKind {
	return ScalarKind
}

// GetDirectives returns the directives added to the scalar
func (s *Scalar) GetDirectives() []TypeSystemDirective {
	return s.Directives
}

// CoerceResult coerces a result into the final type
func (s *Scalar) CoerceResult(i interface{}) (interface{}, error) {
	return s.CoerceResultFunc(i)
}

// CoerceInput coerces the input value to the type used in execution
func (s *Scalar) CoerceInput(i interface{}) (interface{}, error) {
	return s.CoerceInputFunc(i)
}

// String implements the fmt.Stringer
func (s *Scalar) String() string {
	return s.Name
}

/*
 _____ _   _ _   _ __  __ ____
| ____| \ | | | | |  \/  / ___|
|  _| |  \| | | | | |\/| \___ \
| |___| |\  | |_| | |  | |___) |
|_____|_| \_|\___/|_|  |_|____/
*/

/*
Enum types, like scalar types, also represent leaf values in a GraphQL type system.
However Enum types describe the set of possible values.
*/
type Enum struct {
	Name        string
	Description string
	Directives  TypeSystemDirectives
	Values      EnumValues
}

/*
GetDescription returns the description of the Enum
*/
func (e *Enum) GetDescription() string {
	return e.Description
}

/*
GetName returns the name of the Enum
*/
func (e *Enum) GetName() string {
	return e.Name
}

/*
GetDirectives returns all the directives set for the Enum
*/
func (e *Enum) GetDirectives() []TypeSystemDirective {
	return e.Directives
}

/*
GetKind returns the type kind, "Enum"
*/
func (e *Enum) GetKind() TypeKind {
	return EnumKind
}

/*
GetValues returns the values for the Enum
*/
func (e *Enum) GetValues() []*EnumValue {
	return e.Values
}

// String implements the fmt.Stringer
func (e *Enum) String() string {
	return e.Name
}

/*
EnumValues is an alias for a bunch of "EnumValue"s
*/
type EnumValues []*EnumValue

/*
EnumValue is one single value in an Enum
*/
type EnumValue struct {
	Name        string
	Description string
	Directives  TypeSystemDirectives
	Value       interface{}
}

/*
GetDescription returns the description of the value
*/
func (e EnumValue) GetDescription() string {
	return e.Description
}

/*
GetDirectives returns the directives set for the enum value
*/
func (e EnumValue) GetDirectives() []TypeSystemDirective {
	return e.Directives
}

/*
GetValue returns the actual value of the enum value
This value is used in the resolver functions, the enum coerces this value.
*/
func (e EnumValue) GetValue() interface{} {
	return e.Value
}

/*
IsDeprecated show if the deprecated directive is set for the enum value or not
*/
func (e EnumValue) IsDeprecated() bool {
	for _, d := range e.Directives {
		if _, ok := d.(*deprecated); ok {
			return true
		}
	}
	return false
}

/*
  ___  ____      _ _____ ____ _____ ____
 / _ \| __ )    | | ____/ ___|_   _/ ___|
| | | |  _ \ _  | |  _|| |     | | \___ \
| |_| | |_) | |_| | |__| |___  | |  ___) |
 \___/|____/ \___/|_____\____| |_| |____/
*/

/*
Object GraphQL queries are hierarchical and composed, describing a tree of information.
While Scalar types describe the leaf values of these hierarchical queries, Objects describe the intermediate levels

They represent a list of named fields, each of which yield a value of a specific type.

Example code:
	var RootQuery = &gql.Object{
		Name: "Query",
		Fields: gql.Fields{
			"hello": {
				Name: "hello",
				Type: gql.String,
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return "world", nil
				},
			},
		},
	}


All fields defined within an Object type must not have a name which begins with "__" (two underscores),
as this is used exclusively by GraphQL’s introspection system.
*/
type Object struct {
	Description string
	Name        string
	Implements  Interfaces
	Directives  TypeSystemDirectives
	Fields      Fields
}

/*
GetDescription returns the description of the object
*/
func (o *Object) GetDescription() string {
	return o.Description
}

/*
GetName returns the name of the object
*/
func (o *Object) GetName() string {
	return o.Name
}

/*
GetKind returns the type kind, "Object"
*/
func (o *Object) GetKind() TypeKind {
	return ObjectKind
}

/*
GetInterfaces returns all the interfaces that the object implements
*/
func (o *Object) GetInterfaces() []*Interface {
	return o.Implements
}

/*
GetDirectives returns all the directives that are used on the object
*/
func (o *Object) GetDirectives() []TypeSystemDirective {
	return o.Directives
}

/*
GetFields returns all the fields on the object
*/
func (o *Object) GetFields() map[string]*Field {
	return o.Fields
}

/*
AddFields lets you add fields to the object
*/
func (o *Object) AddField(name string, f *Field) {
	o.Fields[name] = f
}

/*
DoesImplement helps decide if the object implements an interface or not
*/
func (o *Object) DoesImplement(i *Interface) bool {
	for _, in := range o.Implements {
		// @Fontinalis: previously tried reflect.DeepEqual, but it's slow.., it's also not used during execution..
		if in.Name == i.Name {
			return true
		}
	}
	return false
}

// String implements the fmt.Stringer
func (o *Object) String() string {
	return o.Name
}

/*
 ___ _   _ _____ _____ ____  _____ _    ____ _____ ____
|_ _| \ | |_   _| ____|  _ \|  ___/ \  / ___| ____/ ___|
 | ||  \| | | | |  _| | |_) | |_ / _ \| |   |  _| \___ \
 | || |\  | | | | |___|  _ <|  _/ ___ \ |___| |___ ___) |
|___|_| \_| |_| |_____|_| \_\_|/_/   \_\____|_____|____/
*/

/*
Interfaces is an alias for more and more graphql interface
*/
type Interfaces []*Interface

/*
Interface represent a list of named fields and their arguments. GraphQL objects can then
implement these interfaces which requires that the object type will define all fields
defined by those interfaces.

Fields on a GraphQL interface have the same rules as fields on a GraphQL object; their type
can be Scalar, Object, Enum, Interface, or Union, or any wrapping type whose base type is one of those five.
*/
type Interface struct {
	Description  string
	Name         string
	Directives   TypeSystemDirectives
	Fields       Fields
	TypeResolver TypeResolver
}

/*
TypeResolver resolves an Interface's return Type
*/
type TypeResolver func(context.Context, interface{}) *Object

/*
GetDescription returns the description of the interface
*/
func (i *Interface) GetDescription() string {
	return i.Description
}

/*
GetName returns the name of the interface
*/
func (i *Interface) GetName() string {
	return i.Name
}

/*
GetKind returns the type kind, "Interface"
*/
func (i *Interface) GetKind() TypeKind {
	return InterfaceKind
}

/*
GetDirectives returns all the directives that are set to the interface
*/
func (i *Interface) GetDirectives() []TypeSystemDirective {
	return i.Directives
}

/*
GetFields returns all the fields that can be implemented on the interface
*/
func (i *Interface) GetFields() map[string]*Field {
	return i.Fields
}

/*
Resolve the return type of the interface
*/
func (i *Interface) Resolve(ctx context.Context, v interface{}) *Object {
	return i.TypeResolver(ctx, v)
}

// String implements the fmt.Stringer
func (i *Interface) String() string {
	return i.Name
}

/*
 _____ ___ _____ _     ____  ____
|  ___|_ _| ____| |   |  _ \/ ___|
| |_   | ||  _| | |   | | | \___ \
|  _|  | || |___| |___| |_| |___) |
|_|   |___|_____|_____|____/|____/
*/

/*
Fields is an alias for more Fields
*/
type Fields map[string]*Field

/*
Add a field to the Fields
*/
func (fs Fields) Add(name string, f *Field) {
	fs[name] = f
}

/*
Resolver function that can resolve a field using the Context from the executor
*/
type Resolver func(Context) (interface{}, error)

/*
Field for an object or interface

For Object, you can provide a resolver if you want to return custom data or have a something to do with the data.
For Interfaces, there's NO need for resolvers, since they will NOT be executed.
*/
type Field struct {
	Description string
	Arguments   Arguments
	Type        Type
	Directives  TypeSystemDirectives
	Resolver    Resolver
}

/*
GetDescription of the field
*/
func (f *Field) GetDescription() string {
	return f.Description
}

/*
GetArguments of the field
*/
func (f *Field) GetArguments() Arguments {
	return f.Arguments
}

/*
GetType returns the type of the field
*/
func (f *Field) GetType() Type {
	return f.Type
}

/*
GetDirectives returns the directives set for the field
*/
func (f *Field) GetDirectives() []TypeSystemDirective {
	return f.Directives
}

/*
IsDeprecated returns if the field is depricated
*/
func (f *Field) IsDeprecated() bool {
	for _, d := range f.Directives {
		if _, ok := d.(*deprecated); ok {
			return true
		}
	}
	return false
}

/*
 _   _ _   _ ___ ___  _   _ ____
| | | | \ | |_ _/ _ \| \ | / ___|
| | | |  \| || | | | |  \| \___ \
| |_| | |\  || | |_| | |\  |___) |
 \___/|_| \_|___\___/|_| \_|____/
*/

/*
Members is an alias for a list of types that are set for a Union
*/
type Members []Type

/*
Union represents an object that could be one of a list of graphql objects types,
but provides for no guaranteed fields between those types. They also differ from interfaces
in that Object types declare what interfaces they implement, but are not aware of what unions contain them
*/
type Union struct {
	Description  string
	Name         string
	Members      Members
	Directives   TypeSystemDirectives
	TypeResolver TypeResolver
}

/*
GetDescription show the description of the Union type
*/
func (u *Union) GetDescription() string {
	return u.Description
}

/*
GetName show the name of the Union type, "Union"
*/
func (u *Union) GetName() string {
	return u.Name
}

/*
GetKind returns the kind of the type, UnionType
*/
func (u *Union) GetKind() TypeKind {
	return UnionKind
}

/*
GetMembers returns the composite types contained by the union
*/
func (u *Union) GetMembers() []Type {
	return u.Members
}

/*
GetDirectives returns all the directives applied to the Union type
*/
func (u *Union) GetDirectives() []TypeSystemDirective {
	return u.Directives
}

/*
Resolve helps decide the executor which contained type to resolve
*/
func (u *Union) Resolve(ctx context.Context, v interface{}) *Object {
	return u.TypeResolver(ctx, v)
}

// String implements the fmt.Stringer
func (u *Union) String() string {
	return u.Name
}

/*
    _    ____   ____ _   _ __  __ _____ _   _ _____ ____
   / \  |  _ \ / ___| | | |  \/  | ____| \ | |_   _/ ___|
  / _ \ | |_) | |  _| | | | |\/| |  _| |  \| | | | \___ \
 / ___ \|  _ <| |_| | |_| | |  | | |___| |\  | | |  ___) |
/_/   \_\_| \_\\____|\___/|_|  |_|_____|_| \_| |_| |____/
*/

/*
Arguments for fields and directives
*/
type Arguments map[string]*Argument

func (args Arguments) String() (out string) {
	if args == nil {
		return
	}
	argsS := []string{}
	for name, arg := range args {
		if arg.Description != "" {
			out += `"` + arg.Description + `" `
		}
		out += name + `: `
		out += fmt.Sprint(arg.Type)
		if arg.IsDefaultValueSet() {
			out += fmt.Sprint(arg.DefaultValue)
		}
		argsS = append(argsS, out)
		out = ""
	}
	out = `(` + strings.Join(argsS, ", ") + `)`
	return
}

/*
Argument defines an argument for a field or a directive. Default value can be provided
in case it's not populated during a query. The type of the argument must be an input type.
*/
type Argument struct {
	Description  string
	Type         Type
	DefaultValue interface{}
}

/*
IsDefaultValueSet helps to know if the default value was set or not
*/
func (a *Argument) IsDefaultValueSet() bool {
	return reflect.ValueOf(a.DefaultValue).IsValid()
}

/*
 ___ _   _ ____  _   _ _____    ___  ____      _ _____ ____ _____ ____
|_ _| \ | |  _ \| | | |_   _|  / _ \| __ )    | | ____/ ___|_   _/ ___|
 | ||  \| | |_) | | | | | |   | | | |  _ \ _  | |  _|| |     | | \___ \
 | || |\  |  __/| |_| | | |   | |_| | |_) | |_| | |__| |___  | |  ___) |
|___|_| \_|_|    \___/  |_|    \___/|____/ \___/|_____\____| |_| |____/
*/

/*
InputObject is an input object from the GraphQL type system. It must have a name and
at least one input field set. Its fields can have default values if needed.
*/
type InputObject struct {
	Description string
	Name        string
	Directives  TypeSystemDirectives
	Fields      InputFields
}

/*
GetDescription returns the description of the input object and also there to
implement the Type interface
*/
func (o *InputObject) GetDescription() string {
	return o.Description
}

/*
GetName returns the name of the input object and also there to
implement the Type interface
*/
func (o *InputObject) GetName() string {
	return o.Name
}

/*
GetKind returns the kind of the type, "InputObject", and also there to
implement the Type interface
*/
func (o *InputObject) GetKind() TypeKind {
	return InputObjectKind
}

/*
GetDirectives returns the directives set for the input object
*/
func (o *InputObject) GetDirectives() []TypeSystemDirective {
	return o.Directives
}

/*
GetFields returns the fields of the input object
*/
func (o *InputObject) GetFields() map[string]*InputField {
	return o.Fields
}

// String implements the fmt.Stringer
func (o *InputObject) String() string {
	return o.Name
}

/*
InputField is a field for an InputObject. As an Argument, it can be used as an input too,
can have a default value and must have an input type.
*/
type InputField struct {
	Description  string
	Type         Type
	DefaultValue interface{}
	Directives   TypeSystemDirectives
}

/*
IsDefaultValueSet helps to know if the default value was set or not
*/
func (o *InputField) IsDefaultValueSet() bool {
	return reflect.ValueOf(o.DefaultValue).IsValid()
}

/*
InputFields is just an alias for a bunch of "InputField"s
*/
type InputFields map[string]*InputField
