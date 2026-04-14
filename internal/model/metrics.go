package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metrics — формат JSON для обмена с сервером (POST /update, POST /value).
// Delta и Value через указатели, чтобы отличать 0 от «поле не передано».
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
