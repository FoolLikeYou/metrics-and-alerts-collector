package storage_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/storage"
)

func TestMemStorage_SetGauge_Replaces(t *testing.T) {
	m := storage.NewMemStorage()
	m.SetGauge("a", 1)
	m.SetGauge("a", 2)
	v, ok := m.GetGauge("a")
	if !ok || v != 2 {
		t.Fatalf("gauge: got %v ok=%v", v, ok)
	}
}

func TestMemStorage_AddCounter_Accumulates(t *testing.T) {
	m := storage.NewMemStorage()
	m.AddCounter("c", 3)
	m.AddCounter("c", 7)
	v, ok := m.GetCounter("c")
	if !ok || v != 10 {
		t.Fatalf("counter: got %d ok=%v", v, ok)
	}
}

func TestMemStorage_ListMetrics(t *testing.T) {
	m := storage.NewMemStorage()
	m.SetGauge("b", 2)
	m.SetGauge("a", 1)
	m.AddCounter("c", 3)
	list := m.ListMetrics()
	if len(list) != 3 {
		t.Fatalf("len: got %d want 3", len(list))
	}
	if list[0].Name != "a" || list[0].MType != "gauge" || list[0].Value != "1" {
		t.Fatalf("first row: %+v", list[0])
	}
}

func TestMemStorage_ConcurrentWrites(t *testing.T) {
	m := storage.NewMemStorage()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.AddCounter("x", 1)
			m.SetGauge("g", 1)
		}()
	}
	wg.Wait()
	c, _ := m.GetCounter("x")
	if c != 50 {
		t.Fatalf("counter: got %d want 50", c)
	}
}

func TestJSONFileSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "m.json")

	m := storage.NewMemStorage()
	m.SetGauge("LastGC", 1.25e18)
	m.AddCounter("NumGC", 42)

	if err := storage.SaveToJSONFile(path, m); err != nil {
		t.Fatal(err)
	}

	m2 := storage.NewMemStorage()
	if err := storage.LoadFromJSONFile(path, m2); err != nil {
		t.Fatal(err)
	}
	g, ok := m2.GetGauge("LastGC")
	if !ok || g != 1.25e18 {
		t.Fatalf("gauge: %v ok=%v", g, ok)
	}
	c, ok := m2.GetCounter("NumGC")
	if !ok || c != 42 {
		t.Fatalf("counter: %d ok=%v", c, ok)
	}
}

func TestSyncPersistMemStorage_WritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "live.json")
	inner := storage.NewMemStorage()
	w := storage.NewSyncPersistMemStorage(inner, path)
	w.SetGauge("x", 1)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty file")
	}
}
