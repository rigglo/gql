package main

import (
	"context"
	"strconv"

	"github.com/rigglo/gql"
)

var (
	Luke           StarWarsChar
	Vader          StarWarsChar
	Han            StarWarsChar
	Leia           StarWarsChar
	Tarkin         StarWarsChar
	Threepio       StarWarsChar
	Artoo          StarWarsChar
	HumanData      map[int]StarWarsChar
	DroidData      map[int]StarWarsChar
	StarWarsSchema *gql.Schema

	humanType *gql.Object
	droidType *gql.Object
)

type StarWarsChar struct {
	ID              string
	Name            string
	Friends         []StarWarsChar
	AppearsIn       []int
	HomePlanet      string
	PrimaryFunction string
}

func init() {
	Luke = StarWarsChar{
		ID:         "1000",
		Name:       "Luke Skywalker",
		AppearsIn:  []int{4, 5, 6},
		HomePlanet: "Tatooine",
	}
	Vader = StarWarsChar{
		ID:         "1001",
		Name:       "Darth Vader",
		AppearsIn:  []int{4, 5, 6},
		HomePlanet: "Tatooine",
	}
	Han = StarWarsChar{
		ID:        "1002",
		Name:      "Han Solo",
		AppearsIn: []int{4, 5, 6},
	}
	Leia = StarWarsChar{
		ID:         "1003",
		Name:       "Leia Organa",
		AppearsIn:  []int{4, 5, 6},
		HomePlanet: "Alderaa",
	}
	Tarkin = StarWarsChar{
		ID:        "1004",
		Name:      "Wilhuff Tarkin",
		AppearsIn: []int{4},
	}
	Threepio = StarWarsChar{
		ID:              "2000",
		Name:            "C-3PO",
		AppearsIn:       []int{4, 5, 6},
		PrimaryFunction: "Protocol",
	}
	Artoo = StarWarsChar{
		ID:              "2001",
		Name:            "R2-D2",
		AppearsIn:       []int{4, 5, 6},
		PrimaryFunction: "Astromech",
	}
	Luke.Friends = append(Luke.Friends, []StarWarsChar{Han, Leia, Threepio, Artoo}...)
	Vader.Friends = append(Luke.Friends, []StarWarsChar{Tarkin}...)
	Han.Friends = append(Han.Friends, []StarWarsChar{Luke, Leia, Artoo}...)
	Leia.Friends = append(Leia.Friends, []StarWarsChar{Luke, Han, Threepio, Artoo}...)
	Tarkin.Friends = append(Tarkin.Friends, []StarWarsChar{Vader}...)
	Threepio.Friends = append(Threepio.Friends, []StarWarsChar{Luke, Han, Leia, Artoo}...)
	Artoo.Friends = append(Artoo.Friends, []StarWarsChar{Luke, Han, Leia}...)
	HumanData = map[int]StarWarsChar{
		1000: Luke,
		1001: Vader,
		1002: Han,
		1003: Leia,
		1004: Tarkin,
	}
	DroidData = map[int]StarWarsChar{
		2000: Threepio,
		2001: Artoo,
	}

	episodeEnum := &gql.Enum{
		Name: "Episode",
		Values: gql.EnumValues{
			&gql.EnumValue{
				Name:        "NEWHOPE",
				Value:       4,
				Description: "Released in 1977",
			},
			&gql.EnumValue{
				Name:        "EMPIRE",
				Value:       5,
				Description: "Released in 1980",
			},
			&gql.EnumValue{
				Name:        "JEDI",
				Value:       6,
				Description: "Released in 1983",
			},
		},
	}

	characterInterface := &gql.Interface{
		Name:        "Character",
		Description: "A character in the Star Wars Trilogy",
		Fields: gql.Fields{
			"id": &gql.Field{
				Type: gql.NewNonNull(gql.String),
			},
			"name": &gql.Field{
				Type: gql.String,
			},
			"appearsIn": &gql.Field{
				Type: gql.NewList(episodeEnum),
			},
		},
		TypeResolver: func(ctx context.Context, value interface{}) *gql.Object {
			if c, ok := value.(StarWarsChar); ok {
				id, _ := strconv.Atoi(c.ID)
				human := GetHuman(id)
				if human.ID != "" {
					return humanType
				}
			}
			return droidType
		},
	}
	characterInterface.Fields.Add("friends", &gql.Field{
		Type: gql.NewList(characterInterface),
	})

	humanType = &gql.Object{
		Name: "Human",
		Implements: gql.Interfaces{
			characterInterface,
		},
		Fields: gql.Fields{
			"id": &gql.Field{
				Type: gql.NewNonNull(gql.String),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if human, ok := ctx.Parent().(StarWarsChar); ok {
						return human.ID, nil
					}
					return nil, nil
				},
			},
			"name": &gql.Field{
				Type: gql.String,
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if human, ok := ctx.Parent().(StarWarsChar); ok {
						return human.Name, nil
					}
					return nil, nil
				},
			},
			"friends": &gql.Field{
				Type: gql.NewList(characterInterface),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if human, ok := ctx.Parent().(StarWarsChar); ok {
						return human.Friends, nil
					}
					return nil, nil
				},
			},
			"appearsIn": &gql.Field{
				Type: gql.NewList(episodeEnum),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if human, ok := ctx.Parent().(StarWarsChar); ok {
						return human.AppearsIn, nil
					}
					return nil, nil
				},
			},
			"homePlanet": &gql.Field{
				Type: gql.String,
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if human, ok := ctx.Parent().(StarWarsChar); ok {
						return human.HomePlanet, nil
					}
					return nil, nil
				},
			},
		},
	}

	droidType = &gql.Object{
		Name: "Droid",
		Fields: gql.Fields{
			"id": &gql.Field{
				Type: gql.NewNonNull(gql.String),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if droid, ok := ctx.Parent().(StarWarsChar); ok {
						return droid.ID, nil
					}
					return nil, nil
				},
			},
			"name": &gql.Field{
				Type: gql.NewNonNull(gql.String),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if droid, ok := ctx.Parent().(StarWarsChar); ok {
						return droid.Name, nil
					}
					return nil, nil
				},
			},
			"friends": &gql.Field{
				Type: gql.NewNonNull(gql.String),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if droid, ok := ctx.Parent().(StarWarsChar); ok {
						friends := []map[string]interface{}{}
						for _, friend := range droid.Friends {
							friends = append(friends, map[string]interface{}{
								"name": friend.Name,
								"id":   friend.ID,
							})
						}
						return droid.Friends, nil
					}
					return []interface{}{}, nil
				},
			},
			"appearsIn": &gql.Field{
				Type: gql.NewList(episodeEnum),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if droid, ok := ctx.Parent().(StarWarsChar); ok {
						return droid.AppearsIn, nil
					}
					return nil, nil
				},
			},
			"primaryFunction": &gql.Field{
				Type: gql.NewNonNull(gql.String),
				Resolver: func(ctx gql.Context) (interface{}, error) {
					if droid, ok := ctx.Parent().(StarWarsChar); ok {
						return droid.PrimaryFunction, nil
					}
					return nil, nil
				},
			},
		},
	}

	queryType := &gql.Object{
		Name: "Query",
		Fields: gql.Fields{
			"hero": &gql.Field{
				Type: characterInterface,
				Arguments: gql.Arguments{
					"episode": &gql.Argument{
						Type: episodeEnum,
					},
				},
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return GetHero(ctx.Args()["episode"]), nil
				},
			},
			"human": &gql.Field{
				Type: humanType,
				Arguments: gql.Arguments{
					"id": &gql.Argument{
						Type: gql.NewNonNull(gql.String),
					},
				},
				Resolver: func(ctx gql.Context) (interface{}, error) {
					id, err := strconv.Atoi(ctx.Args()["id"].(string))
					if err != nil {
						return nil, err
					}
					return GetHuman(id), nil
				},
			},
		},
	}

	StarWarsSchema = &gql.Schema{
		Query: queryType,
	}
}

func GetHuman(id int) StarWarsChar {
	if human, ok := HumanData[id]; ok {
		return human
	}
	return StarWarsChar{}
}
func GetDroid(id int) StarWarsChar {
	if droid, ok := DroidData[id]; ok {
		return droid
	}
	return StarWarsChar{}
}
func GetHero(episode interface{}) interface{} {
	if episode == 5 {
		return Luke
	}
	return Artoo
}
