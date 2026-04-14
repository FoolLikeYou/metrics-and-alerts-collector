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

func TestPostUpdate_StatusCodes(t *testing.T) {
	store := storage.NewMemStorage()
	mux := server.NewRouter(store)

	post := func(path string) int {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("Content-Type", "text/plain")
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		return rr.Code
	}

	cases := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{name: "gauge ok", path: "/update/gauge/m/1.5", wantStatus: http.StatusOK},
		{name: "counter ok", path: "/update/counter/c/2", wantStatus: http.StatusOK},
		{name: "no name", path: "/update/gauge", wantStatus: http.StatusNotFound},
		{name: "bad type", path: "/update/foo/x/1", wantStatus: http.StatusBadRequest},
		{name: "bad gauge", path: "/update/gauge/x/abc", wantStatus: http.StatusBadRequest},
		{name: "bad counter", path: "/update/counter/x/1.2", wantStatus: http.StatusBadRequest},
		{name: "extra path", path: "/update/gauge/x/1/extra", wantStatus: http.StatusNotFound},
		{name: "wrong root", path: "/foo/gauge/x/1", wantStatus: http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := post(tc.path); got != tc.wantStatus {
				t.Fatalf("status: got %d want %d", got, tc.wantStatus)
			}
		})
	}
}

func TestPostUpdate_EmptyMetricName(t *testing.T) {
	store := storage.NewMemStorage()
	h := handler.PostUpdate(store)

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "")
	rctx.URLParams.Add("value", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("empty name: got %d want %d", rr.Code, http.StatusNotFound)
	}
}

func TestPostUpdate_StoresMetrics(t *testing.T) {
	store := storage.NewMemStorage()
	mux := chi.NewRouter()
	mux.Post("/update/{type}/{name}/{value}", handler.PostUpdate(store))

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/G/3", nil)
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("gauge: %d", rr.Code)
	}
	v, ok := store.GetGauge("G")
	if !ok || v != 3 {
		t.Fatalf("GetGauge: got %v ok=%v", v, ok)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/update/counter/C/10", nil)
	req2.Header.Set("Content-Type", "text/plain")
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("counter: %d", rr2.Code)
	}
	c, ok := store.GetCounter("C")
	if !ok || c != 10 {
		t.Fatalf("GetCounter: got %v ok=%v", c, ok)
	}

	req3 := httptest.NewRequest(http.MethodPost, "/update/counter/C/5", nil)
	req3.Header.Set("Content-Type", "text/plain")
	rr3 := httptest.NewRecorder()
	mux.ServeHTTP(rr3, req3)
	c2, _ := store.GetCounter("C")
	if c2 != 15 {
		t.Fatalf("counter sum: got %d want 15", c2)
	}
}

func TestPostUpdate_GET_MethodNotAllowed(t *testing.T) {
	store := storage.NewMemStorage()
	mux := server.NewRouter(store)
	req := httptest.NewRequest(http.MethodGet, "/update/gauge/x/1", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET: want 405 got %d", rr.Code)
	}
}
