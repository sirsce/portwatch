package monitor

import (
	"testing"
	"time"
)

// TestAlertSchedule_IntegrationWithAlertMute verifies that AlertSchedule and
// AlertMute can be composed so that an alert is suppressed either because it
// is outside the active window OR because the port has been manually muted.
func TestAlertSchedule_IntegrationWithAlertMute(t *testing.T) {
	schedule := NewAlertSchedule()
	mute := NewAlertMute()

	// Set schedule to a window that does NOT include hour 3.
	_ = schedule.AddWindow(9, 17)
	schedule.nowFn = func() time.Time {
		return time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC)
	}

	shouldAlert := func(port int) bool {
		if !schedule.IsActive() {
			return false
		}
		if mute.IsMuted(port) {
			return false
		}
		return true
	}

	// Outside window — alert suppressed.
	if shouldAlert(8080) {
		t.Error("expected alert suppressed outside schedule window")
	}

	// Move into active window.
	schedule.nowFn = func() time.Time {
		return time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	}

	// Inside window, not muted — alert allowed.
	if !shouldAlert(8080) {
		t.Error("expected alert allowed inside window and not muted")
	}

	// Mute port 8080 for 1 hour.
	mute.Mute(8080, time.Hour)

	// Inside window but muted — alert suppressed.
	if shouldAlert(8080) {
		t.Error("expected alert suppressed for muted port")
	}

	// Different port is unaffected.
	if !shouldAlert(9090) {
		t.Error("expected alert allowed for unmuted port 9090")
	}
}
