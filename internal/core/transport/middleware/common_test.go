package core_transport_middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestRequestID(t *testing.T) {
	const clientUUID = "3f2a1c9e-1234-4000-8000-000000000000"

	tests := []struct {
		name     string
		clientID string
		wantKept bool // true — клиентский id должен сохраниться как есть
	}{
		{name: "клиент не прислал id", clientID: ""},
		{name: "клиент прислал валидный uuid", clientID: clientUUID, wantKept: true},
		{name: "клиент прислал мусор", clientID: "not-a-uuid"},
		{name: "клиент прислал строку с json", clientID: `{"level":"error"}`},
		{name: "клиент прислал обрезанный uuid", clientID: "3f2a1c9e-1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seenByHandler string

			handler := RequestID()(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				// хендлер получает id из заголовка запроса
				seenByHandler = r.Header.Get(requestIDHeader)
			}))

			r := httptest.NewRequest(http.MethodGet, "/api/v1/applications", nil)
			if tt.clientID != "" {
				r.Header.Set(requestIDHeader, tt.clientID)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, r)

			got := rec.Header().Get(requestIDHeader)
			if got == "" {
				t.Fatal("в ответе нет заголовка X-Request-ID")
			}
			if _, err := uuid.Parse(got); err != nil {
				t.Fatalf("id не является uuid: %q (%v)", got, err)
			}
			if seenByHandler != got {
				t.Fatalf("хендлер увидел %q, а клиенту ушло %q", seenByHandler, got)
			}

			if tt.wantKept && got != tt.clientID {
				t.Fatalf("клиентский id %q затёрт на %q", tt.clientID, got)
			}
			if !tt.wantKept && got == tt.clientID && tt.clientID != "" {
				t.Fatalf("невалидный клиентский id %q принят как есть", tt.clientID)
			}
		})
	}
}

// TestRequestIDGeneratedIsUnique: два запроса без id обязаны получить разные id,
// иначе корреляция логов превращается в мусор.
func TestRequestIDGeneratedIsUnique(t *testing.T) {
	handler := RequestID()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/", nil))

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/", nil))

	if first.Header().Get(requestIDHeader) == second.Header().Get(requestIDHeader) {
		t.Fatal("два запроса получили одинаковый id")
	}
}
