package sender

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

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

// SendMetric выполняет POST /update/{type}/{name}/{value}.
func (s *HTTPSender) SendMetric(ctx context.Context, m metric.Metric) error {
	var value string
	switch m.MType {
	case metric.TypeGauge:
		value = strconv.FormatFloat(m.Gauge, 'f', -1, 64)
	case metric.TypeCounter:
		value = strconv.FormatInt(m.Delta, 10)
	default:
		return fmt.Errorf("unknown metric type: %s", m.MType)
	}

	url := fmt.Sprintf("%s/update/%s/%s/%s", s.baseURL, m.MType, m.Name, value)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

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
