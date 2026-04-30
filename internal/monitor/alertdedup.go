package monitor

import (
	"sync"
	"time"
)

// AlertDedup suppresses duplicate alerts for the same port and state
// within a configurable deduplication window.
type AlertDedup struct {
	mu      sync.Mutex
	window  time.Duration
	records map[string]dedupEntry
	now     func() time.Time
}

type dedupEntry struct {
	state     string
	recordedAt time.Time
}

// NewAlertDedup creates an AlertDedup with the given deduplication window.
// Panics if window is zero or negative.
func NewAlertDedup(window time.Duration) *AlertDedup {
	if window <= 0 {
		panic("alertdedup: window must be positive")
	}
	return &AlertDedup{
		window:  window,
		records: make(map[string]dedupEntry),
		now:     time.Now,
	}
}

// IsDuplicate returns true if the same (port, state) pair was already
// seen within the deduplication window. If not a duplicate, it records
// the event and returns false.
func (d *AlertDedup) IsDuplicate(port int, state string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := itoa(port)
	now := d.now()

	if entry, ok := d.records[key]; ok {
		if entry.state == state && now.Sub(entry.recordedAt) < d.window {
			return true
		}
	}

	d.records[key] = dedupEntry{state: state, recordedAt: now}
	return false
}

// Reset clears the deduplication record for a specific port.
func (d *AlertDedup) Reset(port int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.records, itoa(port))
}

// Len returns the number of tracked ports.
func (d *AlertDedup) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.records)
}
