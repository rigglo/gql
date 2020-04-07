package gql_test

import (
	"context"
	"testing"

	"github.com/rigglo/gql/pkg/testutils/pets"

	"github.com/rigglo/gql/pkg/gql"
)

/*
func Test_SampleTest(t *testing.T) {
	ctx := context.Background()
	type args struct {
		params gql.Params
		schema *gql.Schema
	}
	tests := []struct {
		name  string
		args  args
		valid bool
	}{
		{
			name: "mergeIdenticalFields",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...mergeIdenticalFields
						}
					}
					fragment mergeIdenticalFields on Dog {
						name
						name
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gql.Execute(ctx, tt.args.schema, tt.args.params)
			if (tt.valid && len(r.Errors) != 0) || (!tt.valid && len(r.Errors) == 0) {
				t.Fatalf("%+v, %v", r.Errors, len(r.Errors))
			}
		})
	}
}
*/

func Test_FieldSelectionsOnObjectInterfacesAndUnions(t *testing.T) {
	ctx := context.Background()
	type args struct {
		params gql.Params
		schema *gql.Schema
	}
	tests := []struct {
		name  string
		args  args
		valid bool
	}{
		{
			name: "fieldNotDefined",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...fieldNotDefined
						}
					}
					fragment fieldNotDefined on Dog {
						meowVolume
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
		{
			name: "aliasedLyingFieldTargetNotDefined",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...aliasedLyingFieldTargetNotDefined
						}
					}
					fragment aliasedLyingFieldTargetNotDefined on Dog {
						barkVolume: kawVolume
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
		{
			name: "interfaceFieldSelection",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...interfaceFieldSelection
						}
					}
					fragment interfaceFieldSelection on Pet {
						name
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: true,
		},
		{
			name: "definedOnImplementorsButNotInterface",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...definedOnImplementorsButNotInterface
						}
					}
					fragment definedOnImplementorsButNotInterface on Pet {
						nickname
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
		{
			name: "inDirectFieldSelectionOnUnion",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...inDirectFieldSelectionOnUnion
						}
					}
					fragment inDirectFieldSelectionOnUnion on CatOrDog {
						__typename
						... on Pet {
						  	name
						}
						... on Dog {
						  	barkVolume
						}
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: true,
		},
		{
			name: "directFieldSelectionOnUnion",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...directFieldSelectionOnUnion
						}
					}
					fragment directFieldSelectionOnUnion on CatOrDog {
						name
						barkVolume
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gql.Execute(ctx, tt.args.schema, tt.args.params)
			if (tt.valid && len(r.Errors) != 0) || (!tt.valid && len(r.Errors) == 0) {
				t.Fatalf("%+v, %v", r.Errors, len(r.Errors))
			}
		})
	}
}

