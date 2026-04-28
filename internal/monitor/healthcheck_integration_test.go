package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// TestHealthChecker_IntegrationWithTicker verifies that RunEvery drives
// repeated health checks and that results accumulate correctly.
func TestHealthChecker_IntegrationWithTicker(t *testing.T) {
	var hitCount atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	hc := NewHealthChecker(ts.URL, time.Second)

	var results []HealthStatus
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	ticker := NewTicker(100 * time.Millisecond)
	RunEvery(ctx, ticker, func() {
		status := hc.Check(ctx)
		results = append(results, status)
	})

	if len(results) < 2 {
		t.Fatalf("expected at least 2 health checks, got %d", len(results))
	}
	for i, s := range results {
		if !s.Healthy {
			t.Errorf("result[%d]: expected healthy", i)
		}
		if s.Err != nil {
			t.Errorf("result[%d]: unexpected error: %v", i, s.Err)
		}
	}
}
