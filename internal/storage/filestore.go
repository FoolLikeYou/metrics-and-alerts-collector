package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"

	models "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model"
)

// SaveToJSONFile сохраняет снимок метрик в JSON-массив (атомарная запись).
func SaveToJSONFile(path string, m *MemStorage) error {
	raw, err := m.snapshotJSON()
	if err != nil {
		return err
	}
	return atomicWriteFile(path, raw)
}

// LoadFromJSONFile загружает метрики из файла; при отсутствии файла — без ошибки.
func LoadFromJSONFile(path string, m *MemStorage) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(bytes.TrimSpace(b)) == 0 {
		return nil
	}
	var items []models.Metrics
	if err := json.Unmarshal(b, &items); err != nil {
		return err
	}
	return m.replaceFromMetrics(items)
}

func (m *MemStorage) snapshotJSON() ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := make([]models.Metrics, 0, len(m.gauges)+len(m.counters))
	for n, v := range m.gauges {
		vv := v
		items = append(items, models.Metrics{ID: n, MType: models.Gauge, Value: &vv})
	}
	for n, v := range m.counters {
		vv := v
		items = append(items, models.Metrics{ID: n, MType: models.Counter, Delta: &vv})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].ID != items[j].ID {
			return items[i].ID < items[j].ID
		}
		return items[i].MType < items[j].MType
	})
	return json.MarshalIndent(items, "", "  ")
}

func (m *MemStorage) replaceFromMetrics(items []models.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges = make(map[string]float64)
	m.counters = make(map[string]int64)
	for _, it := range items {
		switch it.MType {
		case models.Gauge:
			if it.Value == nil {
				return errors.New("gauge without value: " + it.ID)
			}
			m.gauges[it.ID] = *it.Value
		case models.Counter:
			if it.Delta == nil {
				return errors.New("counter without delta: " + it.ID)
			}
			m.counters[it.ID] = *it.Delta
		default:
			return errors.New("unknown metric type: " + it.MType)
		}
	}
	return nil
}

func atomicWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
