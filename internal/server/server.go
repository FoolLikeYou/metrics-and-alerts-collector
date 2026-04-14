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
	r.Use(gzipMiddleware)
	// POST /update в двух форматах (оба должны сохраняться):
	// 1) /update/{type}/{name}/{value} — как в ранних инкрементах, тело пустое, Content-Type часто text/plain.
	// 2) /update и /update/ — JSON по схеме internal/model.Metrics, Content-Type application/json.
	// Сначала маршрут с path-параметрами: иначе chi не отличит /update/gauge/... от префикса /update.
	r.Post("/update/{type}/{name}/{value}", handler.PostUpdate(store))
	// Автотесты (resty) часто шлют POST /update/ и POST /value/ с завершающим слэшем.
	r.Post("/update", handler.PostJSONUpdate(store))
	r.Post("/update/", handler.PostJSONUpdate(store))
	r.Post("/value", handler.PostJSONValue(store))
	r.Post("/value/", handler.PostJSONValue(store))
	r.Get("/value/{type}/{name}", handler.GetValue(store))
	r.Get("/", handler.ListHTML(store))
	return r
}
