package monitor

import (
	"fmt"
	"sync"
	"time"
)

// AlertThrottle tracks per-port alert counts within a rolling window and
// suppresses alerts once a maximum burst limit has been reached.
type AlertThrottle struct {
	mu       sync.Mutex
	window   time.Duration
	maxBurst int
	events   map[int][]time.Time
	nowFn    func() time.Time
}

// NewAlertThrottle creates an AlertThrottle with the given rolling window and
// maximum burst count. Panics if window <= 0 or maxBurst <= 0.
func NewAlertThrottle(window time.Duration, maxBurst int) *AlertThrottle {
	if window <= 0 {
		panic("alertthrottle: window must be positive")
	}
	if maxBurst <= 0 {
		panic("alertthrottle: maxBurst must be positive")
	}
	return &AlertThrottle{
		window:   window,
		maxBurst: maxBurst,
		events:   make(map[int][]time.Time),
		nowFn:    time.Now,
	}
}

// Allow reports whether an alert for the given port should be allowed.
// It records the event timestamp and prunes stale entries outside the window.
func (t *AlertThrottle) Allow(port int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.nowFn()
	cutoff := now.Add(-t.window)

	// Prune old events outside the rolling window.
	filtered := t.events[port][:0]
	for _, ts := range t.events[port] {
		if ts.After(cutoff) {
			filtered = append(filtered, ts)
		}
	}
	t.events[port] = filtered

	if len(t.events[port]) >= t.maxBurst {
		return false
	}

	t.events[port] = append(t.events[port], now)
	return true
}

// Count returns the number of recorded events within the window for the given port.
func (t *AlertThrottle) Count(port int) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.nowFn()
	cutoff := now.Add(-t.window)
	count := 0
	for _, ts := range t.events[port] {
		if ts.After(cutoff) {
			count++
		}
	}
	return count
}

// Reset clears all recorded events for the given port.
func (t *AlertThrottle) Reset(port int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.events, port)
}

// String returns a human-readable description of the throttle configuration.
func (t *AlertThrottle) String() string {
	return fmt.Sprintf("AlertThrottle(window=%s, maxBurst=%d)", t.window, t.maxBurst)
}
