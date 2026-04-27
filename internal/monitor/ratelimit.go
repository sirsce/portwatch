package monitor

import (
	"sync"
	"time"
)

// RateLimiter prevents alert flooding by enforcing a minimum interval
// between alerts for the same port.
type RateLimiter struct {
	mu       sync.Mutex
	lastSent map[int]time.Time
	minGap   time.Duration
}

// NewRateLimiter creates a RateLimiter that allows at most one alert per
// port within the given minGap duration.
func NewRateLimiter(minGap time.Duration) *RateLimiter {
	if minGap <= 0 {
		panic("monitor: RateLimiter minGap must be positive")
	}
	return &RateLimiter{
		lastSent: make(map[int]time.Time),
		minGap:   minGap,
	}
}

// Allow returns true if an alert for the given port is permitted at now.
// If allowed, it records now as the last-sent time for that port.
func (r *RateLimiter) Allow(port int, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	last, seen := r.lastSent[port]
	if seen && now.Sub(last) < r.minGap {
		return false
	}
	r.lastSent[port] = now
	return true
}

// Reset clears the rate-limit record for a specific port.
func (r *RateLimiter) Reset(port int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.lastSent, port)
}

// ResetAll clears all rate-limit records.
func (r *RateLimiter) ResetAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastSent = make(map[int]time.Time)
}
