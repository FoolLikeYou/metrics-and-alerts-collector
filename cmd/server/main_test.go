package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/handler"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

func newTestMux() (http.Handler, *repository.MemStorage) {
	st := repository.NewMemStorage()
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handler.UpdateHandler(st))
	mux.HandleFunc("/metrics", handler.ListHandler(st))
	return mux, st
}

func TestServer_updateAndList(t *testing.T) {
	mux, st := newTestMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/update/counter/n/7", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update status %d", resp.StatusCode)
	}

	resp, err = http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "counter n = 7") {
		t.Fatalf("metrics body: %s", body)
	}
	_, c := st.GetAll()
	if c["n"] != 7 {
		t.Fatalf("storage counter n = %d", c["n"])
	}
}
