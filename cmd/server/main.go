package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/storage"
)

const defaultStoreFile = "metrics.store.json"

func main() {
	addrFlag := flag.String("a", "localhost:8080", "HTTP server listen address")
	storeIntervalFlag := flag.Int("i", 300, "save metrics to disk every N seconds; 0 = save synchronously on each update")
	filePathFlag := flag.String("f", defaultStoreFile, "path to JSON metrics file")
	restoreFlag := flag.Bool("r", false, "load metrics from file on startup (true/false)")
	flag.Parse()

	addr := stringFromEnvOrFlag("ADDRESS", *addrFlag)
	storeInterval := intFromEnvOrFlag("STORE_INTERVAL", *storeIntervalFlag)
	filePath := filePathFromEnvOrFlag("FILE_STORAGE_PATH", *filePathFlag)
	restore := restoreFromEnvOrFlag(*restoreFlag)

	if storeInterval < 0 {
		log.Fatal("STORE_INTERVAL / -i must be non-negative")
	}

	store := storage.NewMemStorage()
	if restore {
		if err := storage.LoadFromJSONFile(filePath, store); err != nil {
			log.Fatalf("restore from %q: %v", filePath, err)
		}
	}

	var repo repository.MetricsRepository = store
	if storeInterval == 0 {
		repo = storage.NewSyncPersistMemStorage(store, filePath)
	} else {
		go runPeriodicSave(context.Background(), time.Duration(storeInterval)*time.Second, filePath, store)
	}

	mux := server.NewRouter(repo)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func runPeriodicSave(ctx context.Context, every time.Duration, path string, m *storage.MemStorage) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := storage.SaveToJSONFile(path, m); err != nil {
				log.Printf("metrics persist: %v", err)
			}
		}
	}
}

func stringFromEnvOrFlag(envKey, flagVal string) string {
	if v, ok := os.LookupEnv(envKey); ok {
		if t := strings.TrimSpace(v); t != "" {
			return t
		}
	}
	return flagVal
}

func intFromEnvOrFlag(envKey string, flagVal int) int {
	if v, ok := os.LookupEnv(envKey); ok {
		if t := strings.TrimSpace(v); t != "" {
			n, err := strconv.Atoi(t)
			if err != nil {
				log.Fatalf("%s: %v", envKey, err)
			}
			return n
		}
	}
	return flagVal
}

func filePathFromEnvOrFlag(envKey, flagVal string) string {
	if v, ok := os.LookupEnv(envKey); ok {
		if t := strings.TrimSpace(v); t != "" {
			return t
		}
	}
	return flagVal
}

func restoreFromEnvOrFlag(flagVal bool) bool {
	v, ok := os.LookupEnv("RESTORE")
	if !ok {
		return flagVal
	}
	t := strings.TrimSpace(v)
	if t == "" {
		return flagVal
	}
	b, err := strconv.ParseBool(t)
	if err != nil {
		log.Fatalf("RESTORE: %v", err)
	}
	return b
}
