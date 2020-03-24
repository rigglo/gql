package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rigglo/gql/pkg/gql"
)

func init() {
	DogType.Implements = gql.Interfaces{PetInterface}
	CatType.Implements = gql.Interfaces{PetInterface}
}

func main() {
	res := PetStore.Exec(context.Background(), gql.Params{
		OperationName: "getDogName",
		Query: `
		fragment fragmentOne on Dog {
			name
		  }

		query getDogName {
			dog {
				__typename
			  	name
			}
			cat {
				__typename
				...fragmentOne
			}
			pet {
				__typename
				name
			}
			catOrDog {
				__typename
				...fragmentOne
			}
		  }
		`,
		Variables: map[string]interface{}{},
	})
	bs, _ := json.MarshalIndent(res, "", "  ")
	log.Printf("%s\n", string(bs))
}

type Dog struct {
	Name            string `json:"name"`
	Nickname        string `json:"nickname"`
	BarkVolume      int    `json:"barkVolume"`
	DoesKnowCommand bool   `json:"-"`
	IsHousetrained  bool   `json:"-"`
	Owner           Human  `json:"owner"`
}

type Cat struct {
	Name            string `json:"name"`
	Nickname        string `json:"nickname"`
	DoesKnowCommand bool   `json:"-"`
	MeowVolume      int    `json:"meowVolume"`
}

type Human struct {
	Name string `json:"name"`
}

var (
	Catcy = Cat{
		Name:            "Catcy",
		Nickname:        "cc",
		MeowVolume:      12,
		DoesKnowCommand: true,
	}

	Doggo = Dog{
		Name:            "Doggo",
		Nickname:        "doggo",
		BarkVolume:      42,
		DoesKnowCommand: true,
		IsHousetrained:  true,
		Owner: Human{
			Name: "John Doe",
		},
	}

	PetStore = &gql.Schema{
		Query: Query,
	}

	Query = &gql.Object{
		Name: "Query",
		Fields: gql.Fields{
			&gql.Field{
				Name: "dog",
				Type: DogType,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return Doggo, nil
				},
			},
			&gql.Field{
				Name: "cat",
				Type: CatType,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return Catcy, nil
				},
			},
			&gql.Field{
				Name: "pet",
				Type: gql.NewList(PetInterface),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return []interface{}{Catcy, Doggo}, nil
				},
			},
			&gql.Field{
				Name: "catOrDog",
				Type: gql.NewList(CatOrDogUnion),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return []interface{}{Catcy, Doggo}, nil
				},
			},
		},
	}

	DogCommandEnum = &gql.Enum{
		Name: "DogCommand",
		Values: gql.EnumValues{
			"SIT": &gql.EnumValue{
				Value:       "SIT",
				Description: "SIT",
			},
			"DOWN": &gql.EnumValue{
				Value:       "DOWN",
				Description: "DOWN",
			},
			"HEEL": &gql.EnumValue{
				Value:       "HEEL",
				Description: "HEEL",
			},
		},
	}

	DogType = &gql.Object{
		Name: "Dog",
		Fields: gql.Fields{
			&gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
			},
			&gql.Field{
				Name: "nickname",
				Type: gql.String,
			},
			&gql.Field{
				Name: "barkVolume",
				Type: gql.Int,
			},
			&gql.Field{
				Name: "doesKnowCommand",
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "dogCommand",
						Type: gql.NewNonNull(DogCommandEnum),
					},
				},
				Type: gql.NewNonNull(gql.Boolean),
			},
			&gql.Field{
				Name: "isHousetrained",
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "atOtherHomes",
						Type: gql.Boolean,
					},
				},
				Type: gql.NewNonNull(gql.Boolean),
			},
			&gql.Field{
				Name: "owner",
				Type: HumanType,
			},
		},
	}

	SentientInterface = &gql.Interface{
		Name: "Sentient",
		Fields: gql.Fields{
			&gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
			},
		},
	}

	PetInterface = &gql.Interface{
		Name: "Pet",
		Fields: gql.Fields{
			&gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
			},
		},
		TypeResolver: func(ctx context.Context, r interface{}) gql.ObjectType {
			if _, ok := r.(Dog); ok {
				return DogType
			}
			return CatType
		},
	}

	AlienType = &gql.Object{
		Name:       "Alien",
		Implements: gql.Interfaces{SentientInterface},
		Fields: gql.Fields{
			&gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
			},
			&gql.Field{
				Name: "homePlanet",
				Type: gql.String,
			},
		},
	}

	HumanType = &gql.Object{
		Name:       "Human",
		Implements: gql.Interfaces{SentientInterface},
		Fields: gql.Fields{
			&gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
			},
		},
	}

	CatCommandEnum = &gql.Enum{
		Name: "CatCommand",
		Values: gql.EnumValues{
			"JUMP": &gql.EnumValue{
				Value:       "JUMP",
				Description: "JUMP",
			},
		},
	}

	CatType = &gql.Object{
		Name: "Cat",
		Fields: gql.Fields{
			&gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
			},
			&gql.Field{
				Name: "nickname",
				Type: gql.String,
			},
			&gql.Field{
				Name: "meowVolume",
				Type: gql.Int,
			},
			&gql.Field{
				Name: "doesKnowCommand",
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "catCommand",
						Type: gql.NewNonNull(CatCommandEnum),
					},
				},
				Type: gql.NewNonNull(gql.Boolean),
			},
		},
	}

	CatOrDogUnion = &gql.Union{
		Name:    "CatOrDog",
		Members: gql.Members{CatType, DogType},
		TypeResolver: func(ctx context.Context, r interface{}) gql.ObjectType {
			if _, ok := r.(Dog); ok {
				return DogType
			}
			return CatType
		},
	}

	DogOrHumanUnion = &gql.Union{
		Name:    "DogOrHuman",
		Members: gql.Members{DogType, HumanType},
	}

	HumanOrAlienUnion = &gql.Union{
		Name:    "HumanOrAlien",
		Members: gql.Members{HumanType, AlienType},
	}
)
