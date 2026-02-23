package handler

import (
	models "metrics-and-alerts-collector/internal/model"
	"metrics-and-alerts-collector/internal/repository"
	"net/http"
	"strconv"
	"strings"
)

func UpdateHandler(storage repository.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// ожидаем: /update/<type>/<name>/<value>
		path := strings.TrimPrefix(r.URL.Path, "/update/")
		parts := strings.Split(path, "/")
		if len(parts) != 3 {
			// нет имени метрики / нет значения / неправильный формат
			w.WriteHeader(http.StatusNotFound)
			return
		}

		mType := parts[0]
		name := parts[1]
		valueStr := parts[2]

		if name == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch mType {
		case models.Gauge:
			v, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.UpdateGauge(name, v)
			w.WriteHeader(http.StatusOK)

		case models.Counter:
			v, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.UpdateCounter(name, v)
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}
