package core_transport_request

import "fmt"

type ClassError string

const (
	EmptyError        ClassError = "empty"
	MalformedError    ClassError = "malformed"
	UnknownFieldError ClassError = "unknown_field"
	TrailingError     ClassError = "trailing"
)

type Error struct {
	Class   ClassError
	Field   string
	Message string
	Cause   error
}

func (e Error) Error() string {
	return fmt.Sprintf("Class: %s Field: %s Message: %s, Cause: %v", e.Class, e.Field, e.Message, e.Cause)
}

func (e Error) Unwrap() error {
	return e.Cause
}
