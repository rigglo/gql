package gql

// Result ...
type Result struct {
	Data       map[string]interface{} `json:"data"`
	Extensions map[string]interface{} `json:"extensions"`
	Errors     []error                `json:"errors"`
}
