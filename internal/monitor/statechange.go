package monitor

import "time"

// ChangeKind describes the direction of a port state transition.
type ChangeKind string

const (
	// ChangeOpened means the port transitioned from closed to open.
	ChangeOpened ChangeKind = "opened"
	// ChangeClosed means the port transitioned from open to closed.
	ChangeClosed ChangeKind = "closed"
)

// StateChange records a single port state transition.
type StateChange struct {
	Port      int        `json:"port"`
	Kind      ChangeKind `json:"kind"`
	OccurredAt time.Time `json:"occurred_at"`
}

// StateChangeLog holds an ordered list of state changes for a monitoring
// session, bounded by a configurable maximum size.
type StateChangeLog struct {
	max     int
	changes []StateChange
}

// NewStateChangeLog creates a StateChangeLog that retains at most max entries.
// It panics if max is less than 1.
func NewStateChangeLog(max int) *StateChangeLog {
	if max < 1 {
		panic("statechange: max must be >= 1")
	}
	return &StateChangeLog{max: max, changes: make([]StateChange, 0, max)}
}

// Record appends a new StateChange. When the log is full the oldest entry is
// dropped to make room.
func (l *StateChangeLog) Record(port int, kind ChangeKind, at time.Time) {
	if len(l.changes) >= l.max {
		l.changes = l.changes[1:]
	}
	l.changes = append(l.changes, StateChange{Port: port, Kind: kind, OccurredAt: at})
}

// All returns a copy of the recorded changes in chronological order.
func (l *StateChangeLog) All() []StateChange {
	out := make([]StateChange, len(l.changes))
	copy(out, l.changes)
	return out
}

// Len returns the number of recorded changes.
func (l *StateChangeLog) Len() int { return len(l.changes) }
