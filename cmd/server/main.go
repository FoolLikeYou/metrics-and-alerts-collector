package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/storage"
)

func main() {
	addr := flag.String("a", "localhost:8080", "HTTP server listen address")
	flag.Parse()

	store := storage.NewMemStorage()
	mux := server.NewRouter(store)

	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}
