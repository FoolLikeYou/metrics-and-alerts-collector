package agent

import (
	"context"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/collector"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/metric"
)

// MetricSender отправляет одну метрику (удобно мокать в тестах).
type MetricSender interface {
	SendMetric(ctx context.Context, m metric.Metric) error
}

// Config задаёт интервалы и адрес сервера.
type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerURL      string
}

// DefaultConfig — poll 2s, report 10s, http://localhost:8080.
func DefaultConfig() Config {
	return Config{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		ServerURL:      "http://localhost:8080",
	}
}

// Agent циклически опрашивает runtime и отправляет метрики на сервер.
type Agent struct {
	cfg    Config
	sender MetricSender
	rnd    *rand.Rand

	mu            sync.Mutex
	lastGauges    []metric.Metric
	pollsSinceRep int64
}

// New создаёт агента. rnd должен быть ненулевым (например rand.New(...)).
func New(cfg Config, s MetricSender, rnd *rand.Rand) *Agent {
	return &Agent{cfg: cfg, sender: s, rnd: rnd}
}

// Run блокируется до отмены ctx или ошибки отправки.
func (a *Agent) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Иначе при REPORT_INTERVAL < POLL_INTERVAL первые тики report срабатывают до poll.C
	// и уходят пустые — на сервере нет метрик, автотесты получают 404 и «нет изменения».
	a.pollOnce()

	poll := time.NewTicker(a.cfg.PollInterval)
	report := time.NewTicker(a.cfg.ReportInterval)
	defer poll.Stop()
	defer report.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-poll.C:
			a.pollOnce()
		case <-report.C:
			if err := a.report(ctx); err != nil {
				return err
			}
		}
	}
}

func (a *Agent) pollOnce() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	gauges := collector.CollectRuntimeGauges(&ms, a.rnd)

	a.mu.Lock()
	a.lastGauges = gauges
	a.pollsSinceRep++
	a.mu.Unlock()
}

func (a *Agent) report(ctx context.Context) error {
	a.mu.Lock()
	gauges := append([]metric.Metric(nil), a.lastGauges...)
	polls := a.pollsSinceRep
	a.pollsSinceRep = 0
	a.mu.Unlock()

	for i := range gauges {
		if err := a.sender.SendMetric(ctx, gauges[i]); err != nil {
			return err
		}
	}
	if polls != 0 {
		pc := metric.Metric{Name: "PollCount", MType: metric.TypeCounter, Delta: polls}
		if err := a.sender.SendMetric(ctx, pc); err != nil {
			return err
		}
	}
	return nil
}
