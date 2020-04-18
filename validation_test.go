package gql_test

import (
	"context"
	"testing"

	"github.com/rigglo/gql"
	"github.com/rigglo/gql/pkg/testutil"
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "interfaceFieldSelection",
			args: args{
				params: gql.Params{
					Query: `
					query {
						pet {
							...interfaceFieldSelection
						}
					}
					fragment interfaceFieldSelection on Pet {
						name
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "definedOnImplementorsButNotInterface",
			args: args{
				params: gql.Params{
					Query: `
					query {
						pet {
							...definedOnImplementorsButNotInterface
						}
					}
					fragment definedOnImplementorsButNotInterface on Pet {
						nickname
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "inDirectFieldSelectionOnUnion",
			args: args{
				params: gql.Params{
					Query: `
					query {
						catOrDog {
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
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "directFieldSelectionOnUnion",
			args: args{
				params: gql.Params{
					Query: `
					query {
						catOrDog {
							...directFieldSelectionOnUnion
						}
					}
					fragment directFieldSelectionOnUnion on CatOrDog {
						name
						barkVolume
					}
					`,
				},
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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
				schema: testutil.Schema,
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

func Test_FragmentNameUniqueness(t *testing.T) {
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
			name: "uniqueFragments",
			args: args{
				params: gql.Params{
					Query: `
					{
						dog {
							...fragmentOne
							...fragmentTwo
						}
					}
					  
					fragment fragmentOne on Dog {
						name
					}
					  
					fragment fragmentTwo on Dog {
						owner {
						  	name
						}
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "nonUniqueFragments",
			args: args{
				params: gql.Params{
					Query: `
					{
						dog {
						  	...fragmentOne
						}
					}
					  
					fragment fragmentOne on Dog {
						name
					}
					  
					fragment fragmentOne on Dog {
						owner {
						  	name
						}
					}
					`,
				},
				schema: testutil.Schema,
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

func Test_FragmentsOnCompositeTypes(t *testing.T) {
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
			name: "fragOnObject",
			args: args{
				params: gql.Params{
					Query: `
					{
						dog {
							...fragOnObject
						}
					}
					  
					fragment fragOnObject on Dog {
						name
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "fragOnInterface",
			args: args{
				params: gql.Params{
					Query: `
					{
						dog {
							...fragOnInterface
						}
					}
					  
					fragment fragOnInterface on Pet {
						name
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "fragOnUnion",
			args: args{
				params: gql.Params{
					Query: `
					{
						dog {
							...fragOnUnion
						}
					}
					  
					fragment fragOnUnion on CatOrDog {
						... on Dog {
						  	name
						}
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "fragOnUnion",
			args: args{
				params: gql.Params{
					Query: `
					{
						dog {
							...fragOnUnion
						}
					}
					  
					fragment fragOnUnion on CatOrDog {
						... on Dog {
						  	name
						}
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "inlineFragOnScalar",
			args: args{
				params: gql.Params{
					Query: `
					{
						dog {
							...inlineFragOnScalar
						}
					}
					  
					fragment inlineFragOnScalar on Dog {
						... on Boolean {
						  somethingElse
						}
					  }
					`,
				},
				schema: testutil.Schema,
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

func Test_FragmensMustBeUsed(t *testing.T) {
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
			name: "unusedFragment",
			args: args{
				params: gql.Params{
					Query: `
					fragment nameFragment on Dog { # unused
						name
					}
					  
					{
						dog {
						  	name
						}
					}
					`,
				},
				schema: testutil.Schema,
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

func Test_FragmentSpreadTargetDefined(t *testing.T) {
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
			name: "undefinedFragment",
			args: args{
				params: gql.Params{
					Query: `
					{
						dog {
						  	...undefinedFragment
						}
					}
					`,
				},
				schema: testutil.Schema,
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

func Test_FragmentSpreadsMustNotFormCycles(t *testing.T) {
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
			name: "nameFragment",
			args: args{
				params: gql.Params{
					Query: `
					{
						dog {
						  ...nameFragment
						}
					}
					  
					fragment nameFragment on Dog {
						name
						...barkVolumeFragment
					}
					  
					fragment barkVolumeFragment on Dog {
						barkVolume
						...nameFragment
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "dogFragment",
			args: args{
				params: gql.Params{
					Query: `
					{
						dog {
						  ...dogFragment
						}
					}
					  
					fragment dogFragment on Dog {
						name
						owner {
						  	...ownerFragment
						}
					}
					  
					fragment ownerFragment on Dog {
						name
						pets {
						  	...dogFragment
						}
					}
					`,
				},
				schema: testutil.Schema,
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

func Test_FragmentSpreadIsPossible(t *testing.T) {
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
			name: "dogFragment",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...dogFragment
						}
					}
					fragment dogFragment on Dog {
						... on Dog {
						  	barkVolume
						}
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "catInDogFragmentInvalid",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...catInDogFragmentInvalid
						}
					}
					fragment catInDogFragmentInvalid on Dog {
						... on Cat {
						  	meowVolume
						}
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "interfaceWithinObjectFragment",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...interfaceWithinObjectFragment
						}
					}
					fragment petNameFragment on Pet {
						name
					}
					fragment interfaceWithinObjectFragment on Dog {
						...petNameFragment
					  }
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "unionWithObjectFragment",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...unionWithObjectFragment
						}
					}

					fragment catOrDogNameFragment on CatOrDog {
						... on Cat {
						  	meowVolume
						}
					}
					  
					fragment unionWithObjectFragment on Dog {
						...catOrDogNameFragment
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "humanOrAlienFragment",
			args: args{
				params: gql.Params{
					Query: `
					query {
						humanOrAlien {
							...humanOrAlienFragment
						}
					}
					  
					fragment humanOrAlienFragment on HumanOrAlien {
						... on Cat {
						  	meowVolume
						}
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "sentientFragment",
			args: args{
				params: gql.Params{
					Query: `
					query {
						sentient {
							...sentientFragment
						}
					}

					fragment sentientFragment on Sentient {
						... on Dog {
						  	barkVolume
						}
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "unionWithInterface",
			args: args{
				params: gql.Params{
					Query: `
					query {
						pet {
							...unionWithInterface
						}
					}

					fragment unionWithInterface on Pet {
						...dogOrHumanFragment
					}
					  
					fragment dogOrHumanFragment on DogOrHuman {
						... on Dog {
						  	barkVolume
						}
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "nonIntersectingInterfaces",
			args: args{
				params: gql.Params{
					Query: `
					query {
						pet {
							...nonIntersectingInterfaces
						}
					}

					fragment nonIntersectingInterfaces on Pet {
						...sentientFragment
					}
					  
					fragment sentientFragment on Sentient {
						name
					}
					`,
				},
				schema: testutil.Schema,
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

func Test_Directives(t *testing.T) {
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
			name: "invalidSkipLocation",
			args: args{
				params: gql.Params{
					Query: `
					query @skip(if: $foo) {
						dog {
							name
						}
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "nonUniquePerLocation",
			args: args{
				params: gql.Params{
					Query: `
					query ($foo: Boolean = true, $bar: Boolean = false) {
						dog @skip(if: $foo) @skip(if: $bar) {
							name
						}
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "validDirectivesOnGoodLoc",
			args: args{
				params: gql.Params{
					Query: `
					query ($foo: Boolean = true, $bar: Boolean = false) {
						dog @skip(if: $foo) {
							name
						}
						dog @skip(if: $bar) {
							barkVolume
						}
					}
					`,
				},
				schema: testutil.Schema,
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

func Test_ArgumentNames(t *testing.T) {
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
			name: "argOnRequiredArg",
			args: args{
				params: gql.Params{
					Query: `
						query {
							dog {
								...argOnRequiredArg
							}
						}
						fragment argOnRequiredArg on Dog {
							doesKnowCommand(dogCommand: SIT)
						}
						`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "argOnOptional",
			args: args{
				params: gql.Params{
					Query: `
						query {
							dog {
								...argOnOptional
							}
						}
						fragment argOnOptional on Dog {
							isHousetrained(atOtherHomes: true) @include(if: true)
						}
						`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "invalidArgName",
			args: args{
				params: gql.Params{
					Query: `
						query {
							dog {
								...invalidArgName
							}
						}
						fragment invalidArgName on Dog {
							doesKnowCommand(command: CLEAN_UP_HOUSE)
						}
						`,
				},
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "invalidArgName",
			args: args{
				params: gql.Params{
					Query: `
					query {
						dog {
							...invalidArgName
						}
					}
					fragment invalidArgName on Dog {
						isHousetrained(atOtherHomes: true) @include(unless: false)
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "multipleArgs",
			args: args{
				params: gql.Params{
					Query: `
					query {
						arguments {
							...multipleArgs
						}
					}
					fragment multipleArgs on Arguments {
						multipleReqs(x: 1, y: 2)
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "multipleArgsReverseOrder",
			args: args{
				params: gql.Params{
					Query: `
					query {
						arguments {
							...multipleArgsReverseOrder
						}
					}
					fragment multipleArgsReverseOrder on Arguments {
						multipleReqs(y: 1, x: 2)
					}
					`,
				},
				schema: testutil.Schema,
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

func Test_ArgumentUniqueness(t *testing.T) {
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
			name: "goodBooleanArg",
			args: args{
				params: gql.Params{
					Query: `
					query {
						arguments {
							...goodBooleanArg
						}
					}
					fragment goodBooleanArg on Arguments {
						booleanArgField(booleanArg: true)
					}					  
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "goodNonNullArg",
			args: args{
				params: gql.Params{
					Query: `
					query {
						arguments {
							...goodNonNullArg
						}
					}
					fragment goodNonNullArg on Arguments {
						nonNullBooleanArgField(nonNullBooleanArg: true)
					}				  
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "goodBooleanArgDefault",
			args: args{
				params: gql.Params{
					Query: `
					query {
						arguments {
							...goodBooleanArgDefault
						}
					}
					fragment goodBooleanArgDefault on Arguments {
						booleanArgField
					}					  
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "missingRequiredArg",
			args: args{
				params: gql.Params{
					Query: `
					query {
						arguments {
							...missingRequiredArg
						}
					}
					fragment missingRequiredArg on Arguments {
						nonNullBooleanArgField
					}					  
					`,
				},
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "missingRequiredArg",
			args: args{
				params: gql.Params{
					Query: `
					query {
						arguments {
							...missingRequiredArg
						}
					}
					fragment missingRequiredArg on Arguments {
						nonNullBooleanArgField(nonNullBooleanArg: null)
					}					  
					`,
				},
				schema: testutil.Schema,
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

func Test_ValuesOfCorrectType(t *testing.T) {
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
			name: "goodBooleanArg",
			args: args{
				params: gql.Params{
					Query: `
					query {
						arguments {
							...goodBooleanArg
						}
					}
					fragment goodBooleanArg on Arguments {
						booleanArgField(booleanArg: true)
					}					  
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "coercedIntIntoFloatArg",
			args: args{
				params: gql.Params{
					Query: `
					query {
						arguments {
							...coercedIntIntoFloatArg
						}
					}
					fragment coercedIntIntoFloatArg on Arguments {
						# Note: The input coercion rules for Float allow Int literals.
						floatArgField(floatArg: 123)
					}									
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "goodComplexDefaultValue",
			args: args{
				params: gql.Params{
					Query: `
					query goodComplexDefaultValue($search: ComplexInput = { name: "Fido" }) {
						findDog(complex: $search) {
							nickname
						}
					}				  
					`,
				},
				schema: testutil.Schema,
			},
			valid: true,
		},
		{
			name: "stringIntoInt",
			args: args{
				params: gql.Params{
					Query: `
					query {
						arguments {
							...stringIntoInt
						}
					}
					fragment stringIntoInt on Arguments {
						intArgField(intArg: "123")
					}					  
					`,
				},
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "badComplexValue",
			args: args{
				params: gql.Params{
					Query: `
					query {
						arguments {
							...badComplexValue
						}
					}
					query badComplexValue {
						findDog(complex: { name: 123 })
					}					  
					`,
				},
				schema: testutil.Schema,
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

func Test_VariableUniqueness(t *testing.T) {
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
			name: "houseTrainedQuery",
			args: args{
				params: gql.Params{
					Query: `
					query houseTrainedQuery($atOtherHomes: Boolean, $atOtherHomes: Boolean) {
						dog {
						  	isHousetrained(atOtherHomes: $atOtherHomes)
						}
					}
					`,
				},
				schema: testutil.Schema,
			},
			valid: false,
		},
		{
			name: "HouseTrainedFragment",
			args: args{
				params: gql.Params{
					Query: `
					query A($atOtherHomes: Boolean) {
						...HouseTrainedFragment
					}
					  
					query B($atOtherHomes: Boolean) {
						...HouseTrainedFragment
					}
					  
					fragment HouseTrainedFragment on Query {
						dog {
						  	isHousetrained(atOtherHomes: $atOtherHomes)
						}
					}
					`,
					OperationName: "A",
				},
				schema: testutil.Schema,
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
