package schema

import (
	"context"
)

type Errors []*Error

type Error struct {
	Message    string                 `json:"message"`
	Locations  []*ErrorLocation       `json:"locations"`
	Path       []interface{}          `json:"path"`
	Extensions map[string]interface{} `json:"extensions"`
}

func (e *Error) Error() string {
	return e.Message
}

type ErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func NewError(ctx context.Context, msg string, extensions map[string]interface{}) *Error {
	return &Error{
		Message:    msg,
		Extensions: extensions,
	}
}
