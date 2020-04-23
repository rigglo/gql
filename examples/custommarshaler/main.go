package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rigglo/gql"
	"github.com/rigglo/gql/pkg/handler"
)

func main() {
	h := handler.New(handler.Config{
		Executor: gql.DefaultExecutor(Schema),
		GraphiQL: true,
	})
	http.Handle("/graphql", h)
	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Println(err)
	}
}

var (
	Schema = &gql.Schema{
		Query: Query,
	}

	Query = &gql.Object{
		Name: "Query",
		Fields: gql.Fields{
			"customVal": &gql.Field{
				Type: gql.ID,
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return myVal{"SomeValue"}, nil
				},
			},
		},
	}
)

type myVal struct {
	val string
}

func (v myVal) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"--------%s--------\"", v.val)), nil
}
