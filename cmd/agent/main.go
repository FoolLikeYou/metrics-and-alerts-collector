package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/sender"
)

func main() {
	srvAddr := flag.String("a", "localhost:8080", "HTTP server address (host:port or URL)")
	reportSec := flag.Int("r", 10, "report interval, seconds")
	pollSec := flag.Int("p", 2, "poll interval, seconds")
	flag.Parse()

	addr := *srvAddr
	if v, ok := os.LookupEnv("ADDRESS"); ok {
		if t := strings.TrimSpace(v); t != "" {
			addr = t
		}
	}

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
