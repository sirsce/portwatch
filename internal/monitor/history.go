package monitor

import (
	"sync"
	"time"
)

// PortStatus represents the observed state of a single port at a point in time.
type PortStatus struct {
	Port      int
	Open      bool
	CheckedAt time.Time
}

// StateChange records a transition in port state.
type StateChange struct {
	Port      int
	WasOpen   bool
	NowOpen   bool
	ChangedAt time.Time
}

// History tracks the last known status and state changes for monitored ports.
type History struct {
	mu      sync.RWMutex
	latest  map[int]PortStatus
	changes []StateChange
	maxChanges int
}

// NewHistory creates a History that retains at most maxChanges state transitions.
func NewHistory(maxChanges int) *History {
	if maxChanges <= 0 {
		maxChanges = 100
	}
	return &History{
		latest:     make(map[int]PortStatus),
		maxChanges: maxChanges,
	}
}

// Record updates the latest status for a port and appends a StateChange if the
// open/closed state differs from the previously recorded status.
// Returns true if a state change was detected.
func (h *History) Record(port int, open bool) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	prev, exists := h.latest[port]

	h.latest[port] = PortStatus{Port: port, Open: open, CheckedAt: now}

	if exists && prev.Open == open {
		return false
	}
	if !exists {
		return false
	}

	change := StateChange{
		Port:      port,
		WasOpen:   prev.Open,
		NowOpen:   open,
		ChangedAt: now,
	}
	h.changes = append(h.changes, change)
	if len(h.changes) > h.maxChanges {
		h.changes = h.changes[len(h.changes)-h.maxChanges:]
	}
	return true
}

// Latest returns the most recently recorded status for the given port.
// The second return value is false if the port has never been recorded.
func (h *History) Latest(port int) (PortStatus, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	s, ok := h.latest[port]
	return s, ok
}

// Changes returns a copy of the recorded state-change log.
func (h *History) Changes() []StateChange {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]StateChange, len(h.changes))
	copy(out, h.changes)
	return out
}
