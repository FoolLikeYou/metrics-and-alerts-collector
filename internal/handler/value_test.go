package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/handler"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/storage"
)

func TestGetValue_OK(t *testing.T) {
	store := storage.NewMemStorage()
	store.SetGauge("g", 3.5)
	store.AddCounter("c", 7)

	mux := server.NewRouter(store)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/g", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("gauge status: %d", rr.Code)
	}
	if got := rr.Body.String(); got != "3.5" {
		t.Fatalf("gauge body: %q", got)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/value/counter/c", nil)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("counter status: %d", rr2.Code)
	}
	if got := rr2.Body.String(); got != "7" {
		t.Fatalf("counter body: %q", got)
	}
}

func TestGetValue_NotFound(t *testing.T) {
	store := storage.NewMemStorage()
	mux := server.NewRouter(store)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/missing", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("unknown gauge: %d", rr.Code)
	}
}

func TestGetValue_InvalidType(t *testing.T) {
	store := storage.NewMemStorage()
	h := handler.GetValue(store)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "histogram")
	rctx.URLParams.Add("name", "x")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("invalid type: %d", rr.Code)
	}
}
