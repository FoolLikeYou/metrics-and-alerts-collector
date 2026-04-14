package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

// GetValue обрабатывает GET /value/{type}/{name} — текстовое значение метрики.
func GetValue(store repository.MetricsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")

		if name == "" {
			http.NotFound(w, r)
			return
		}

		switch metricType {
		case metricGauge:
			v, ok := store.GetGauge(name)
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(strconv.FormatFloat(v, 'f', -1, 64)))
		case metricCounter:
			v, ok := store.GetCounter(name)
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(strconv.FormatInt(v, 10)))
		default:
			http.Error(w, "invalid metric type", http.StatusBadRequest)
			return
		}
	}
}
