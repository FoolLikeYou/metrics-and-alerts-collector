package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

func TestUpdateHandler_gauge(t *testing.T) {
	st := repository.NewMemStorage()
	h := UpdateHandler(st)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/TestGauge/12.5", nil)
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rr.Code)
	}
	g, _ := st.GetAll()
	if g["TestGauge"] != 12.5 {
		t.Fatalf("gauge = %v, want 12.5", g["TestGauge"])
	}
}

func TestUpdateHandler_counterAccumulates(t *testing.T) {
	st := repository.NewMemStorage()
	h := UpdateHandler(st)

	for _, v := range []string{"2", "3"} {
		req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/"+v, nil)
		rr := httptest.NewRecorder()
		h(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status %d for value %s", rr.Code, v)
		}
	}
	_, c := st.GetAll()
	if c["PollCount"] != 5 {
		t.Fatalf("PollCount = %d, want 5", c["PollCount"])
	}
}

func TestUpdateHandler_methodNotAllowed(t *testing.T) {
	st := repository.NewMemStorage()
	h := UpdateHandler(st)
	req := httptest.NewRequest(http.MethodGet, "/update/gauge/x/1", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d, want 405", rr.Code)
	}
}

func TestUpdateHandler_badPath(t *testing.T) {
	st := repository.NewMemStorage()
	h := UpdateHandler(st)
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/onlytwo", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404", rr.Code)
	}
}

func TestUpdateHandler_emptyName(t *testing.T) {
	st := repository.NewMemStorage()
	h := UpdateHandler(st)
	req := httptest.NewRequest(http.MethodPost, "/update/gauge//1", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404", rr.Code)
	}
}

func TestUpdateHandler_badGaugeValue(t *testing.T) {
	st := repository.NewMemStorage()
	h := UpdateHandler(st)
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/x/notafloat", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestUpdateHandler_badCounterValue(t *testing.T) {
	st := repository.NewMemStorage()
	h := UpdateHandler(st)
	req := httptest.NewRequest(http.MethodPost, "/update/counter/x/1.5", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestUpdateHandler_unknownType(t *testing.T) {
	st := repository.NewMemStorage()
	h := UpdateHandler(st)
	req := httptest.NewRequest(http.MethodPost, "/update/unknown/x/1", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
}

func TestListHandler(t *testing.T) {
	st := repository.NewMemStorage()
	st.UpdateGauge("G", 1)
	st.UpdateCounter("C", 2)
	h := ListHandler(st)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "gauge G = 1.000000") {
		t.Fatalf("body %q", body)
	}
	if !strings.Contains(body, "counter C = 2") {
		t.Fatalf("body %q", body)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/plain" {
		t.Fatalf("Content-Type %q", ct)
	}
}
