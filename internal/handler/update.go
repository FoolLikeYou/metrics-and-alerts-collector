package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

const (
	metricCounter = "counter"
	metricGauge   = "gauge"
)

// PostUpdate обрабатывает POST /update/{type}/{name}/{value} (старый контракт).
// JSON-обновления — отдельный маршрут POST /update, см. PostJSONUpdate.
func PostUpdate(store repository.MetricsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")
		valueStr := chi.URLParam(r, "value")

		if name == "" {
			http.NotFound(w, r)
			return
		}

		switch metricType {
		case metricGauge:
			v, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				http.Error(w, "invalid gauge value", http.StatusBadRequest)
				return
			}
			store.SetGauge(name, v)
		case metricCounter:
			v, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid counter value", http.StatusBadRequest)
				return
			}
			store.AddCounter(name, v)
		default:
			http.Error(w, "invalid metric type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
