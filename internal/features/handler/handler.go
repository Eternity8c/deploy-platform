package features_handler

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	core_logger "github.com/Eternity8c/deploy-platform/internal/core/logger"
	core_repo_postgres "github.com/Eternity8c/deploy-platform/internal/core/repository/postgres"
	core_transport_response "github.com/Eternity8c/deploy-platform/internal/core/transport/response"
)

func Healthz(w http.ResponseWriter, r *http.Request) {
	logger := core_logger.FromContext(r.Context())
	response := core_transport_response.NewHTTPResponseHandler(logger, w)
	response.JSONResponse(
		map[string]string{"status": "ok"},
		http.StatusOK,
	)
}

func Readyz(p *core_repo_postgres.ConnectionPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := core_logger.FromContext(r.Context())
		response := core_transport_response.NewHTTPResponseHandler(logger, w)
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := p.Ping(ctx); err != nil {
			logger.Error("readyz handler:", zap.Error(err))
			response.JSONResponse(map[string]string{"status": "unavailable"}, http.StatusServiceUnavailable)
			return
		}

		response.JSONResponse(
			map[string]string{"status": "ok"},
			http.StatusOK,
		)
	}
}
