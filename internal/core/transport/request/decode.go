package core_transport_request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func Decode(r *http.Request, dest any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(dest)
	if err != nil {
		return classify(err)
	}

	var extra any
	err = decoder.Decode(&extra)
	if !errors.Is(err, io.EOF) {
		return Error{
			Class:   TrailingError,
			Message: "body must contain exactly one JSON value",
			Cause:   err,
		}
	}

	return nil
}

func classify(err error) error {
	if errors.Is(err, io.EOF) {
		return Error{
			Class:   EmptyError,
			Message: "request body is empty",
		}
	}

	unknownFieldPrefix := "json: unknown field "

	msg := err.Error()
	if strings.HasPrefix(msg, unknownFieldPrefix) {
		quoted := strings.TrimPrefix(msg, unknownFieldPrefix)
		if name, uerr := strconv.Unquote(quoted); uerr == nil {
			return Error{
				Class:   UnknownFieldError,
				Field:   name,
				Message: "field is not part of the schema",
			}
		}
	}

	return Error{
		Class:   MalformedError,
		Message: "request body is not valid JSON",
		Cause:   err,
	}
}
