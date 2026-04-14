package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/storage"
)

func main() {
	addrFlag := flag.String("a", "localhost:8080", "HTTP server listen address")
	flag.Parse()

	addr := *addrFlag
	if v, ok := os.LookupEnv("ADDRESS"); ok {
		if t := strings.TrimSpace(v); t != "" {
			addr = t
		}
	}

	store := storage.NewMemStorage()
	mux := server.NewRouter(store)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
