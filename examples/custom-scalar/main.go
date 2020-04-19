package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/rigglo/gql"
	"github.com/rigglo/gql/pkg/handler"
)

func main() {
	h := handler.New(handler.Config{
		Executor: gql.DefaultExecutor(TimeMachine),
	})
	http.Handle("/graphql", h)
	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Println(err)
	}
}

type Movie struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

var (
	TimeMachine = &gql.Schema{
		Query:    Query,
		Mutation: nil,
	}

	Query = &gql.Object{
		Name:        "Query",
		Description: "a box..",
		Fields: gql.Fields{
			"server_time": &gql.Field{
				Type:        UnixTimestampScalar,
				Description: "",
				Resolver: func(ctx gql.Context) (interface{}, error) {
					return time.Now(), nil
				},
			},
		},
	}

	UnixTimestampScalar *gql.Scalar = &gql.Scalar{
		Name:        "UnixTimestamp",
		Description: "This is a custom scalar that converts time.Time into a unix timestamp and a string formatted unix timestamp to time.Time",
		CoerceResultFunc: func(i interface{}) (interface{}, error) {
			if i == nil {
				return nil, nil
			}
			switch i.(type) {
			case time.Time:
				return fmt.Sprintf("%v", i.(time.Time).Unix()), nil
			default:
				return nil, errors.New("invalid value to coerce")
			}
		},
		CoerceInputFunc: func(i interface{}) (interface{}, error) {
			var (
				unix int64
				err  error
			)

			switch i.(type) {
			case []byte:
				var s string
				err = json.Unmarshal(i.([]byte), &s)
				if err != nil {
					return nil, err
				}
				unix, err = strconv.ParseInt(s, 10, 64)
				if err != nil {
					return nil, err
				}
			case string:
				unix, err = strconv.ParseInt(i.(string), 10, 64)
				if err != nil {
					return nil, err
				}
			}
			return time.Unix(unix, 0), nil

		},
	}
)
