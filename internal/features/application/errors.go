package features_application

import (
	"errors"
	"strings"
)

var ErrAlreadyExists = errors.New("application already exists")

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}

	return e.Field + ": " + e.Message
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	parts := make([]string, 0, len(e))
	for _, one := range e {
		parts = append(parts, one.Error())
	}

	return strings.Join(parts, "; ")
}
