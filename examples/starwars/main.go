package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rigglo/gql/pkg/gql"
)

func main() {
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "couldn't read body", http.StatusInternalServerError)
		}
		params := gql.Params{}
		err = json.Unmarshal(bs, &params)
		if err != nil {
			http.Error(w, "couldn't read body", http.StatusInternalServerError)
		}
		result := StarWarsSchema.Execute(&params)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
	fmt.Println("Now server is running on port 8080")
	fmt.Println("Test with Get      : curl -g 'http://localhost:8080/graphql?query={hero{name}}'")
	http.ListenAndServe(":8080", nil)
}
