package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	models "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

const maxJSONBody = 1 << 20

func isJSONContentType(ct string) bool {
	ct = strings.TrimSpace(strings.SplitN(ct, ";", 2)[0])
	return ct == "application/json"
}

// PostJSONUpdate обрабатывает POST /update (и /update/) с телом application/json.
// Path-формат /update/{type}/{name}/{value} обрабатывается в PostUpdate.
func PostJSONUpdate(store repository.MetricsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "" && !isJSONContentType(ct) {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		var m models.Metrics
		if err := decodeJSONMetric(r.Body, &m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if m.ID == "" {
			http.Error(w, "missing metric id", http.StatusBadRequest)
			return
		}

		var out models.Metrics
		switch m.MType {
		case models.Gauge:
			if m.Value == nil {
				http.Error(w, "gauge requires value", http.StatusBadRequest)
				return
			}
			store.SetGauge(m.ID, *m.Value)
			v, _ := store.GetGauge(m.ID)
			out = models.Metrics{ID: m.ID, MType: models.Gauge, Value: &v}
		case models.Counter:
			if m.Delta == nil {
				http.Error(w, "counter requires delta", http.StatusBadRequest)
				return
			}
			store.AddCounter(m.ID, *m.Delta)
			c, _ := store.GetCounter(m.ID)
			out = models.Metrics{ID: m.ID, MType: models.Counter, Delta: &c}
		default:
			http.Error(w, "invalid metric type", http.StatusBadRequest)
			return
		}

		raw, err := json.Marshal(out)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(raw)
	}
}

// PostJSONValue обрабатывает POST /value: запрос с id и type, ответ — полный JSON метрики.
func PostJSONValue(store repository.MetricsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "" && !isJSONContentType(ct) {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		var req models.Metrics
		if err := decodeJSONMetric(r.Body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.ID == "" {
			http.Error(w, "missing metric id", http.StatusBadRequest)
			return
		}

		out := models.Metrics{ID: req.ID, MType: req.MType}

		switch req.MType {
		case models.Gauge:
			v, ok := store.GetGauge(req.ID)
			if !ok {
				http.Error(w, "unknown metric", http.StatusNotFound)
				return
			}
			out.Value = &v
		case models.Counter:
			c, ok := store.GetCounter(req.ID)
			if !ok {
				http.Error(w, "unknown metric", http.StatusNotFound)
				return
			}
			out.Delta = &c
		default:
			http.Error(w, "invalid metric type", http.StatusBadRequest)
			return
		}

		raw, err := json.Marshal(out)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(raw)
	}
}

func decodeJSONMetric(r io.Reader, dst *models.Metrics) error {
	b, err := io.ReadAll(io.LimitReader(r, maxJSONBody))
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(b)) == 0 {
		return errors.New("empty body")
	}
	return json.Unmarshal(b, dst)
}
