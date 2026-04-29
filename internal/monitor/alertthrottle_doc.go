// Package monitor provides monitoring primitives for portwatch.
//
// # AlertThrottle
//
// AlertThrottle limits the number of alerts fired for a given port within a
// rolling time window. This is complementary to AlertFilter (which suppresses
// repeated state-change alerts during a cooldown period): AlertThrottle caps
// the absolute burst volume of alerts regardless of state transitions.
//
// Usage:
//
//	throttle := monitor.NewAlertThrottle(5*time.Minute, 3)
//
//	if throttle.Allow(port) {
//	    // send alert
//	}
//
// The window and maxBurst parameters must both be positive; NewAlertThrottle
// panics otherwise to surface misconfiguration early.
//
// AlertThrottle is safe for concurrent use by multiple goroutines.
package monitor
