package storage

import (
	"sort"
	"strconv"
	"sync"
)

// MetricListItem — строка списка метрик для HTML и отладки.
type MetricListItem struct {
	Name  string
	MType string
	Value string
}

// MetricsStorage — контракт чтения и записи метрик (MemStorage, моки).
type MetricsStorage interface {
	SetGauge(name string, value float64)
	AddCounter(name string, delta int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	ListMetrics() []MetricListItem
}

// MemStorage хранит метрики в памяти процесса.
type MemStorage struct {
	mu       sync.Mutex
	gauges   map[string]float64
	counters map[string]int64
}

// NewMemStorage создаёт пустое in-memory хранилище.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// SetGauge сохраняет gauge; новое значение замещает предыдущее.
func (m *MemStorage) SetGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

// AddCounter увеличивает counter на delta относительно уже известного значения.
func (m *MemStorage) AddCounter(name string, delta int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += delta
}

// GetGauge возвращает последнее значение gauge.
func (m *MemStorage) GetGauge(name string) (float64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.gauges[name]
	return v, ok
}

// GetCounter возвращает накопленное значение counter.
func (m *MemStorage) GetCounter(name string) (int64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.counters[name]
	return v, ok
}

// ListMetrics возвращает отсортированный снимок всех метрик.
func (m *MemStorage) ListMetrics() []MetricListItem {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]MetricListItem, 0, len(m.gauges)+len(m.counters))
	for n, v := range m.gauges {
		out = append(out, MetricListItem{
			Name:  n,
			MType: "gauge",
			Value: strconv.FormatFloat(v, 'f', -1, 64),
		})
	}
	for n, v := range m.counters {
		out = append(out, MetricListItem{
			Name:  n,
			MType: "counter",
			Value: strconv.FormatInt(v, 10),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].MType < out[j].MType
	})
	return out
}
