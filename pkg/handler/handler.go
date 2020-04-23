package handler

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"

	"github.com/rigglo/gql"
)

type Config struct {
	Executor *gql.Executor
	GraphiQL bool
	Pretty   bool
}

func New(c Config) http.Handler {
	return &handler{
		conf: c,
	}
}

type handler struct {
	conf Config
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := new(gql.Params)
	switch r.Method {
	case http.MethodGet:
		{
			params = &gql.Params{
				Query:         html.UnescapeString(r.URL.Query().Get("query")),
				Variables:     map[string]interface{}{}, // TODO: find a way of doing this..
				OperationName: r.URL.Query().Get("operationName"),
			}
			if r.URL.Query().Get("variables") != "" {
				varsRaw := html.UnescapeString(r.URL.Query().Get("variables"))
				err := json.Unmarshal([]byte(varsRaw), &params.Variables)
				if err != nil {
					http.Error(w, `{"error": "invalid variables format"}`, http.StatusBadRequest)
					return
				}
			}
			if h.conf.GraphiQL {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprint(w, graphiql)
				return
			}
		}
	case http.MethodPost:
		{
			if r.Header.Get("Content-Type") == "application/json" {
				bs, err := ioutil.ReadAll(r.Body)
				if err != nil {
					break
				}

				err = json.Unmarshal(bs, params)
				if err != nil {
					http.Error(w, `{"error": "invalid parameters format"}`, http.StatusBadRequest)
					return
				}
			}
		}
	}
	if params != nil {
		var (
			bs  []byte
			err error
		)
		if h.conf.Pretty {
			bs, err = json.MarshalIndent(h.conf.Executor.Execute(r.Context(), *params), "", "\t")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			bs, err = json.Marshal(h.conf.Executor.Execute(r.Context(), *params))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(bs))
		return
	}
	http.Error(w, "invalid query parameters", http.StatusBadRequest)
}
