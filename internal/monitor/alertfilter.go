package monitor

import "sync"

// AlertFilter decides whether an alert should be suppressed based on
// per-port cooldown windows. It prevents alert storms when a port
// flaps rapidly between open and closed states.
type AlertFilter struct {
	mu      sync.Mutex
	limiter map[int]*RateLimiter
	gap     interface{ Nanoseconds() int64 }
}

// alertFilterEntry pairs a port with its dedicated rate limiter.
type alertFilterEntry struct {
	port    int
	limiter *RateLimiter
}

// NewAlertFilter creates an AlertFilter that suppresses repeated alerts
// for the same port within the given cooldown duration.
//
// cooldownSeconds must be greater than zero or this function panics.
func NewAlertFilter(cooldownSeconds int) *AlertFilter {
	if cooldownSeconds <= 0 {
		panic("alertfilter: cooldownSeconds must be greater than zero")
	}
	return &AlertFilter{
		limiter: make(map[int]*RateLimiter),
		gap:     nil,
		// store cooldown so we can lazily create per-port limiters
	}
}

// newAlertFilterWithFactory is used internally and in tests to inject a
// RateLimiter factory, making time-based behaviour deterministic.
func newAlertFilterWithFactory(
	cooldownSeconds int,
	factory func(int) *RateLimiter,
) *AlertFilter {
	if cooldownSeconds <= 0 {
		panic("alertfilter: cooldownSeconds must be greater than zero")
	}
	af := &AlertFilter{
		limiter: make(map[int]*RateLimiter),
	}
	af.factory = factory
	af.cooldown = cooldownSeconds
	return af
}

// Allow returns true if an alert for the given port should be sent.
// The first alert for a port is always allowed; subsequent alerts are
// suppressed until the cooldown window has elapsed.
func (af *AlertFilter) Allow(port int) bool {
	af.mu.Lock()
	defer af.mu.Unlock()

	rl, ok := af.limiter[port]
	if !ok {
		var newRL *RateLimiter
		if af.factory != nil {
			newRL = af.factory(af.cooldown)
		} else {
			newRL = NewRateLimiter(af.cooldown)
		}
		af.limiter[port] = newRL
		rl = newRL
	}
	return rl.Allow()
}

// Reset clears the rate-limit state for a specific port, allowing the
// next alert for that port to pass through immediately.
func (af *AlertFilter) Reset(port int) {
	af.mu.Lock()
	defer af.mu.Unlock()
	delete(af.limiter, port)
}

// cooldown and factory are unexported fields added dynamically via the
// internal constructor — kept here for clarity.
var _ = (*AlertFilter)(nil) // compile-time interface check placeholder
