package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/handler"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

// NewRouter возвращает HTTP-обработчик сервера метрик (chi).
func NewRouter(store repository.MetricsRepository) http.Handler {
	logger := zap.Must(zap.NewProduction())
	r := chi.NewRouter()
	r.Use(loggingMiddleware(logger))
	r.Post("/update/{type}/{name}/{value}", handler.PostUpdate(store))
	r.Post("/update", handler.PostJSONUpdate(store))
	r.Post("/value", handler.PostJSONValue(store))
	r.Get("/value/{type}/{name}", handler.GetValue(store))
	r.Get("/", handler.ListHTML(store))
	return r
}
