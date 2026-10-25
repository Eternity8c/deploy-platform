package features_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core_logger "github.com/Eternity8c/deploy-platform/internal/core/logger"
	core_transport_response "github.com/Eternity8c/deploy-platform/internal/core/transport/response"
	features_application "github.com/Eternity8c/deploy-platform/internal/features/application"
)

// validManifest — заведомо валидный запрос. Каждый тест ломает в нём одну вещь.
const validManifest = `{
	"apiVersion": "platform.deploy/v1",
	"kind": "Application",
	"metadata": {"name": "shop-api"},
	"spec": {
		"tenant": "shop",
		"image": "ghcr.io/o/shop-api",
		"tag": "3f2a1c9",
		"replicas": 2,
		"port": 8080,
		"env": {"LOG_LEVEL": "info"},
		"ingress": {"host": "demo.example.ts.net", "path": "/"}
	}
}`

type fakeStore struct {
	err   error
	calls int
	got   features_application.Application
}

func (f *fakeStore) Create(_ context.Context, app features_application.Application) (features_application.CreateResult, error) {
	f.calls++
	f.got = app

	if f.err != nil {
		return features_application.CreateResult{}, f.err
	}

	return features_application.CreateResult{ID: 1, Version: 1}, nil
}

func TestCreateApplication(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		storeErr     error
		wantStatus   int
		wantFields   int
		wantContains string
		forbidden    string
		wantStoreHit bool
	}{
		{
			name:         "валидный манифест",
			body:         validManifest,
			wantStatus:   http.StatusCreated,
			wantStoreHit: true,
		},
		{
			name:         "пустое тело",
			body:         ``,
			wantStatus:   http.StatusBadRequest,
			wantContains: "empty",
		},
		{
			name:         "битый JSON",
			body:         `{"apiVersion":`,
			wantStatus:   http.StatusBadRequest,
			wantContains: "not valid JSON",
		},
		{
			name:         "опечатка в имени поля",
			body:         strings.Replace(validManifest, `"replicas": 2`, `"replicats": 2`, 1),
			wantStatus:   http.StatusBadRequest,
			wantFields:   1,
			wantContains: "not part of the schema",
		},
		{
			name:         "невалидные значения: имя и порт",
			body:         strings.Replace(strings.Replace(validManifest, `"shop-api"`, `"Shop_API"`, 1), `8080`, `99999`, 1),
			wantStatus:   http.StatusBadRequest,
			wantFields:   2,
			wantContains: "validation",
		},
		{
			name:         "store: уже существует",
			body:         validManifest,
			storeErr:     fmt.Errorf("create application: %w", features_application.ErrAlreadyExists),
			wantStatus:   http.StatusConflict,
			wantContains: "already exists",
			wantStoreHit: true,
		},
		{
			name:         "store: внутренняя ошибка не утекает",
			body:         validManifest,
			storeErr:     fmt.Errorf("insert into applications: %w", errDBExploded),
			wantStatus:   http.StatusInternalServerError,
			wantContains: "internal server error",
			forbidden:    "insert into applications",
			wantStoreHit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{err: tt.storeErr}
			handler := NewApplicationHandler(store)

			r := httptest.NewRequest(http.MethodPost, "/api/v1/applications", strings.NewReader(tt.body))
			// хендлер берёт логгер из контекста и паникует без него
			r.Header.Set("X-Request-ID", testRequestID)
			r = r.WithContext(core_logger.WithLogger(r.Context(), nopLogger()))
			rec := httptest.NewRecorder()

			handler.CreateApplication(rec, r)

			if rec.Code != tt.wantStatus {
				t.Fatalf("статус %d, ожидался %d (%s)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if !tt.wantStoreHit && store.calls != 0 {
				t.Fatalf("store вызван %d раз, а запрос должен был отвалиться раньше", store.calls)
			}
			if tt.wantStoreHit && store.calls != 1 {
				t.Fatalf("store вызван %d раз, ожидался 1", store.calls)
			}

			if rec.Code == http.StatusCreated {
				return // успешный ответ проверяем отдельным тестом
			}

			var body core_transport_response.Error
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("ответ не JSON: %v (%s)", err, rec.Body.String())
			}
			if tt.wantContains != "" && !strings.Contains(strings.ToLower(body.Detail.Message), strings.ToLower(tt.wantContains)) {
				t.Fatalf("message %q не содержит %q", body.Detail.Message, tt.wantContains)
			}
			if tt.forbidden != "" && strings.Contains(rec.Body.String(), tt.forbidden) {
				t.Fatalf("в теле утекло %q: %s", tt.forbidden, rec.Body.String())
			}
			if len(body.Detail.Fields) != tt.wantFields {
				t.Fatalf("полей %d, ожидалось %d: %s", len(body.Detail.Fields), tt.wantFields, rec.Body.String())
			}
			if body.Detail.RequestID != testRequestID {
				t.Fatalf("request_id %q, ожидался %q", body.Detail.RequestID, testRequestID)
			}
		})
	}
}

// errDBExploded — текст, который не имеет права доехать до клиента.
var errDBExploded = fmt.Errorf("connection refused to 10.0.0.5")

// TestCreateApplicationCreatedBody проверяет, что успех не обёрнут в error-форму:
// клиент различает успех и ошибку по форме, а не по догадке.
func TestCreateApplicationCreatedBody(t *testing.T) {
	store := &fakeStore{}
	handler := NewApplicationHandler(store)

	r := httptest.NewRequest(http.MethodPost, "/api/v1/applications", strings.NewReader(validManifest))
	r.Header.Set("X-Request-ID", testRequestID)
	r = r.WithContext(core_logger.WithLogger(r.Context(), nopLogger()))
	rec := httptest.NewRecorder()

	handler.CreateApplication(rec, r)

	if rec.Code != http.StatusCreated {
		t.Fatalf("статус %d, ожидался 201 (%s)", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("разбор тела: %v", err)
	}
	if _, ok := body["error"]; ok {
		t.Fatalf("успешный ответ обёрнут в error: %s", rec.Body.String())
	}
	if _, ok := body["id"]; !ok {
		t.Fatalf("в ответе нет id: %s", rec.Body.String())
	}
	// успешный ответ не должен тащить request_id в теле — он в заголовке
	if _, ok := body["request_id"]; ok {
		t.Fatalf("успешный ответ содержит request_id: %s", rec.Body.String())
	}
	// tenant не пришёл в запросе — значит проверяем, что дефолт домена доехал в store
	if store.got.Spec.Tenant != "shop" {
		t.Fatalf("tenant в store = %q, ожидался shop", store.got.Spec.Tenant)
	}
}
