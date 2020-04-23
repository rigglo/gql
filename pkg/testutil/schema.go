package testutil

import (
	"context"

	"github.com/rigglo/gql"
)

func init() {
	DogType.Implements = gql.Interfaces{PetInterface}
	CatType.Implements = gql.Interfaces{PetInterface}
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

type Alien struct {
	Name       string `json:"name"`
	HomePlanet string `json:"homePlanet"`
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

	Bob = Human{
		Name: "Bob",
	}

	Elon = Alien{
		Name:       "Musk",
		HomePlanet: "Teslaversum",
	}

	Schema = &gql.Schema{
		Query: Query,
	}

	Query = &gql.Object{
		Name: "Query",
		Fields: gql.Fields{
			"dog": &gql.Field{
				Type: DogType,
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return Doggo, nil
				},
			},
			"cat": &gql.Field{
				Type: CatType,
				Resolver: func(ctx gql.Context) (interface{}, error) {
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
			"dogOrHuman": &gql.Field{
				Type: gql.NewList(DogOrHumanUnion),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return []interface{}{Doggo, Bob}, nil
				},
			},
			"humanOrAlian": &gql.Field{
				Type: gql.NewList(HumanOrAlienUnion),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return []interface{}{Bob, Elon}, nil
				},
			},
			"sentient": &gql.Field{
				Type: gql.NewList(SentientInterface),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return []interface{}{Bob, Elon}, nil
				},
			},
			"someErr": &gql.Field{
				Type: gql.NewList(CatOrDogUnion),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return nil, gql.NewError("asdasdasd", map[string]interface{}{"code": "SOME_ERR_CODE"})
				},
			},
			"arguments": &gql.Field{
				Type: Arguments,
				Resolver: func(c gql.Context) (interface{}, error) {
					return true, nil
				},
			},
			"findDog": &gql.Field{
				Type: DogType,
				Arguments: gql.Arguments{
					"complex": &gql.Argument{
						Type: ComplexInput,
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return Doggo, nil
				},
			},
		},
	}

	Arguments = &gql.Object{
		Name: "Arguments",
		Fields: gql.Fields{
			"multipleReqs": &gql.Field{
				Type: gql.NewNonNull(gql.Int),
				Arguments: gql.Arguments{
					"x": &gql.Argument{
						Type: gql.NewNonNull(gql.Int),
					},
					"y": &gql.Argument{
						Type: gql.NewNonNull(gql.Int),
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return 42, nil
				},
			},
			"booleanArgField": &gql.Field{
				Type: gql.Boolean,
				Arguments: gql.Arguments{
					"booleanArg": &gql.Argument{
						Type: gql.Boolean,
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return true, nil
				},
			},
			"floatArgField": &gql.Field{
				Type: gql.Float,
				Arguments: gql.Arguments{
					"floatArg": &gql.Argument{
						Type: gql.Float,
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return 0.42, nil
				},
			},
			"intArgField": &gql.Field{
				Type: gql.Int,
				Arguments: gql.Arguments{
					"intArg": &gql.Argument{
						Type: gql.Int,
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return 42, nil
				},
			},
			"nonNullBooleanArgField": &gql.Field{
				Type: gql.NewNonNull(gql.Boolean),
				Arguments: gql.Arguments{
					"nonNullBooleanArg": &gql.Argument{
						Type: gql.NewNonNull(gql.Boolean),
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return true, nil
				},
			},
			"booleanListArgField": &gql.Field{
				Type: gql.NewList(gql.Boolean),
				Arguments: gql.Arguments{
					"booleanListArg": &gql.Argument{
						Type: gql.NewNonNull(gql.NewList(gql.Boolean)),
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return []bool{true, false}, nil
				},
			},
			"optionalNonNullBooleanArgField": &gql.Field{
				Type: gql.Int,
				Arguments: gql.Arguments{
					"optionalBooleanArg": &gql.Argument{
						Type:         gql.NewNonNull(gql.Boolean),
						DefaultValue: false,
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return true, nil
				},
			},
		},
	}

	ComplexInput = &gql.InputObject{
		Name: "ComplexInput",
		Fields: gql.InputFields{
			"name": &gql.InputField{
				Type: gql.String,
			},
			"owner": &gql.InputField{
				Type: gql.String,
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
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return true, nil
				},
			},
			"isHousetrained": &gql.Field{
				Arguments: gql.Arguments{
					"atOtherHomes": &gql.Argument{
						Type: gql.Boolean,
					},
				},
				Type: gql.NewNonNull(gql.Boolean),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return true, nil
				},
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
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return true, nil
				},
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
