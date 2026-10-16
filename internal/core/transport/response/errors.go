package core_transport_response

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
