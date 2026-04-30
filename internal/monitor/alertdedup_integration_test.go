package monitor

import (
	"testing"
	"time"
)

// TestAlertDedup_IntegrationWithStateChangeLog verifies that AlertDedup
// correctly suppresses repeated alerts when integrated with StateChangeLog.
func TestAlertDedup_IntegrationWithStateChangeLog(t *testing.T) {
	log := NewStateChangeLog(10)
	dedup := NewAlertDedup(time.Minute)

	// Simulate state changes recorded in the log
	log.Record(StateChange{Port: 8080, OldState: "open", NewState: "closed", At: time.Now()})
	log.Record(StateChange{Port: 8080, OldState: "closed", NewState: "open", At: time.Now()})

	changes := log.All()
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}

	// First alert for port 8080 closed → not duplicate
	if dedup.IsDuplicate(8080, "closed") {
		t.Fatal("first closed alert should not be duplicate")
	}

	// Repeated closed alert within window → duplicate
	if !dedup.IsDuplicate(8080, "closed") {
		t.Fatal("repeated closed alert within window should be duplicate")
	}

	// State changed to open → not duplicate
	if dedup.IsDuplicate(8080, "open") {
		t.Fatal("open alert after state change should not be duplicate")
	}

	// Repeated open alert within window → duplicate
	if !dedup.IsDuplicate(8080, "open") {
		t.Fatal("repeated open alert within window should be duplicate")
	}
}

// TestAlertDedup_IntegrationWithAlertFilter verifies that AlertDedup and
// AlertFilter work independently on the same port without interference.
func TestAlertDedup_IntegrationWithAlertFilter(t *testing.T) {
	dedup := NewAlertDedup(time.Minute)
	filter := NewAlertFilter(time.Minute)

	// Both should pass on first call
	if dedup.IsDuplicate(8080, "closed") {
		t.Fatal("dedup: first call should not be duplicate")
	}
	if !filter.Allow(8080) {
		t.Fatal("filter: first call should be allowed")
	}

	// Dedup suppresses; filter still blocks within cooldown
	if !dedup.IsDuplicate(8080, "closed") {
		t.Fatal("dedup: second call should be duplicate")
	}
	if filter.Allow(8080) {
		t.Fatal("filter: second call within cooldown should be denied")
	}
}
