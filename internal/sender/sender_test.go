package sender_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/metric"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/sender"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/storage"
)

func TestHTTPSender_SendMetric_OK(t *testing.T) {
	store := storage.NewMemStorage()
	srv := httptest.NewServer(server.NewRouter(store))
	t.Cleanup(srv.Close)

	s := sender.NewHTTPSender(srv.Client(), srv.URL)
	ctx := context.Background()

	g := metric.Metric{Name: "G", MType: metric.TypeGauge, Gauge: 1.25}
	if err := s.SendMetric(ctx, g); err != nil {
		t.Fatal(err)
	}
	v, ok := store.GetGauge("G")
	if !ok || v != 1.25 {
		t.Fatalf("gauge stored: %v ok=%v", v, ok)
	}

	c := metric.Metric{Name: "C", MType: metric.TypeCounter, Delta: 3}
	if err := s.SendMetric(ctx, c); err != nil {
		t.Fatal(err)
	}
	n, ok := store.GetCounter("C")
	if !ok || n != 3 {
		t.Fatalf("counter stored: %d ok=%v", n, ok)
	}
}

func TestHTTPSender_SendMetric_NonOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	t.Cleanup(ts.Close)

	s := sender.NewHTTPSender(ts.Client(), ts.URL)
	err := s.SendMetric(context.Background(), metric.Metric{Name: "x", MType: metric.TypeGauge, Gauge: 1})
	if err == nil {
		t.Fatal("expected error")
	}
}
