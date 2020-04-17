package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/rigglo/gql/pkg/gql"
	"github.com/rigglo/gql/pkg/handler"
)

func init() {
	DogType.Implements = gql.Interfaces{PetInterface}
	CatType.Implements = gql.Interfaces{PetInterface}
}

func main() {
	h := handler.New(handler.Config{
		Executor: gql.DefaultExecutor(PetStore),
		GraphiQL: true,
	})
	http.Handle("/graphql", h)
	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Println(err)
	}
}

/*
func main() {
	res := gql.Execute(context.Background(), PetStore, gql.Params{
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
			someErr {
				__typename
			}
		}
		`,
		Variables: "",
	})
	bs, _ := json.MarshalIndent(res, "", "  ")
	log.Printf("%s\n", string(bs))
} */

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
			"dog": &gql.Field{
				Name: "dog",
				Type: DogType,
				Resolver: func(ctx gql.Context) (interface{}, error) {
					log.Println("start dog", time.Now())
					time.Sleep(2 * time.Second)
					log.Println("finish dog", time.Now())
					return Doggo, nil
				},
			},
			"cat": &gql.Field{
				Name: "cat",
				Type: CatType,
				Resolver: func(ctx gql.Context) (interface{}, error) {
					log.Println("start cat", time.Now())
					time.Sleep(4 * time.Second)
					log.Println("finish cat", time.Now())
					return Catcy, nil
				},
			},
			"pet": &gql.Field{
				Name: "pet",
				Type: gql.NewList(PetInterface),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return []interface{}{Catcy, Doggo}, nil
				},
			},
			"catOrDog": &gql.Field{
				Name: "catOrDog",
				Type: gql.NewList(CatOrDogUnion),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return []interface{}{Catcy, Doggo}, nil
				},
			},
			"someErr": &gql.Field{
				Name: "someErr",
				Type: gql.NewList(CatOrDogUnion),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return nil, gql.NewError("asdasdasd", map[string]interface{}{"code": "SOME_ERR_CODE"})
				},
			},
		},
	}

	DogCommandEnum = &gql.Enum{
		Name: "DogCommand",
		Values: gql.EnumValues{
			&gql.EnumValue{
				Name:        "SIT",
				Value:       "SIT",
				Description: "SIT command, usually the dog would sit to this command",
			},
			&gql.EnumValue{
				Name:        "DOWN",
				Value:       "DOWN",
				Description: "DOWN",
			},
			&gql.EnumValue{
				Name:        "HEEL",
				Value:       "HEEL",
				Description: "HEEL",
			},
		},
	}

	DogType = &gql.Object{
		Name: "Dog",
		Fields: gql.Fields{
			"name": &gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
			},
			"nickname": &gql.Field{
				Name: "nickname",
				Type: gql.String,
			},
			"barkVolume": &gql.Field{
				Name: "barkVolume",
				Type: gql.Int,
			},
			"doesKnowCommand": &gql.Field{
				Name: "doesKnowCommand",
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "dogCommand",
						Type: gql.NewNonNull(DogCommandEnum),
					},
				},
				Type: gql.NewNonNull(gql.Boolean),
			},
			"isHousetrained": &gql.Field{
				Name: "isHousetrained",
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "atOtherHomes",
						Type: gql.Boolean,
					},
				},
				Type: gql.NewNonNull(gql.Boolean),
			},
			"owner": &gql.Field{
				Name: "owner",
				Type: HumanType,
			},
		},
	}

	SentientInterface = &gql.Interface{
		Name: "Sentient",
		Fields: gql.Fields{
			"name": &gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
			},
		},
	}

	PetInterface = &gql.Interface{
		Name: "Pet",
		Fields: gql.Fields{
			"name": &gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
			},
		},
		TypeResolver: func(ctx context.Context, r interface{}) *gql.Object {
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
			"name": &gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
			},
			"homePlanet": &gql.Field{
				Name: "homePlanet",
				Type: gql.String,
			},
		},
	}

	HumanType = &gql.Object{
		Name:       "Human",
		Implements: gql.Interfaces{SentientInterface},
		Fields: gql.Fields{
			"name": &gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
			},
		},
	}

	CatCommandEnum = &gql.Enum{
		Name: "CatCommand",
		Values: gql.EnumValues{
			&gql.EnumValue{
				Name:        "JUMP",
				Value:       "JUMP",
				Description: "JUMP",
			},
		},
	}

	CatType = &gql.Object{
		Name: "Cat",
		Fields: gql.Fields{
			"name": &gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
			},
			"nickname": &gql.Field{
				Name: "nickname",
				Type: gql.String,
			},
			"meowVolume": &gql.Field{
				Name: "meowVolume",
				Type: gql.Int,
			},
			"doesKnowCommand": &gql.Field{
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
		TypeResolver: func(ctx context.Context, r interface{}) *gql.Object {
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
