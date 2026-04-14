package agent_test

import (
	"context"
	"math/rand"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/metric"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/sender"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/storage"
)

type mockSender struct {
	mu    sync.Mutex
	calls []metric.Metric
}

func (m *mockSender) SendMetric(_ context.Context, met metric.Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, met)
	return nil
}

func (m *mockSender) snapshot() []metric.Metric {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]metric.Metric, len(m.calls))
	copy(out, m.calls)
	return out
}

func TestAgent_ReportSendsGaugesAndPollCount(t *testing.T) {
	ms := &mockSender{}
	rnd := rand.New(rand.NewSource(1))
	cfg := agent.Config{
		PollInterval:   15 * time.Millisecond,
		ReportInterval: 40 * time.Millisecond,
		ServerURL:      "http://unused",
	}
	a := agent.New(cfg, ms, rnd)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	go func() { _ = a.Run(ctx) }()

	time.Sleep(100 * time.Millisecond)

	got := ms.snapshot()
	if len(got) == 0 {
		t.Fatal("expected some SendMetric calls")
	}

	seen := make(map[string]int)
	for _, m := range got {
		seen[m.Name]++
	}
	if seen["RandomValue"] == 0 {
		t.Fatal("expected RandomValue metric")
	}
	if seen["PollCount"] == 0 {
		t.Fatal("expected PollCount counter sends")
	}
	var pollDelta int64
	for _, m := range got {
		if m.Name == "PollCount" && m.MType == metric.TypeCounter {
			pollDelta += m.Delta
		}
	}
	if pollDelta == 0 {
		t.Fatal("expected positive PollCount deltas")
	}
}

// При REPORT_INTERVAL < POLL_INTERVAL без начального pollOnce первые отчёты уходят с пустым lastGauges.
func TestAgent_ReportFasterThanPollStillPersistsToServer(t *testing.T) {
	store := storage.NewMemStorage()
	srv := httptest.NewServer(server.NewRouter(store))
	t.Cleanup(srv.Close)

	s := sender.NewHTTPSender(srv.Client(), srv.URL)
	rnd := rand.New(rand.NewSource(7))
	cfg := agent.Config{
		PollInterval:   200 * time.Millisecond,
		ReportInterval: 12 * time.Millisecond,
		ServerURL:      srv.URL,
	}
	a := agent.New(cfg, s, rnd)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	go func() { _ = a.Run(ctx) }()

	time.Sleep(25 * time.Millisecond)
	if _, ok := store.GetGauge("Alloc"); !ok {
		t.Fatal("expected runtime gauge on server before first poll ticker (initial collect + JSON report)")
	}
}
