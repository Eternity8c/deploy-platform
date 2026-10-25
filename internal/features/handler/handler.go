package features_handler

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	core_logger "github.com/Eternity8c/deploy-platform/internal/core/logger"
	core_repo_postgres "github.com/Eternity8c/deploy-platform/internal/core/repository/postgres"
	core_transport_request "github.com/Eternity8c/deploy-platform/internal/core/transport/request"
	core_transport_response "github.com/Eternity8c/deploy-platform/internal/core/transport/response"
	features_application "github.com/Eternity8c/deploy-platform/internal/features/application"
)

type ApplicationHandler struct {
	store Store
}

type Store interface {
	Create(ctx context.Context, app features_application.Application) (features_application.CreateResult, error)
}

func NewApplicationHandler(s Store) *ApplicationHandler {
	return &ApplicationHandler{
		store: s,
	}
}

func (h *ApplicationHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	logger := core_logger.FromContext(r.Context())
	responseHandler := core_transport_response.NewHTTPResponseHandler(logger, w)
	responseHandler.JSONResponse(
		map[string]string{"status": "ok"},
		http.StatusOK,
	)
}

func (h *ApplicationHandler) Readyz(p *core_repo_postgres.ConnectionPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := core_logger.FromContext(r.Context())
		responseHandler := core_transport_response.NewHTTPResponseHandler(logger, w)
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := p.Ping(ctx); err != nil {
			logger.Error("readyz handler:", zap.Error(err))
			responseHandler.JSONResponse(map[string]string{"status": "unavailable"}, http.StatusServiceUnavailable)
			return
		}

		responseHandler.JSONResponse(
			map[string]string{"status": "ok"},
			http.StatusOK,
		)
	}
}

func (h *ApplicationHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	logger := core_logger.FromContext(r.Context())
	responseHandler := core_transport_response.NewHTTPResponseHandler(logger, w)
	ctx := r.Context()

	var app features_application.Application
	if err := core_transport_request.Decode(r, &app); err != nil {
		writeError(logger, w, r, err)
		return
	}

	if err := app.Validate(); err != nil {
		writeError(logger, w, r, err)
		return
	}

	response, err := h.store.Create(ctx, app)
	if err != nil {
		writeError(logger, w, r, err)
		return
	}

	responseHandler.JSONResponse(response, http.StatusCreated)
}
