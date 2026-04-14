package metric

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

// Metric описывает одну метрику для отправки на сервер.
type Metric struct {
	Name   string
	MType  string
	Gauge  float64
	Delta  int64
}
