// Package monitor provides alert deduplication via AlertDedup.
//
// # AlertDedup
//
// AlertDedup prevents the same alert from being sent multiple times for
// the same port and state within a configurable time window. This is
// distinct from rate limiting (AlertFilter) and burst control (AlertThrottle):
//
//   - AlertDedup: suppresses identical (port, state) pairs within a window.
//   - AlertFilter: enforces a minimum cooldown between any alerts for a port.
//   - AlertThrottle: caps the number of alerts per port in a sliding window.
//
// # Usage
//
//	dedup := monitor.NewAlertDedup(5 * time.Minute)
//
//	if dedup.IsDuplicate(port, newState) {
//		// skip — same state already alerted recently
//		return
//	}
//	// send alert
//
// Call Reset to clear the record for a port when it is removed from monitoring.
package monitor
