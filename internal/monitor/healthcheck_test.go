package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHealthChecker_PanicsOnEmptyURL(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty URL")
		}
	}()
	NewHealthChecker("", time.Second)
}

func TestNewHealthChecker_PanicsOnZeroTimeout(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero timeout")
		}
	}()
	NewHealthChecker("http://example.com", 0)
}

func TestHealthChecker_Healthy2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	hc := NewHealthChecker(ts.URL, time.Second)
	status := hc.Check(context.Background())

	if status.Err != nil {
		t.Fatalf("unexpected error: %v", status.Err)
	}
	if !status.Healthy {
		t.Errorf("expected healthy, got unhealthy (status %d)", status.StatusCode)
	}
	if status.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", status.StatusCode)
	}
	if status.URL != ts.URL {
		t.Errorf("URL mismatch: got %s", status.URL)
	}
}

func TestHealthChecker_Unhealthy5xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	hc := NewHealthChecker(ts.URL, time.Second)
	status := hc.Check(context.Background())

	if status.Err != nil {
		t.Fatalf("unexpected error: %v", status.Err)
	}
	if status.Healthy {
		t.Error("expected unhealthy for 500 response")
	}
}

func TestHealthChecker_UnreachableHost(t *testing.T) {
	hc := NewHealthChecker("http://127.0.0.1:19999", 200*time.Millisecond)
	status := hc.Check(context.Background())

	if status.Err == nil {
		t.Error("expected error for unreachable host")
	}
	if status.Healthy {
		t.Error("expected unhealthy for unreachable host")
	}
}

func TestHealthChecker_CancelledContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	hc := NewHealthChecker(ts.URL, time.Second)
	status := hc.Check(ctx)

	if status.Err == nil {
		t.Error("expected error for cancelled context")
	}
}
