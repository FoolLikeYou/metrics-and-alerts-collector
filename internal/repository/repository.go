package repository

import "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/storage"

// MetricsRepository — контракт чтения и записи метрик; совпадает с MemStorage для подмены в тестах.
type MetricsRepository = storage.MetricsStorage
