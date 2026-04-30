package monitor

import (
	"testing"
	"time"
)

func escalationLevels() []EscalationLevel {
	return []EscalationLevel{
		{Name: "warn", Threshold: 2, Cooldown: 10 * time.Minute},
		{Name: "critical", Threshold: 5, Cooldown: 5 * time.Minute},
	}
}

func TestNewAlertEscalation_PanicsOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on empty levels")
		}
	}()
	NewAlertEscalation(nil)
}

func TestAlertEscalation_BelowThreshold(t *testing.T) {
	ae := NewAlertEscalation(escalationLevels())
	// First call: count becomes 1, below threshold of 2
	lvl, err := ae.Record(8080)
	if err == nil {
		t.Fatalf("expected error below threshold, got level %s", lvl.Name)
	}
}

func TestAlertEscalation_ReachesWarnLevel(t *testing.T) {
	ae := NewAlertEscalation(escalationLevels())
	ae.Record(8080) // count=1, below warn
	lvl, err := ae.Record(8080) // count=2, reaches warn
	if err != nil {
		t.Fatalf("expected warn level, got error: %v", err)
	}
	if lvl.Name != "warn" {
		t.Fatalf("expected warn, got %s", lvl.Name)
	}
}

func TestAlertEscalation_EscalatesToCritical(t *testing.T) {
	now := time.Now()
	ae := NewAlertEscalation(escalationLevels())
	ae.timeNow = func() time.Time { return now }

	for i := 0; i < 4; i++ {
		ae.Record(9090)
	}
	// Advance past warn cooldown so critical can fire
	now = now.Add(11 * time.Minute)
	lvl, err := ae.Record(9090) // count=5, critical
	if err != nil {
		t.Fatalf("expected critical, got error: %v", err)
	}
	if lvl.Name != "critical" {
		t.Fatalf("expected critical, got %s", lvl.Name)
	}
}

func TestAlertEscalation_CooldownSuppresses(t *testing.T) {
	now := time.Now()
	ae := NewAlertEscalation(escalationLevels())
	ae.timeNow = func() time.Time { return now }

	ae.Record(1234) // count=1
	ae.Record(1234) // count=2, warn fires, lastAt set

	// Immediately try again — still in cooldown
	_, err := ae.Record(1234)
	if err == nil {
		t.Fatal("expected cooldown suppression")
	}
}

func TestAlertEscalation_Reset(t *testing.T) {
	ae := NewAlertEscalation(escalationLevels())
	ae.Record(7070)
	ae.Record(7070)
	ae.Reset(7070)
	if ae.Count(7070) != 0 {
		t.Fatal("expected count 0 after reset")
	}
	// After reset, count starts fresh — below threshold again
	_, err := ae.Record(7070)
	if err == nil {
		t.Fatal("expected below-threshold error after reset")
	}
}

func TestAlertEscalation_DifferentPortsAreIndependent(t *testing.T) {
	ae := NewAlertEscalation(escalationLevels())
	ae.Record(1111)
	ae.Record(2222)
	if ae.Count(1111) != 1 || ae.Count(2222) != 1 {
		t.Fatal("ports should track counts independently")
	}
}
