package monitor_test

import (
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/monitor"
)

// TestStateChangeLog_IntegrationWithHistory verifies that StateChangeLog and
// History can be used together to produce a coherent change record.
func TestStateChangeLog_IntegrationWithHistory(t *testing.T) {
	h := monitor.NewHistory(20)
	log := monitor.NewStateChangeLog(20)
	now := time.Now()

	ports := []int{80, 443, 8080}

	// First scan — all open; History should report no transitions.
	for _, p := range ports {
		if changed, _, newState := h.Record(p, true); changed {
			kind := monitor.ChangeOpened
			if !newState {
				kind = monitor.ChangeClosed
			}
			log.Record(p, kind, now)
		}
	}
	if log.Len() != 0 {
		t.Fatalf("expected 0 changes after first scan, got %d", log.Len())
	}

	// Second scan — port 8080 goes down.
	secondScan := map[int]bool{80: true, 443: true, 8080: false}
	for _, p := range ports {
		if changed, _, newState := h.Record(p, secondScan[p]); changed {
			kind := monitor.ChangeOpened
			if !newState {
				kind = monitor.ChangeClosed
			}
			log.Record(p, kind, now.Add(time.Second))
		}
	}

	if log.Len() != 1 {
		t.Fatalf("expected 1 change, got %d", log.Len())
	}
	changes := log.All()
	if changes[0].Port != 8080 || changes[0].Kind != monitor.ChangeClosed {
		t.Errorf("unexpected change: %+v", changes[0])
	}
}
