package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	models "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/metric"
)

// HTTPSender отправляет метрики на HTTP-сервер.
type HTTPSender struct {
	client  *http.Client
	baseURL string
}

// NewHTTPSender создаёт отправитель. baseURL без завершающего слэша, например http://localhost:8080.
func NewHTTPSender(client *http.Client, baseURL string) *HTTPSender {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPSender{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

// SendMetric выполняет POST /update с телом JSON (application/json).
func (s *HTTPSender) SendMetric(ctx context.Context, m metric.Metric) error {
	body, err := jsonBody(m)
	if err != nil {
		return err
	}

	url := s.baseURL + "/update"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func jsonBody(m metric.Metric) ([]byte, error) {
	switch m.MType {
	case metric.TypeGauge:
		v := m.Gauge
		return json.Marshal(models.Metrics{
			ID:    m.Name,
			MType: models.Gauge,
			Value: &v,
		})
	case metric.TypeCounter:
		d := m.Delta
		return json.Marshal(models.Metrics{
			ID:    m.Name,
			MType: models.Counter,
			Delta: &d,
		})
	default:
		return nil, fmt.Errorf("unknown metric type: %s", m.MType)
	}
}
