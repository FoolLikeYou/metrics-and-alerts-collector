package handler

import (
	"fmt"
	"net/http"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

func ListHandler(storage repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gauges, counters := storage.GetAll()

		w.Header().Set("Content-Type", "text/plain")

		for name, value := range gauges {
			fmt.Fprintf(w, "gauge %s = %f\n", name, value)
		}

		for name, value := range counters {
			fmt.Fprintf(w, "counter %s = %d\n", name, value)
		}
	}
}
