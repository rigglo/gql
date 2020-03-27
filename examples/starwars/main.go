package main

import (
	"fmt"
	"net/http"

	"github.com/rigglo/gql/pkg/handler"
)

func main() {
	http.Handle("/graphql", handler.New(handler.Config{
		Schema: StarWarsSchema,
	}))
	fmt.Println("Now server is running on port 8080")
	fmt.Println("Test with Get      : curl -g 'http://localhost:8080/graphql?query={hero{name}}'")
	http.ListenAndServe(":9999", nil)
}
