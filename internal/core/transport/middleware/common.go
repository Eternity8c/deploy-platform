package core_transport_middleware

import (
	"context"
	"net/http"

	core_logger "github.com/Eternity8c/deploy-platform/internal/core/logger"
	core_transport_responce "github.com/Eternity8c/deploy-platform/internal/core/transport/responce"
	"go.uber.org/zap"
)

const requestIDHeader = "X-Request-ID"

func Logger(log *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)

			l := log.With(
				zap.String("request_id", requestID),
				zap.String("url", r.URL.String()),
			)

			ctx := context.WithValue(r.Context(), "log", l)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Panic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			responceHandler := core_transport_responce.NewHTTPResponceHandler(log, w)

			defer func() {
				if p := recover(); p != nil {
					responceHandler.PanicResponce(
						p,
						"during handle HTTP request got unexpcted panic",
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
