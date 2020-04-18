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
				Name: "dog",
				Type: DogType,
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return Doggo, nil
				},
			},
			"cat": &gql.Field{
				Name: "cat",
				Type: CatType,
				Resolver: func(ctx gql.Context) (interface{}, error) {
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
			"dogOrHuman": &gql.Field{
				Name: "dogOrHuman",
				Type: gql.NewList(DogOrHumanUnion),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return []interface{}{Doggo, Bob}, nil
				},
			},
			"humanOrAlian": &gql.Field{
				Name: "humanOrAlian",
				Type: gql.NewList(HumanOrAlienUnion),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return []interface{}{Bob, Elon}, nil
				},
			},
			"sentient": &gql.Field{
				Name: "sentient",
				Type: gql.NewList(SentientInterface),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return []interface{}{Bob, Elon}, nil
				},
			},
			"someErr": &gql.Field{
				Name: "someErr",
				Type: gql.NewList(CatOrDogUnion),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return nil, gql.NewError("asdasdasd", map[string]interface{}{"code": "SOME_ERR_CODE"})
				},
			},
			"arguments": &gql.Field{
				Name: "arguments",
				Type: Arguments,
				Resolver: func(c gql.Context) (interface{}, error) {
					return true, nil
				},
			},
			"findDog": &gql.Field{
				Name: "findDog",
				Type: DogType,
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "complex",
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
				Name: "multipleReqs",
				Type: gql.NewNonNull(gql.Int),
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "x",
						Type: gql.NewNonNull(gql.Int),
					},
					&gql.Argument{
						Name: "y",
						Type: gql.NewNonNull(gql.Int),
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return 42, nil
				},
			},
			"booleanArgField": &gql.Field{
				Name: "booleanArgField",
				Type: gql.Boolean,
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "booleanArg",
						Type: gql.Boolean,
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return true, nil
				},
			},
			"floatArgField": &gql.Field{
				Name: "floatArgField",
				Type: gql.Float,
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "floatArg",
						Type: gql.Float,
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return 0.42, nil
				},
			},
			"intArgField": &gql.Field{
				Name: "intArgField",
				Type: gql.Int,
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "intArg",
						Type: gql.Int,
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return 42, nil
				},
			},
			"nonNullBooleanArgField": &gql.Field{
				Name: "nonNullBooleanArgField",
				Type: gql.NewNonNull(gql.Boolean),
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "nonNullBooleanArg",
						Type: gql.NewNonNull(gql.Boolean),
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return true, nil
				},
			},
			"booleanListArgField": &gql.Field{
				Name: "booleanListArgField",
				Type: gql.NewList(gql.Boolean),
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "booleanListArg",
						Type: gql.NewNonNull(gql.NewList(gql.Boolean)),
					},
				},
				Resolver: func(c gql.Context) (interface{}, error) {
					return 42, nil
				},
			},
			"optionalNonNullBooleanArgField": &gql.Field{
				Name: "optionalNonNullBooleanArgField",
				Type: gql.Int,
				Arguments: gql.Arguments{
					&gql.Argument{
						Name:         "optionalBooleanArg",
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
			&gql.InputField{
				Name: "name",
				Type: gql.String,
			},
			&gql.InputField{
				Name: "owner",
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
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return true, nil
				},
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
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return true, nil
				},
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
