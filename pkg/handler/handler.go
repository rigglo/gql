package handler

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"

	"github.com/rigglo/gql/pkg/gql"
)

type Config struct {
	Schema *gql.Schema
}

func New(c Config) http.Handler {
	return &handler{
		conf: c,
	}
}

type postParams struct {
	Query         string          `json:"query"`
	Variables     json.RawMessage `json:"variables"`
	OperationName string          `json:"operationName"`
}

type handler struct {
	conf Config
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var params *gql.ExecuteParams
	switch r.Method {
	case http.MethodGet:
		{
			params = &gql.ExecuteParams{
				Query:         html.UnescapeString(r.URL.Query().Get("query")),
				Variables:     html.UnescapeString(r.URL.Query().Get("variables")),
				OperationName: r.URL.Query().Get("operationName"),
			}
		}
	case http.MethodPost:
		{
			bs, err := ioutil.ReadAll(r.Body)
			if err != nil {
				break
			}

			p := postParams{}

			err = json.Unmarshal(bs, &p)
			if err != nil {
				break
			}

			params = &gql.ExecuteParams{
				Query:         html.UnescapeString(p.Query),
				Variables:     string(p.Variables),
				OperationName: p.OperationName,
			}
		}
	}
	if params != nil {
		bs, err := json.MarshalIndent(gql.Execute(r.Context(), h.conf.Schema, *params), "", "\t")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(bs))
		return
	}
	http.Error(w, "invalid query parameters", http.StatusBadRequest)
}
