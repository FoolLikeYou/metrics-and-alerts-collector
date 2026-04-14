package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/storage"
)

func TestNewRouter_AcceptsUpdate(t *testing.T) {
	store := storage.NewMemStorage()
	mux := server.NewRouter(store)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/X/2", nil)
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
	v, ok := store.GetGauge("X")
	if !ok || v != 2 {
		t.Fatalf("stored: %v ok=%v", v, ok)
	}
}
