package collector

import (
	"math/rand"
	"runtime"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/metric"
)

// CollectRuntimeGauges собирает gauge-метрики из runtime.MemStats и RandomValue (rnd не должен быть nil).
func CollectRuntimeGauges(ms *runtime.MemStats, rnd *rand.Rand) []metric.Metric {
	out := make([]metric.Metric, 0, 32)

	add := func(name string, v float64) {
		out = append(out, metric.Metric{Name: name, MType: metric.TypeGauge, Gauge: v})
	}

	add("Alloc", float64(ms.Alloc))
	add("BuckHashSys", float64(ms.BuckHashSys))
	add("Frees", float64(ms.Frees))
	add("GCCPUFraction", ms.GCCPUFraction)
	add("GCSys", float64(ms.GCSys))
	add("HeapAlloc", float64(ms.HeapAlloc))
	add("HeapIdle", float64(ms.HeapIdle))
	add("HeapInuse", float64(ms.HeapInuse))
	add("HeapObjects", float64(ms.HeapObjects))
	add("HeapReleased", float64(ms.HeapReleased))
	add("HeapSys", float64(ms.HeapSys))
	add("LastGC", float64(ms.LastGC))
	add("Lookups", float64(ms.Lookups))
	add("MCacheInuse", float64(ms.MCacheInuse))
	add("MCacheSys", float64(ms.MCacheSys))
	add("MSpanInuse", float64(ms.MSpanInuse))
	add("MSpanSys", float64(ms.MSpanSys))
	add("Mallocs", float64(ms.Mallocs))
	add("NextGC", float64(ms.NextGC))
	add("NumForcedGC", float64(ms.NumForcedGC))
	add("NumGC", float64(ms.NumGC))
	add("OtherSys", float64(ms.OtherSys))
	add("PauseTotalNs", float64(ms.PauseTotalNs))
	add("StackInuse", float64(ms.StackInuse))
	add("StackSys", float64(ms.StackSys))
	add("Sys", float64(ms.Sys))
	add("TotalAlloc", float64(ms.TotalAlloc))

	add("RandomValue", rnd.Float64())

	return out
}
