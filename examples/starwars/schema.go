package main

import (
	"context"
	"strconv"

	"github.com/rigglo/gql/pkg/gql"
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
			"NEWHOPE": &gql.EnumValue{
				Value:       4,
				Description: "Released in 1977",
			},
			"EMPIRE": &gql.EnumValue{
				Value:       5,
				Description: "Released in 1980",
			},
			"JEDI": &gql.EnumValue{
				Value:       6,
				Description: "Released in 1983",
			},
		},
	}

	characterInterface := &gql.Interface{
		Name:        "Character",
		Description: "A character in the Star Wars Trilogy",
		Fields: gql.Fields{
			&gql.Field{
				Name: "id",
				Type: gql.NewNonNull(gql.String),
			},
			&gql.Field{
				Name: "name",
				Type: gql.String,
			},
			&gql.Field{
				Name: "appearsIn",
				Type: gql.NewList(episodeEnum),
			},
		},
		TypeResolver: func(ctx context.Context, value interface{}) gql.ObjectType {
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
	characterInterface.Fields.Add(&gql.Field{
		Name: "friends",
		Type: gql.NewList(characterInterface),
	})

	humanType = &gql.Object{
		Name: "Human",
		Implements: gql.Interfaces{
			characterInterface,
		},
		Fields: gql.Fields{
			&gql.Field{
				Name: "id",
				Type: gql.NewNonNull(gql.String),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if human, ok := parent.(StarWarsChar); ok {
						return human.ID, nil
					}
					return nil, nil
				},
			},
			&gql.Field{
				Name: "name",
				Type: gql.String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if human, ok := parent.(StarWarsChar); ok {
						return human.Name, nil
					}
					return nil, nil
				},
			},
			&gql.Field{
				Name: "friends",
				Type: gql.NewList(characterInterface),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if human, ok := parent.(StarWarsChar); ok {
						return human.Friends, nil
					}
					return nil, nil
				},
			},
			&gql.Field{
				Name: "appearsIn",
				Type: gql.NewList(episodeEnum),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if human, ok := parent.(StarWarsChar); ok {
						return human.AppearsIn, nil
					}
					return nil, nil
				},
			},
			&gql.Field{
				Name: "homePlanet",
				Type: gql.String,
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if human, ok := parent.(StarWarsChar); ok {
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
			&gql.Field{
				Name: "id",
				Type: gql.NewNonNull(gql.String),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if droid, ok := parent.(StarWarsChar); ok {
						return droid.ID, nil
					}
					return nil, nil
				},
			},
			&gql.Field{
				Name: "name",
				Type: gql.NewNonNull(gql.String),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if droid, ok := parent.(StarWarsChar); ok {
						return droid.Name, nil
					}
					return nil, nil
				},
			},
			&gql.Field{
				Name: "friends",
				Type: gql.NewNonNull(gql.String),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if droid, ok := parent.(StarWarsChar); ok {
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
			&gql.Field{
				Name: "appearsIn",
				Type: gql.NewList(gql.String),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if droid, ok := parent.(StarWarsChar); ok {
						return droid.AppearsIn, nil
					}
					return nil, nil
				},
			},
			&gql.Field{
				Name: "primaryFunction",
				Type: gql.NewNonNull(gql.String),
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					if droid, ok := parent.(StarWarsChar); ok {
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
			&gql.Field{
				Name: "hero",
				Type: characterInterface,
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "episode",
						Type: episodeEnum,
					},
				},
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					return GetHero(args["episode"]), nil
				},
			},
			&gql.Field{
				Name: "human",
				Type: humanType,
				Arguments: gql.Arguments{
					&gql.Argument{
						Name: "id",
						Type: gql.NewNonNull(gql.String),
					},
				},
				Resolver: func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
					id, err := strconv.Atoi(args["id"].(string))
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