func Test_FieldSelectionMerging(t *testing.T) {
	ctx := context.Background()
	type args struct {
		params gql.Params
		schema *gql.Schema
	}
	tests := []struct {
		name  string
		args  args
		valid bool
	}{
		{
			name: "mergeIdenticalFields",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...mergeIdenticalFields
						}
					}
					fragment mergeIdenticalFields on Dog {
						name
						name
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: true,
		},
		{
			name: "mergeIdenticalAliasesAndFields",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...mergeIdenticalAliasesAndFields
						}
					}
					fragment mergeIdenticalAliasesAndFields on Dog {
						otherName: name
						otherName: name
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: true,
		},
		{
			name: "conflictingBecauseAlias",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...conflictingBecauseAlias
						}
					}
					fragment conflictingBecauseAlias on Dog {
						name: nickname
						name
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
		{
			name: "mergeIdenticalFieldsWithIdenticalArgs",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...mergeIdenticalFieldsWithIdenticalArgs
						}
					}
					fragment mergeIdenticalFieldsWithIdenticalArgs on Dog {
						doesKnowCommand(dogCommand: SIT)
						doesKnowCommand(dogCommand: SIT)
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: true,
		},
		{
			name: "conflictingArgsOnValues",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...conflictingArgsOnValues
						}
					}
					fragment conflictingArgsOnValues on Dog {
						doesKnowCommand(dogCommand: SIT)
						doesKnowCommand(dogCommand: HEEL)
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
		{
			name: "conflictingArgsValueAndVar",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...conflictingArgsValueAndVar
						}
					}
					fragment conflictingArgsValueAndVar on Dog {
						doesKnowCommand(dogCommand: SIT)
						doesKnowCommand(dogCommand: $dogCommand)
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
		{
			name: "conflictingArgsWithVars",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...conflictingArgsWithVars
						}
					}
					fragment conflictingArgsWithVars on Dog {
						doesKnowCommand(dogCommand: $varOne)
						doesKnowCommand(dogCommand: $varTwo)
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
		{
			name: "differingArgs",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...differingArgs
						}
					}
					fragment differingArgs on Dog {
						doesKnowCommand(dogCommand: SIT)
						doesKnowCommand
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
		{
			name: "safeDifferingFields",
			args: args{
				params: gql.Params{
					Query: `
					query {
						pet {
							...safeDifferingFields
						}
					}
					fragment safeDifferingFields on Pet {
						... on Dog {
						  	volume: barkVolume
						}
						... on Cat {
						  	volume: meowVolume
						}
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: true,
		},
		{
			name: "conflictingDifferingResponses",
			args: args{
				params: gql.Params{
					Query: `
					query {
						pet {
							...conflictingDifferingResponses
						}
					}
					fragment conflictingDifferingResponses on Pet {
						... on Dog {
						  	someValue: nickname
						}
						... on Cat {
						  	someValue: meowVolume
						}
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
		{
			name: "safeDifferingArgs",
			args: args{
				params: gql.Params{
					Query: `
					query {
						pet {
							...safeDifferingArgs
						}
					}
					fragment safeDifferingArgs on Pet {
						... on Dog {
						  	doesKnowCommand(dogCommand: SIT)
						}
						... on Cat {
						  	doesKnowCommand(catCommand: JUMP)
						}
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: true,
		},
		{
			name: "conflictingDifferingResponses",
			args: args{
				params: gql.Params{
					Query: `
					query {
						pet {
							...conflictingDifferingResponses
						}
					}
					fragment conflictingDifferingResponses on Pet {
						... on Dog {
						  someValue: nickname
						}
						... on Cat {
						  someValue: meowVolume
						}
					  }
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gql.Execute(ctx, tt.args.schema, tt.args.params)
			if (tt.valid && len(r.Errors) != 0) || (!tt.valid && len(r.Errors) == 0) {
				t.Fatalf("%+v, %v", r.Errors, len(r.Errors))
			}
		})
	}
}

func Test_LeadFieldSelections(t *testing.T) {
	ctx := context.Background()
	type args struct {
		params gql.Params
		schema *gql.Schema
	}
	tests := []struct {
		name  string
		args  args
		valid bool
	}{
		{
			name: "scalarSelection",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...scalarSelection
						}
					}
					fragment scalarSelection on Dog {
						barkVolume
					  }
					`,
				},
				schema: pets.PetStore,
			},
			valid: true,
		},
		{
			name: "scalarSelectionsNotAllowedOnInt",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...scalarSelectionsNotAllowedOnInt
						}
					}
					fragment scalarSelectionsNotAllowedOnInt on Dog {
						barkVolume {
						  	sinceWhen
						}
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
		{
			name: "directQueryOnObjectWithoutSubFields",
			args: args{
				params: gql.Params{
					Query: `
					query directQueryOnObjectWithoutSubFields {
						human
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
		{
			name: "directQueryOnInterfaceWithoutSubFields",
			args: args{
				params: gql.Params{
					Query: `
					query directQueryOnInterfaceWithoutSubFields {
						pet
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
		{
			name: "directQueryOnUnionWithoutSubFields",
			args: args{
				params: gql.Params{
					Query: `
					query directQueryOnUnionWithoutSubFields {
						catOrDog
					}
					`,
				},
				schema: pets.PetStore,
			},
			valid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gql.Execute(ctx, tt.args.schema, tt.args.params)
			if (tt.valid && len(r.Errors) != 0) || (!tt.valid && len(r.Errors) == 0) {
				t.Fatalf("%+v, %v", r.Errors, len(r.Errors))
			}
		})
	}
}
