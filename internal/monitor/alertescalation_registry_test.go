package monitor

import (
	"sync"
	"testing"
	"time"
)

// TestAlertEscalation_ConcurrentAccess verifies that concurrent Record and
// Reset calls on the same port do not cause data races.
func TestAlertEscalation_ConcurrentAccess(t *testing.T) {
	levels := []EscalationLevel{
		{Name: "warn", Threshold: 3, Cooldown: 0},
	}
	ae := NewAlertEscalation(levels)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			port := 8000 + (n % 5)
			ae.Record(port)
		}(i)
	}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ae.Reset(8000 + n%5)
		}(i)
	}
	wg.Wait()
}

// TestAlertEscalation_CountAccuracy verifies Count matches the number of
// Record calls when no Reset has occurred.
func TestAlertEscalation_CountAccuracy(t *testing.T) {
	levels := []EscalationLevel{
		{Name: "warn", Threshold: 100, Cooldown: time.Hour},
	}
	ae := NewAlertEscalation(levels)
	port := 5555
	for i := 0; i < 7; i++ {
		ae.Record(port)
	}
	if got := ae.Count(port); got != 7 {
		t.Fatalf("expected count 7, got %d", got)
	}
}

// TestAlertEscalation_MultiplePortsIsolated ensures escalation state for one
// port does not bleed into another.
func TestAlertEscalation_MultiplePortsIsolated(t *testing.T) {
	levels := []EscalationLevel{
		{Name: "warn", Threshold: 2, Cooldown: 0},
	}
	ae := NewAlertEscalation(levels)

	// Drive port A to warn level
	ae.Record(1010)
	ae.Record(1010)

	// Port B should still be below threshold
	_, err := ae.Record(2020)
	if err == nil {
		t.Fatal("port 2020 should be below threshold")
	}

	// Port A should be at warn
	// Reset and verify counts are separate
	ae.Reset(1010)
	if ae.Count(1010) != 0 {
		t.Fatal("port 1010 count should be 0 after reset")
	}
	if ae.Count(2020) != 1 {
		t.Fatalf("port 2020 count should be 1, got %d", ae.Count(2020))
	}
}
