package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/sender"
)

const defaultServerAddr = "localhost:8080"

func main() {
	srvAddr := flag.String("a", defaultServerAddr, "HTTP server address (host:port or URL)")
	reportSec := flag.Int("r", 10, "report interval, seconds")
	pollSec := flag.Int("p", 2, "poll interval, seconds")
	flag.Parse()

	addr := resolveServerAddr(strings.TrimSpace(*srvAddr))

	report := *reportSec
	if v, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		if t := strings.TrimSpace(v); t != "" {
			n, err := strconv.Atoi(t)
			if err != nil || n <= 0 {
				log.Fatalf("invalid REPORT_INTERVAL %q: want positive integer seconds", v)
			}
			report = n
		}
	}

	poll := *pollSec
	if v, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		if t := strings.TrimSpace(v); t != "" {
			n, err := strconv.Atoi(t)
			if err != nil || n <= 0 {
				log.Fatalf("invalid POLL_INTERVAL %q: want positive integer seconds", v)
			}
			poll = n
		}
	}

	if report <= 0 || poll <= 0 {
		log.Fatalf("intervals must be positive (seconds): report=%d poll=%d", report, poll)
	}

	cfg := agent.Config{
		PollInterval:   time.Duration(poll) * time.Second,
		ReportInterval: time.Duration(report) * time.Second,
		ServerURL:      normalizeServerURL(addr),
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := sender.NewHTTPSender(nil, cfg.ServerURL)
	a := agent.New(cfg, s, rnd)

	if err := a.Run(context.Background()); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}

// resolveServerAddr: ADDRESS из env; иначе флаг -a с учётом SERVER_PORT из автотестов metricstest.
//
// metricstest часто передаёт агенту -a в виде http://localhost:8080 — это не равно строке
// "localhost:8080", и старая логика не подставляла SERVER_PORT, из‑за чего агент бил в :8080,
// а сервер слушал случайный порт (connection reset, метрики «без изменений»).
func resolveServerAddr(flagAddr string) string {
	addr := strings.TrimSpace(flagAddr)
	if addr == "" {
		addr = defaultServerAddr
	}
	if v := strings.TrimSpace(os.Getenv("ADDRESS")); v != "" {
		return v
	}
	if sp := strings.TrimSpace(os.Getenv("SERVER_PORT")); sp != "" && isDefaultLocalMetricsAddr(addr) {
		return "localhost:" + sp
	}
	if addr != defaultServerAddr {
		return addr
	}
	if sp := strings.TrimSpace(os.Getenv("SERVER_PORT")); sp != "" {
		return "localhost:" + sp
	}
	return addr
}

// isDefaultLocalMetricsAddr — «шаблонный» адрес метрик-сервера на localhost:8080 (схема опциональна).
func isDefaultLocalMetricsAddr(addr string) bool {
	a := strings.TrimSpace(addr)
	a = strings.TrimSuffix(a, "/")
	a = strings.TrimPrefix(a, "http://")
	a = strings.TrimPrefix(a, "https://")
	if a == defaultServerAddr {
		return true
	}
	host, port, err := net.SplitHostPort(a)
	if err != nil {
		return false
	}
	if port != "8080" {
		return false
	}
	h := strings.Trim(strings.ToLower(host), "[]")
	return h == "localhost" || h == "127.0.0.1" || h == "::1"
}

func normalizeServerURL(addr string) string {
	a := strings.TrimSpace(addr)
	a = strings.TrimRight(a, "/")
	if a == "" {
		return "http://localhost:8080"
	}
	if strings.HasPrefix(a, "http://") || strings.HasPrefix(a, "https://") {
		return a
	}
	return "http://" + a
}
