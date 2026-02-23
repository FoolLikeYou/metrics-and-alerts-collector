package main

import (
	"metrics-and-alerts-collector/internal/handler"
	"metrics-and-alerts-collector/internal/repository"
	"net/http"
)

func main() {
	st := repository.NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handler.UpdateHandler(st))
	mux.HandleFunc("/metrics", handler.ListHandler(st))

	_ = http.ListenAndServe(":8080", mux)
}
