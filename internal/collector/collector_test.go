package collector_test

import (
	"math/rand"
	"runtime"
	"testing"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/collector"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/metric"
)

func requiredGaugeNames() map[string]struct{} {
	names := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys",
		"LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys",
		"Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs",
		"StackInuse", "StackSys", "Sys", "TotalAlloc",
		"RandomValue",
	}
	m := make(map[string]struct{}, len(names))
	for _, n := range names {
		m[n] = struct{}{}
	}
	return m
}

func TestCollectRuntimeGauges_Names(t *testing.T) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	rnd := rand.New(rand.NewSource(42))
	got := collector.CollectRuntimeGauges(&ms, rnd)

	want := requiredGaugeNames()
	seen := make(map[string]struct{}, len(got))
	for _, mtr := range got {
		if mtr.MType != metric.TypeGauge {
			t.Fatalf("unexpected type for %s: %s", mtr.Name, mtr.MType)
		}
		seen[mtr.Name] = struct{}{}
	}
	for name := range want {
		if _, ok := seen[name]; !ok {
			t.Fatalf("missing metric: %s", name)
		}
	}
	if len(seen) != len(want) {
		t.Fatalf("extra or duplicate metrics: got %d want %d", len(seen), len(want))
	}
}
