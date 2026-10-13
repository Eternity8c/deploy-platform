package features_handler

import (
	"net/http"

	core_logger "github.com/Eternity8c/deploy-platform/internal/core/logger"
	core_transport_responce "github.com/Eternity8c/deploy-platform/internal/core/transport/responce"
)

func Healthz(w http.ResponseWriter, r *http.Request) {
	logger := core_logger.FromContext(r.Context())
	responce := core_transport_responce.NewHTTPResponceHandler(logger, w)
	responce.JSONResponce(
		map[string]string{"status": "ok"},
		http.StatusOK,
	)
}
