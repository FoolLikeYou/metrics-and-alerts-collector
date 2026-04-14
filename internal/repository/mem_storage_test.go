package repository

import "testing"

func TestMemStorage_UpdateGauge(t *testing.T) {
	m := NewMemStorage()
	m.UpdateGauge("a", 1)
	m.UpdateGauge("a", 2)
	g, _ := m.GetAll()
	if g["a"] != 2 {
		t.Fatalf("gauge a = %v, want 2", g["a"])
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	m := NewMemStorage()
	m.UpdateCounter("c", 10)
	m.UpdateCounter("c", 5)
	_, c := m.GetAll()
	if c["c"] != 15 {
		t.Fatalf("counter c = %d, want 15", c["c"])
	}
}

func TestMemStorage_GetAll_mapsNotNil(t *testing.T) {
	m := NewMemStorage()
	g, c := m.GetAll()
	if g == nil || c == nil {
		t.Fatal("expected non-nil maps")
	}
}
