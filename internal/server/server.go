package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/handler"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

// NewRouter возвращает HTTP-обработчик сервера метрик (chi).
func NewRouter(store repository.MetricsRepository) http.Handler {
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handler.PostUpdate(store))
	r.Get("/value/{type}/{name}", handler.GetValue(store))
	r.Get("/", handler.ListHTML(store))
	return r
}
