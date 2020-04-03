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
	Executor *gql.Executor
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
	var params *gql.Params
	switch r.Method {
	case http.MethodGet:
		{
			params = &gql.Params{
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

			params = &gql.Params{
				Query:         html.UnescapeString(p.Query),
				Variables:     string(p.Variables),
				OperationName: p.OperationName,
			}
		}
	}
	if params != nil {
		bs, err := json.MarshalIndent(h.conf.Executor.Execute(r.Context(), *params), "", "\t")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(bs))
		return
	}
	http.Error(w, "invalid query parameters", http.StatusBadRequest)
}
