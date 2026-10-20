package features_application

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func invalid(field, msg string) error {
	return ValidationError{
		Field:   field,
		Message: msg,
	}
}
