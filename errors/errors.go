package errors

type Error struct {
	Message     string
	Description error
	Type        string
}

func (e *Error) Err() error {
	return e
}

func (e *Error) Error() string {
	return e.Message
}

func New(message string) *Error {
	return &Error{Message: message}
}

func NewTransportError(reason string, description error) *Error {
	return &Error{
		Message:     reason,
		Description: description,
		Type:        "TransportError",
	}
}
