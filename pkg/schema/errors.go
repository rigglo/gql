package schema

import (
	"context"
)

type Errors []*Error

type Error struct {
	Message    string                 `json:"message"`
	Locations  []*ErrorLocation       `json:"locations"`
	Path       []interface{}          `json:"path"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
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

const (
	ErrValidateOperationName                  = "Operation name '%s' is defined multiple times"
	ErrAnonymousOperationDefinitions          = "Can not use anonymous operation where multiple operation definitions exist"
	ErrLeafFieldSelectionsSelectionNotAllowed = "Selection on type '%s' is not allowed"
	ErrLeafFieldSelectionsSelectionMissing    = "Selection on type '%s' is missing"
	ErrFieldDoesNotExist                      = "Field '%s' does not exist on type '%s'"
)
