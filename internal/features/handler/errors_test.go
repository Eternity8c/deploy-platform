package features_handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	core_logger "github.com/Eternity8c/deploy-platform/internal/core/logger"
	core_transport_request "github.com/Eternity8c/deploy-platform/internal/core/transport/request"
	core_transport_response "github.com/Eternity8c/deploy-platform/internal/core/transport/response"
	features_application "github.com/Eternity8c/deploy-platform/internal/features/application"
)

// nopLogger — логгер, который никуда не пишет: маппер логирует только в default,
// а файл на диске в тестах не нужен.
func nopLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantStatus   int
		wantFields   int
		wantContains string // чего ждём в message ("" — не проверяем)
		forbidden    string // чего в теле быть НЕ должно
	}{
		{
			name: "валидация: две претензии",
			err: features_application.ValidationErrors{
				{Field: "metadata.name", Message: "must be a lowercase DNS label"},
				{Field: "spec.port", Message: "must be between 1 and 65535"},
			},
			wantStatus:   http.StatusBadRequest,
			wantFields:   2,
			wantContains: "validation",
		},
		{
			name:         "валидация сквозь обёртку",
			err:          fmt.Errorf("store: %w", features_application.ValidationErrors{{Field: "spec.tag", Message: "must not be empty"}}),
			wantStatus:   http.StatusBadRequest,
			wantFields:   1,
			wantContains: "validation",
		},
		{
			name:         "ресурс уже есть",
			err:          fmt.Errorf("store: %w", features_application.ErrAlreadyExists),
			wantStatus:   http.StatusConflict,
			wantFields:   0,
			wantContains: "already exists",
		},
		{
			name:         "пустое тело",
			err:          core_transport_request.Error{Class: core_transport_request.EmptyError, Message: "request body is empty"},
			wantStatus:   http.StatusBadRequest,
			wantFields:   0,
			wantContains: "empty",
		},
		{
			name:         "битый JSON",
			err:          core_transport_request.Error{Class: core_transport_request.MalformedError, Message: "request body is not valid JSON"},
			wantStatus:   http.StatusBadRequest,
			wantFields:   0,
			wantContains: "not valid JSON",
		},
		{
			name:         "лишние данные в теле",
			err:          core_transport_request.Error{Class: core_transport_request.TrailingError, Message: "trailing"},
			wantStatus:   http.StatusBadRequest,
			wantFields:   0,
			wantContains: "exactly one JSON value",
		},
		{
			name: "неизвестное поле",
			err: core_transport_request.Error{
				Class:   core_transport_request.UnknownFieldError,
				Field:   "replicats",
				Message: "field is not part of the schema",
			},
			wantStatus:   http.StatusBadRequest,
			wantFields:   1,
			wantContains: "not part of the schema",
		},
		{
			// Главный кейс: клиент не должен узнать, что сломалось внутри.
			name:         "внутренняя ошибка не утекает",
			err:          errors.New("db exploded: connection refused to 10.0.0.5"),
			wantStatus:   http.StatusInternalServerError,
			wantFields:   0,
			wantContains: "internal server error",
			forbidden:    "db exploded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			writeError(nopLogger(), rec, tt.err)

			if rec.Code != tt.wantStatus {
				t.Fatalf("статус %d, ожидался %d", rec.Code, tt.wantStatus)
			}

			raw := rec.Body.Bytes()

			// форма: {"error": {...}} — обёртка обязана быть, иначе тело ошибки
			// неотличимо от тела ресурса
			var wrapped map[string]json.RawMessage
			if err := json.Unmarshal(raw, &wrapped); err != nil {
				t.Fatalf("ответ не JSON: %v (%s)", err, raw)
			}
			if _, ok := wrapped["error"]; !ok {
				t.Fatalf("в ответе нет обёртки error: %s", raw)
			}

			var body core_transport_response.Error
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Fatalf("разбор тела: %v", err)
			}

			if tt.wantContains != "" && !strings.Contains(strings.ToLower(body.Detail.Message), strings.ToLower(tt.wantContains)) {
				t.Fatalf("message %q не содержит %q", body.Detail.Message, tt.wantContains)
			}
			if tt.forbidden != "" && strings.Contains(string(raw), tt.forbidden) {
				t.Fatalf("в теле ответа утекло %q: %s", tt.forbidden, raw)
			}
			if len(body.Detail.Fields) != tt.wantFields {
				t.Fatalf("полей %d, ожидалось %d: %s", len(body.Detail.Fields), tt.wantFields, raw)
			}
			if tt.wantFields > 0 && body.Detail.Fields[0].Field == "" {
				t.Fatalf("поле пустое: %s", raw)
			}
		})
	}
}

// TestWriteErrorUnknownFieldKeepsName — имя поля обязан доехать до клиента:
// именно оно говорит, что чинить.
func TestWriteErrorUnknownFieldKeepsName(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(nopLogger(), rec, core_transport_request.Error{
		Class:   core_transport_request.UnknownFieldError,
		Field:   "replicats",
		Message: "field is not part of the schema",
	})

	var body core_transport_response.Error
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("разбор тела: %v", err)
	}
	if len(body.Detail.Fields) != 1 || body.Detail.Fields[0].Field != "replicats" {
		t.Fatalf("ожидалось поле replicats, получено: %s", rec.Body.String())
	}
}
