package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rigglo/gql/pkg/gql"
)

type Config struct {
	schema *gql.Schema
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
	var params *gql.ExecuteParams
	switch r.Method {
	case http.MethodGet:
		{
			params = &gql.ExecuteParams{
				Query:         r.URL.Query().Get("query"),
				Variables:     r.URL.Query().Get("variables"),
				OperationName: r.URL.Query().Get("operationName"),
			}
		}
	}
	if params != nil {
		bs, err := json.MarshalIndent(gql.Execute(r.Context(), h.conf.schema, *params), "", "\t")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, bs)
	}
	http.Error(w, "invalid query parameters", http.StatusBadRequest)
}
