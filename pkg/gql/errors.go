package gql

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

func NewError(msg string, extensions map[string]interface{}) CustomError {
	return &cErr{
		msg:  msg,
		exts: extensions,
	}
}

type cErr struct {
	msg  string
	exts map[string]interface{}
}

func (c *cErr) Error() string {
	return c.msg
}

func (c *cErr) GetMessage() string {
	return c.msg
}

func (c *cErr) GetExtensions() map[string]interface{} {
	return c.exts
}

type CustomError interface {
	error
	GetMessage() string
	GetExtensions() map[string]interface{}
}

const (
	ErrValidateOperationName                  = "Operation name '%s' is defined multiple times"
	ErrAnonymousOperationDefinitions          = "Can not use anonymous operation where multiple operation definitions exist"
	ErrLeafFieldSelectionsSelectionNotAllowed = "Selection on type '%s' is not allowed"
	ErrLeafFieldSelectionsSelectionMissing    = "Selection on type '%s' is missing"
	ErrFieldDoesNotExist                      = "Field '%s' does not exist on type '%s'"
	ErrResponseShapeMismatch                  = "fields in set can not be merged: %s"
)
