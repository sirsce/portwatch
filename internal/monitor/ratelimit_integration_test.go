package monitor_test

import (
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/monitor"
)

// TestRateLimiter_IntegrationWithStateChangeLog verifies that a RateLimiter
// correctly suppresses duplicate alerts when used alongside a StateChangeLog.
func TestRateLimiter_IntegrationWithStateChangeLog(t *testing.T) {
	log := monitor.NewStateChangeLog(10)
	rl := monitor.NewRateLimiter(5 * time.Minute)

	base := time.Now()

	// Simulate port 8080 going down — first alert should be allowed.
	log.Record(monitor.StateChange{Port: 8080, Open: false, At: base})
	if !rl.Allow(8080, base) {
		t.Fatal("expected first alert to be allowed")
	}

	// Duplicate event shortly after — should be suppressed.
	log.Record(monitor.StateChange{Port: 8080, Open: false, At: base.Add(10 * time.Second)})
	if rl.Allow(8080, base.Add(10*time.Second)) {
		t.Error("expected duplicate alert within gap to be suppressed")
	}

	// After the gap expires the alert should fire again.
	log.Record(monitor.StateChange{Port: 8080, Open: false, At: base.Add(6 * time.Minute)})
	if !rl.Allow(8080, base.Add(6*time.Minute)) {
		t.Error("expected alert after gap to be allowed")
	}

	// Sanity-check: all three events were recorded in the log.
	if got := log.Len(); got != 3 {
		t.Errorf("expected 3 state changes recorded, got %d", got)
	}
}
