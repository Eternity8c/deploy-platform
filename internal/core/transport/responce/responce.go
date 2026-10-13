package core_transport_responce

import (
	"encoding/json"
	"fmt"
	"net/http"

	core_logger "github.com/Eternity8c/deploy-platform/internal/core/logger"
	"go.uber.org/zap"
)

type HTTPResponceHandler struct {
	log *core_logger.Logger
	rw  http.ResponseWriter
}

func NewHTTPResponceHandler(
	log *core_logger.Logger,
	rw http.ResponseWriter,
) *HTTPResponceHandler {
	return &HTTPResponceHandler{
		log: log,
		rw:  rw,
	}
}

func (h *HTTPResponceHandler) JSONResponce(responceBody any, statusCode int) {
	h.rw.WriteHeader(statusCode)

	if err := json.NewEncoder(h.rw).Encode(responceBody); err != nil {
		h.log.Error("write HTTP responce", zap.Error(err))
	}
}

func (h *HTTPResponceHandler) PanicResponce(p any, msg string) {
	statusCode := http.StatusInternalServerError
	err := fmt.Errorf("unexpected panic: %v", p)

	h.log.Error(msg, zap.Error(err))

	h.errorResponce(statusCode, err, msg)
}

func (h *HTTPResponceHandler) errorResponce(statusCode int, err error, msg string) {
	responce := ErorrReponce{
		Error:   err.Error(),
		Message: msg,
	}

	h.JSONResponce(responce, statusCode)
}
