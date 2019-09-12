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
	ErrNoResolver = "No resolver found"
)
