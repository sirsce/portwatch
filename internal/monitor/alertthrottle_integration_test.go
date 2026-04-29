package monitor

import (
	"testing"
	"time"
)

// TestAlertThrottle_IntegrationWithAlertFilter verifies that AlertThrottle and
// AlertFilter can be composed together: the filter suppresses repeated state
// changes while the throttle independently caps burst volume per port.
func TestAlertThrottle_IntegrationWithAlertFilter(t *testing.T) {
	const port = 8080

	filter := NewAlertFilter(200 * time.Millisecond)
	throttle := NewAlertThrottle(time.Minute, 2)

	allowedByBoth := func(p int, state string) bool {
		return filter.Allow(p, state) && throttle.Allow(p)
	}

	// First transition: open -> closed should pass both.
	if !allowedByBoth(port, "closed") {
		t.Fatal("first transition should be allowed by both")
	}

	// Immediate repeat: filter should block (cooldown), throttle would allow.
	if allowedByBoth(port, "closed") {
		t.Fatal("repeat within cooldown should be blocked by filter")
	}

	// Wait for filter cooldown to expire.
	time.Sleep(250 * time.Millisecond)

	// Second transition after cooldown: filter allows, throttle allows (count=2).
	if !allowedByBoth(port, "closed") {
		t.Fatal("second transition after cooldown should be allowed")
	}

	// Third transition: filter cooldown not expired yet, but even if it were,
	// throttle burst limit (2) is now exhausted.
	time.Sleep(250 * time.Millisecond)
	if allowedByBoth(port, "closed") {
		t.Fatal("third transition should be blocked by throttle burst limit")
	}

	// Reset throttle and confirm filter+throttle allows again after cooldown.
	throttle.Reset(port)
	time.Sleep(250 * time.Millisecond)
	if !allowedByBoth(port, "closed") {
		t.Fatal("should be allowed after throttle reset and filter cooldown")
	}
}
