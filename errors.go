package gql

type Error struct {
	Path    []interface{}
	Message string
	Err     error
}

func (e *Error) Error() string {
	return e.Message
}

const (
	ErrNoResolver   = "no resolver found"
	ErrNullValue    = "got null where should not"
	ErrArgNotExists = "the given argument does not exist"
)
