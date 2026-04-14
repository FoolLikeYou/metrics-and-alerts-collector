package repository

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

// Интерфейс хранилища
type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, delta int64)
	GetAll() (map[string]float64, map[string]int64)
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemStorage) GetAll() (map[string]float64, map[string]int64) {
	return m.gauges, m.counters
}

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.gauges[name] = value
}

func (m *MemStorage) UpdateCounter(name string, delta int64) {
	m.counters[name] += delta
}
