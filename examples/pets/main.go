package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/rigglo/gql"
	"github.com/rigglo/gql/pkg/handler"
)

func init() {
	DogType.Implements = gql.Interfaces{PetInterface}
	CatType.Implements = gql.Interfaces{PetInterface}
}

func main() {
	h := handler.New(handler.Config{
		Executor:   gql.DefaultExecutor(PetStore),
		Playground: true,
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
				Type: DogType,
				Resolver: func(ctx gql.Context) (interface{}, error) {
					log.Println("start dog", time.Now())
					time.Sleep(2 * time.Second)
					log.Println("finish dog", time.Now())
					return Doggo, nil
				},
			},
			"cat": &gql.Field{
				Type: CatType,
				Resolver: func(ctx gql.Context) (interface{}, error) {
					log.Println("start cat", time.Now())
					time.Sleep(4 * time.Second)
					log.Println("finish cat", time.Now())
					return Catcy, nil
				},
			},
			"pet": &gql.Field{
				Type: gql.NewList(PetInterface),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return []interface{}{Catcy, Doggo}, nil
				},
			},
			"catOrDog": &gql.Field{
				Type: gql.NewList(CatOrDogUnion),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return []interface{}{Catcy, Doggo}, nil
				},
			},
			"someErr": &gql.Field{
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
				Type: gql.NewNonNull(gql.String),
			},
			"nickname": &gql.Field{
				Type: gql.String,
			},
			"barkVolume": &gql.Field{
				Type: gql.Int,
			},
			"doesKnowCommand": &gql.Field{
				Arguments: gql.Arguments{
					"dogCommand": &gql.Argument{
						Type: gql.NewNonNull(DogCommandEnum),
					},
				},
				Type: gql.NewNonNull(gql.Boolean),
			},
			"isHousetrained": &gql.Field{
				Arguments: gql.Arguments{
					"atOtherHomes": &gql.Argument{
						Type: gql.Boolean,
					},
				},
				Type: gql.NewNonNull(gql.Boolean),
			},
			"owner": &gql.Field{
				Type: HumanType,
			},
		},
	}

	SentientInterface = &gql.Interface{
		Name: "Sentient",
		Fields: gql.Fields{
			"name": &gql.Field{
				Type: gql.NewNonNull(gql.String),
			},
		},
	}

	PetInterface = &gql.Interface{
		Name: "Pet",
		Fields: gql.Fields{
			"name": &gql.Field{
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
				Type: gql.NewNonNull(gql.String),
			},
			"homePlanet": &gql.Field{
				Type: gql.String,
			},
		},
	}

	HumanType = &gql.Object{
		Name:       "Human",
		Implements: gql.Interfaces{SentientInterface},
		Fields: gql.Fields{
			"name": &gql.Field{
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
				Type: gql.NewNonNull(gql.String),
			},
			"nickname": &gql.Field{
				Type: gql.String,
			},
			"meowVolume": &gql.Field{
				Type: gql.Int,
			},
			"doesKnowCommand": &gql.Field{
				Arguments: gql.Arguments{
					"catCommand": &gql.Argument{
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
