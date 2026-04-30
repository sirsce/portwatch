package monitor

import (
	"fmt"
	"sync"
	"time"
)

// EscalationLevel represents a named escalation tier.
type EscalationLevel struct {
	Name      string
	Threshold int           // number of consecutive alerts before escalating
	Cooldown  time.Duration // minimum time between escalated alerts
}

// AlertEscalation tracks per-port alert counts and escalates
// to higher-priority notifiers after a threshold is breached.
type AlertEscalation struct {
	mu      sync.Mutex
	levels  []EscalationLevel
	counts  map[int]int
	lastAt  map[int]time.Time
	timeNow func() time.Time
}

// NewAlertEscalation creates an AlertEscalation with the given levels.
// Levels must be provided in ascending threshold order.
// Panics if levels is empty.
func NewAlertEscalation(levels []EscalationLevel) *AlertEscalation {
	if len(levels) == 0 {
		panic("alertescalation: at least one escalation level is required")
	}
	return &AlertEscalation{
		levels: levels,
		counts: make(map[int]int),
		lastAt: make(map[int]time.Time),
		timeNow: time.Now,
	}
}

// Record increments the alert count for the given port and returns the
// active EscalationLevel if an escalated alert should fire, or an error
// if the cooldown has not elapsed or no escalation is warranted.
func (a *AlertEscalation) Record(port int) (*EscalationLevel, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.counts[port]++
	count := a.counts[port]
	now := a.timeNow()

	var active *EscalationLevel
	for i := len(a.levels) - 1; i >= 0; i-- {
		if count >= a.levels[i].Threshold {
			active = &a.levels[i]
			break
		}
	}
	if active == nil {
		return nil, fmt.Errorf("alertescalation: port %d count %d below all thresholds", port, count)
	}

	last, seen := a.lastAt[port]
	if seen && now.Sub(last) < active.Cooldown {
		return nil, fmt.Errorf("alertescalation: port %d in cooldown for level %s", port, active.Name)
	}

	a.lastAt[port] = now
	return active, nil
}

// Reset clears the alert count and last-alert time for the given port.
func (a *AlertEscalation) Reset(port int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.counts, port)
	delete(a.lastAt, port)
}

// Count returns the current alert count for the given port.
func (a *AlertEscalation) Count(port int) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.counts[port]
}
