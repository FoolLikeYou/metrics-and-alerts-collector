package agent_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/metric"
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
