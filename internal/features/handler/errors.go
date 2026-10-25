package features_handler

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	core_logger "github.com/Eternity8c/deploy-platform/internal/core/logger"
	core_transport_middleware "github.com/Eternity8c/deploy-platform/internal/core/transport/middleware"
	core_transport_request "github.com/Eternity8c/deploy-platform/internal/core/transport/request"
	core_transport_response "github.com/Eternity8c/deploy-platform/internal/core/transport/response"
	features_application "github.com/Eternity8c/deploy-platform/internal/features/application"
)

func writeError(logger *core_logger.Logger, w http.ResponseWriter, r *http.Request, err error) {
	responseHandler := core_transport_response.NewHTTPResponseHandler(logger, w)

	// id запроса выставляет мидлварь RequestID. Если её нет в цепочке,
	// поле останется пустым — это видно в ответе, а не тихо ломает форму.
	requestID := core_transport_middleware.RequestIDFromRequest(r)

	var (
		validationErr features_application.ValidationErrors
		requestErr    core_transport_request.Error
	)

	switch {
	case errors.As(err, &validationErr):
		responseHandler.WriteError(http.StatusBadRequest, errorBody(
			"request validation failed",
			fieldsFrom(validationErr),
			requestID,
		))

	case errors.Is(err, features_application.ErrAlreadyExists):
		responseHandler.WriteError(http.StatusConflict, errorBody(
			"application already exists",
			nil,
			requestID,
		))

	case errors.As(err, &requestErr):
		responseHandler.WriteError(http.StatusBadRequest, requestErrorBody(requestErr, requestID))

	default:
		logger.Error("handle request", zap.Error(err))

		responseHandler.WriteError(http.StatusInternalServerError, errorBody(
			"internal server error",
			nil,
			requestID,
		))
	}
}

func requestErrorBody(e core_transport_request.Error, requestID string) core_transport_response.Error {
	switch e.Class {
	case core_transport_request.EmptyError:
		return errorBody("request body is empty", nil, requestID)

	case core_transport_request.MalformedError:
		return errorBody("request body is not valid JSON", nil, requestID)

	case core_transport_request.TrailingError:
		return errorBody("request body must contain exactly one JSON value", nil, requestID)

	case core_transport_request.UnknownFieldError:
		return errorBody("request body contains a field that is not part of the schema",
			[]core_transport_response.FieldError{{Field: e.Field, Message: e.Message}},
			requestID,
		)

	default:
		return errorBody("request body is not valid", nil, requestID)
	}
}

func fieldsFrom(ves features_application.ValidationErrors) []core_transport_response.FieldError {
	fields := make([]core_transport_response.FieldError, 0, len(ves))
	for _, one := range ves {
		fields = append(fields, core_transport_response.FieldError{
			Field:   one.Field,
			Message: one.Message,
		})
	}

	return fields
}

func errorBody(
	message string,
	fields []core_transport_response.FieldError,
	requestID string,
) core_transport_response.Error {
	return core_transport_response.Error{
		Detail: core_transport_response.ErrorBody{
			Message:   message,
			Fields:    fields,
			RequestID: requestID,
		},
	}
}
