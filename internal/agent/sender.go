package agent

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func postMetric(ctx context.Context, client *http.Client, baseURL, mType, name, value string) error {
	base := strings.TrimSuffix(baseURL, "/")
	url := fmt.Sprintf("%s/update/%s/%s/%s", base, mType, name, value)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}
	return nil
}

func formatGauge(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}
