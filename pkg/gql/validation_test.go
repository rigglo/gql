package gql_test

import (
	"context"
	"testing"

	"github.com/rigglo/gql/pkg/testutils/pets"

	"github.com/rigglo/gql/pkg/gql"
)

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
