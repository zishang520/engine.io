package errors

type Error struct {
	Message     string
	Type        string
	Description string
}

func New(message string) *Error {
	return &Error{Message: message}
}

func (e *Error) Err() error {
	return e
}

func (e *Error) Error() string {
	return e.Message
}
