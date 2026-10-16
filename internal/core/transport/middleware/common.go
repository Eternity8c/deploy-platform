package core_transport_middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	core_logger "github.com/Eternity8c/deploy-platform/internal/core/logger"
	core_transport_response "github.com/Eternity8c/deploy-platform/internal/core/transport/response"
)

type contextKey string

const (
	requestIDHeader            = "X-Request-ID"
	logKey          contextKey = "log"
)

func Logger(log *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)

			l := log.With(
				zap.String("request_id", requestID),
				zap.String("url", r.URL.String()),
			)

			ctx := context.WithValue(r.Context(), logKey, l)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Panic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			responseHandler := core_transport_response.NewHTTPResponseHandler(log, w)

			defer func() {
				if p := recover(); p != nil {
					responseHandler.PanicResponse(
						p,
						"during handle HTTP request got unexpcted panic",
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
