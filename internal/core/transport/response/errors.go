package core_transport_response

const ()

type Error struct {
	Detail ErrorBody `json:"error"`
}

type ErrorBody struct {
	Message   string       `json:"message"`
	Fields    []FieldError `json:"fields,omitempty"`
	RequestID string       `json:"request_id,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
