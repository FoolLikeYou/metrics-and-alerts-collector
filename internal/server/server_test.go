package server_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model"
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

func TestNewRouter_GzipJSONUpdateAndValue(t *testing.T) {
	store := storage.NewMemStorage()
	mux := server.NewRouter(store)

	raw, _ := json.Marshal(models.Metrics{ID: "G", MType: models.Gauge, Value: ptrFloat(9)})
	var gzbuf bytes.Buffer
	gw := gzip.NewWriter(&gzbuf)
	_, _ = gw.Write(raw)
	_ = gw.Close()

	req := httptest.NewRequest(http.MethodPost, "/update", &gzbuf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("update status: %d body %s", rr.Code, rr.Body.String())
	}
	v, ok := store.GetGauge("G")
	if !ok || v != 9 {
		t.Fatalf("gauge: %v ok=%v", v, ok)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBufferString(`{"id":"G","type":"gauge"}`))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Accept-Encoding", "gzip")
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("value status: %d", rr2.Code)
	}
	if rr2.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("want gzip response, got %q", rr2.Header().Get("Content-Encoding"))
	}
	zr, err := gzip.NewReader(rr2.Body)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(zr)
	_ = zr.Close()
	if err != nil {
		t.Fatal(err)
	}
	var got models.Metrics
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != "G" || got.Value == nil || *got.Value != 9 {
		t.Fatalf("value json: %+v", got)
	}
}

func ptrFloat(f float64) *float64 { return &f }
