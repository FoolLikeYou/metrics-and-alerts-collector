package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
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

	if *reportSec <= 0 || *pollSec <= 0 {
		log.Fatalf("intervals must be positive (seconds): -r=%d -p=%d", *reportSec, *pollSec)
	}

	cfg := agent.Config{
		PollInterval:   time.Duration(*pollSec) * time.Second,
		ReportInterval: time.Duration(*reportSec) * time.Second,
		ServerURL:      normalizeServerURL(*srvAddr),
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
