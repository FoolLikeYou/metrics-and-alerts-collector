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

const (
	defaultHost       = "localhost"
	defaultPort       = "8080"
	defaultServerAddr = defaultHost + ":" + defaultPort
)

func main() {
	srvAddr := flag.String("a", defaultServerAddr, "HTTP server address (host:port or URL)")
	// metricstest может передавать порт так же, как в свой бинарник: -server-port=N
	// Без регистрации флага агент падает на flag.Parse с «flag provided but not defined».
	serverPortFlag := flag.String("server-port", "", "metrics server TCP port (optional, used by autotests)")
	reportSec := flag.Int("r", 10, "report interval, seconds")
	pollSec := flag.Int("p", 2, "poll interval, seconds")
	flag.Parse()

	addr := resolveServerAddr(strings.TrimSpace(*srvAddr), strings.TrimSpace(*serverPortFlag), flag.Args())

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

// resolveServerAddr задаёт host:port (или URL) сервера метрик для агента.
//
// Порядок сознательно совпадает с cmd/server (ADDRESS из env важнее флага -a),
// чтобы агент ходил туда же, куда слушает сервер, если раннер выставляет ADDRESS.
// Дальше — варианты из metricstest (iter4+ передаёт -a=..., iter7 часто только env).
func resolveServerAddr(flagAddr, serverPortFromFlag string, positional []string) string {
	if v := strings.TrimSpace(os.Getenv("ADDRESS")); v != "" {
		return v
	}
	if sp := strings.TrimSpace(os.Getenv("SERVER_PORT")); sp != "" {
		return defaultHost + ":" + sp
	}
	if serverPortFromFlag != "" && isAllDecimalDigits(serverPortFromFlag) {
		return defaultHost + ":" + serverPortFromFlag
	}
	if fromPos := addrFromPositionalArgs(positional); fromPos != "" {
		return fromPos
	}
	addr := strings.TrimSpace(flagAddr)
	if addr == "" {
		return defaultServerAddr
	}
	return addr
}

func addrFromPositionalArgs(positional []string) string {
	if len(positional) != 1 {
		return ""
	}
	p := strings.TrimSpace(positional[0])
	if p == "" {
		return ""
	}
	if isAllDecimalDigits(p) {
		return defaultHost + ":" + p
	}
	return p
}

func isAllDecimalDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func normalizeServerURL(addr string) string {
	a := strings.TrimSpace(addr)
	a = strings.TrimRight(a, "/")
	if a == "" {
		return "http://" + defaultServerAddr
	}
	if strings.HasPrefix(a, "http://") || strings.HasPrefix(a, "https://") {
		return a
	}
	return "http://" + a
}
