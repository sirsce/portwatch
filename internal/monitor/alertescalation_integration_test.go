package monitor

import (
	"testing"
	"time"
)

// TestAlertEscalation_IntegrationWithStateChangeLog verifies that escalation
// levels align with the number of state changes recorded in the log.
func TestAlertEscalation_IntegrationWithStateChangeLog(t *testing.T) {
	log := NewStateChangeLog(20)
	levels := []EscalationLevel{
		{Name: "warn", Threshold: 2, Cooldown: 0},
		{Name: "critical", Threshold: 4, Cooldown: 0},
	}
	ae := NewAlertEscalation(levels)

	port := 8080
	states := []bool{true, false, true, false}

	var lastLevel string
	for _, open := range states {
		log.Record(port, open)
		if lvl, err := ae.Record(port); err == nil {
			lastLevel = lvl.Name
		}
	}

	if lastLevel != "critical" {
		t.Fatalf("expected critical escalation after %d changes, got %q", len(states), lastLevel)
	}

	if log.Len() != len(states) {
		t.Fatalf("expected %d log entries, got %d", len(states), log.Len())
	}
}

// TestAlertEscalation_IntegrationWithAlertFilter verifies that escalation
// and alert filtering work together: escalation fires while filter may suppress.
func TestAlertEscalation_IntegrationWithAlertFilter(t *testing.T) {
	levels := []EscalationLevel{
		{Name: "warn", Threshold: 1, Cooldown: 0},
	}
	ae := NewAlertEscalation(levels)
	filter := NewAlertFilter(30*time.Second, func() time.Time { return time.Now() })

	port := 9999

	// First alert: escalation fires, filter allows
	lvl, escErr := ae.Record(port)
	allowed := filter.Allow(port)

	if escErr != nil {
		t.Fatalf("expected escalation to fire: %v", escErr)
	}
	if lvl.Name != "warn" {
		t.Fatalf("expected warn level, got %s", lvl.Name)
	}
	if !allowed {
		t.Fatal("expected filter to allow first alert")
	}

	// Second alert: escalation fires (cooldown=0), filter suppresses
	_, escErr2 := ae.Record(port)
	allowed2 := filter.Allow(port)

	if escErr2 != nil {
		t.Fatalf("expected escalation to fire again: %v", escErr2)
	}
	if allowed2 {
		t.Fatal("expected filter to suppress second alert within cooldown")
	}
}
