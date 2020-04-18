package gql

/*
Errors is an alias for a bunch of Error
*/
type Errors []*Error

/*
Error for the Result of the validation/execution
*/
type Error struct {
	Message    string                 `json:"message"`
	Locations  []*ErrorLocation       `json:"locations"`
	Path       []interface{}          `json:"path"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

/*
Error implements the error interface
*/
func (e *Error) Error() string {
	return e.Message
}

/*
ErrorLocation represents the location of an error in the query
*/
type ErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

/*
NewError ...
*/
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
	errValidateOperationName                  = "Operation name '%s' is defined multiple times"
	errAnonymousOperationDefinitions          = "Can not use anonymous operation where multiple operation definitions exist"
	errLeafFieldSelectionsSelectionNotAllowed = "Selection on type '%s' is not allowed"
	errLeafFieldSelectionsSelectionMissing    = "Selection on type '%s' is missing"
	errFieldDoesNotExist                      = "Field '%s' does not exist on type '%s'"
	errResponseShapeMismatch                  = "fields in set can not be merged: %s"
)
