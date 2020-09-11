package federation

import "github.com/rigglo/gql"

// Extends directive
// Apollo Federation supports using an @extends directive in place of extend type to annotate type references
func Extends() gql.TypeSystemDirective {
	return &fExtendsDirective{}
}

type fExtendsDirective struct{}

func (d *fExtendsDirective) GetName() string {
	return "extends"
}

func (d *fExtendsDirective) GetDescription() string {
	return "Apollo Federation supports using an @extends directive in place of extend type to annotate type references"
}

func (d *fExtendsDirective) GetArguments() gql.Arguments {
	return gql.Arguments{}
}

func (d *fExtendsDirective) GetLocations() []gql.DirectiveLocation {
	return []gql.DirectiveLocation{
		gql.ObjectLoc,
		gql.InterfaceLoc,
	}
}

func (d *fExtendsDirective) GetValues() map[string]interface{} {
	return map[string]interface{}{}
}

// Key directive is used to indicate a combination of fields that can be used to uniquely identify and fetch an object or interface
func Key(fields string) gql.TypeSystemDirective {
	return &keyDirective{fields}
}

type keyDirective struct {
	fields string
}

func (d *keyDirective) GetName() string {
	return "key"
}

func (d *keyDirective) GetDescription() string {
	return "key directive is used to indicate a combination of fields that can be used to uniquely identify and fetch an object or interface"
}

func (d *keyDirective) GetArguments() gql.Arguments {
	return gql.Arguments{
		"fields": &gql.Argument{
			Type: gql.NewNonNull(fieldSetScalar),
		},
	}
}

func (d *keyDirective) GetLocations() []gql.DirectiveLocation {
	return []gql.DirectiveLocation{
		gql.ObjectLoc,
		gql.InterfaceLoc,
	}
}

func (d *keyDirective) GetValues() map[string]interface{} {
	return map[string]interface{}{
		"fields": d.fields,
	}
}

// External directive is used to mark a field as owned by another service
func External() gql.TypeSystemDirective {
	return &externalDirective{}
}

type externalDirective struct{}

func (d *externalDirective) GetName() string {
	return "external"
}

func (d *externalDirective) GetDescription() string {
	return "@external directive is used to mark a field as owned by another service"
}

func (d *externalDirective) GetArguments() gql.Arguments {
	return gql.Arguments{}
}

func (d *externalDirective) GetLocations() []gql.DirectiveLocation {
	return []gql.DirectiveLocation{
		gql.FieldDefinitionLoc,
	}
}

func (d *externalDirective) GetValues() map[string]interface{} {
	return map[string]interface{}{}
}

// Requires directive is used to annotate the required input fieldset from a base type for a resolver. It is used to develop a query plan where the required fields may not be needed by the client, but the service may need additional information from other services
func Requires(fields string) gql.TypeSystemDirective {
	return &requiresDirective{fields}
}

type requiresDirective struct {
	fields string
}

func (d *requiresDirective) GetName() string {
	return "requires"
}

func (d *requiresDirective) GetDescription() string {
	return "@requires directive is used to annotate the required input fieldset from a base type for a resolver. It is used to develop a query plan where the required fields may not be needed by the client, but the service may need additional information from other services"
}

func (d *requiresDirective) GetArguments() gql.Arguments {
	return gql.Arguments{
		"fields": &gql.Argument{
			Type: gql.NewNonNull(fieldSetScalar),
		},
	}
}

func (d *requiresDirective) GetLocations() []gql.DirectiveLocation {
	return []gql.DirectiveLocation{
		gql.FieldDefinitionLoc,
	}
}

func (d *requiresDirective) GetValues() map[string]interface{} {
	return map[string]interface{}{
		"fields": d.fields,
	}
}

// Provides directive is used to annotate the expected returned fieldset from a field on a base type that is guaranteed to be selectable by the gateway
func Provides(fields string) gql.TypeSystemDirective {
	return &providesDirective{fields}
}

type providesDirective struct {
	fields string
}

func (d *providesDirective) GetName() string {
	return "provides"
}

func (d *providesDirective) GetDescription() string {
	return "@provides directive is used to annotate the expected returned fieldset from a field on a base type that is guaranteed to be selectable by the gateway"
}

func (d *providesDirective) GetArguments() gql.Arguments {
	return gql.Arguments{
		"fields": &gql.Argument{
			Type: gql.NewNonNull(fieldSetScalar),
		},
	}
}

func (d *providesDirective) GetLocations() []gql.DirectiveLocation {
	return []gql.DirectiveLocation{
		gql.FieldDefinitionLoc,
	}
}

func (d *providesDirective) GetValues() map[string]interface{} {
	return map[string]interface{}{
		"fields": d.fields,
	}
}
