package monitor

import (
	"fmt"
	"strings"
	"time"
)

// AlertSummary aggregates state change events over a rolling window and
// produces a human-readable digest suitable for inclusion in alert messages.
type AlertSummary struct {
	window time.Duration
	log    *StateChangeLog
}

// NewAlertSummary creates an AlertSummary that looks back at most window
// duration when building a digest. It panics if window is zero or negative.
func NewAlertSummary(window time.Duration, log *StateChangeLog) *AlertSummary {
	if window <= 0 {
		panic("alertsummary: window must be positive")
	}
	if log == nil {
		panic("alertsummary: log must not be nil")
	}
	return &AlertSummary{window: window, log: log}
}

// DigestEntry represents a single summarised event within the window.
type DigestEntry struct {
	Port      int
	OldState  string
	NewState  string
	ChangedAt time.Time
}

// Digest returns all state changes that occurred within the configured window,
// ordered from oldest to newest.
func (s *AlertSummary) Digest(now time.Time) []DigestEntry {
	cutoff := now.Add(-s.window)
	all := s.log.All()
	var entries []DigestEntry
	for _, sc := range all {
		if sc.At.After(cutoff) {
			entries = append(entries, DigestEntry{
				Port:      sc.Port,
				OldState:  sc.OldState,
				NewState:  sc.NewState,
				ChangedAt: sc.At,
			})
		}
	}
	return entries
}

// Format renders the digest as a multi-line human-readable string.
// Returns a placeholder message when there are no entries.
func (s *AlertSummary) Format(now time.Time) string {
	entries := s.Digest(now)
	if len(entries) == 0 {
		return fmt.Sprintf("No state changes in the last %s.", s.window)
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("State changes in the last %s (%d total):\n", s.window, len(entries)))
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("  port %-6d  %s -> %s  at %s\n",
			e.Port, e.OldState, e.NewState,
			e.ChangedAt.UTC().Format(time.RFC3339)))
	}
	return strings.TrimRight(sb.String(), "\n")
}
