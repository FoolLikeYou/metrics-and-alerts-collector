package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	models "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/storage"
)

func TestPostJSONUpdate_OK(t *testing.T) {
	store := storage.NewMemStorage()
	mux := server.NewRouter(store)

	body := `{"id":"x","type":"gauge","value":1.5}`
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	v, ok := store.GetGauge("x")
	if !ok || v != 1.5 {
		t.Fatalf("gauge: %v ok=%v", v, ok)
	}

	body2 := `{"id":"c","type":"counter","delta":3}`
	req2 := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(body2))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("counter status %d", rr2.Code)
	}
	n, ok := store.GetCounter("c")
	if !ok || n != 3 {
		t.Fatalf("counter: %d ok=%v", n, ok)
	}
}

func TestPostJSONValue_OK(t *testing.T) {
	store := storage.NewMemStorage()
	store.SetGauge("g", 2.25)
	store.AddCounter("n", 10)

	mux := server.NewRouter(store)

	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBufferString(`{"id":"g","type":"gauge"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); strings.TrimSpace(strings.SplitN(ct, ";", 2)[0]) != "application/json" {
		t.Fatalf("Content-Type: %q", ct)
	}
	var got models.Metrics
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != "g" || got.MType != models.Gauge || got.Value == nil || *got.Value != 2.25 {
		t.Fatalf("response: %+v", got)
	}
}
