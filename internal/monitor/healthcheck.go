package monitor

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HealthStatus represents the result of a health check against an HTTP endpoint.
type HealthStatus struct {
	URL        string
	StatusCode int
	Healthy    bool
	Err        error
	CheckedAt  time.Time
}

// HealthChecker performs HTTP health checks against a configured endpoint.
type HealthChecker struct {
	url     string
	client  *http.Client
	timeout time.Duration
}

// NewHealthChecker creates a HealthChecker for the given URL.
// timeout controls how long each HTTP request may take.
func NewHealthChecker(url string, timeout time.Duration) *HealthChecker {
	if timeout <= 0 {
		panic("healthcheck: timeout must be positive")
	}
	if url == "" {
		panic("healthcheck: url must not be empty")
	}
	return &HealthChecker{
		url:     url,
		timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

// Check performs a single HTTP GET against the configured URL.
// A 2xx response is considered healthy.
func (h *HealthChecker) Check(ctx context.Context) HealthStatus {
	status := HealthStatus{
		URL:       h.url,
		CheckedAt: time.Now(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.url, nil)
	if err != nil {
		status.Err = fmt.Errorf("healthcheck: build request: %w", err)
		return status
	}

	resp, err := h.client.Do(req)
	if err != nil {
		status.Err = fmt.Errorf("healthcheck: request failed: %w", err)
		return status
	}
	defer resp.Body.Close()

	status.StatusCode = resp.StatusCode
	status.Healthy = resp.StatusCode >= 200 && resp.StatusCode < 300
	return status
}
