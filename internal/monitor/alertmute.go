package monitor

import (
	"sync"
	"time"
)

// AlertMute allows temporarily silencing alerts for specific ports.
// While a port is muted, calls to IsMuted return true and alerts
// should be suppressed by the caller.
type AlertMute struct {
	mu    sync.Mutex
	mutes map[int]time.Time
	now   func() time.Time
}

// NewAlertMute creates a new AlertMute instance.
func NewAlertMute() *AlertMute {
	return &AlertMute{
		mutes: make(map[int]time.Time),
		now:   time.Now,
	}
}

// Mute silences alerts for the given port for the specified duration.
// Calling Mute on an already-muted port extends or replaces the mute window.
func (m *AlertMute) Mute(port int, duration time.Duration) {
	if duration <= 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mutes[port] = m.now().Add(duration)
}

// Unmute removes any active mute for the given port immediately.
func (m *AlertMute) Unmute(port int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.mutes, port)
}

// IsMuted reports whether the given port is currently muted.
func (m *AlertMute) IsMuted(port int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	expiry, ok := m.mutes[port]
	if !ok {
		return false
	}
	if m.now().After(expiry) {
		delete(m.mutes, port)
		return false
	}
	return true
}

// MutedPorts returns a snapshot of all currently muted ports and their expiry times.
func (m *AlertMute) MutedPorts() map[int]time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	out := make(map[int]time.Time)
	for port, expiry := range m.mutes {
		if now.Before(expiry) {
			out[port] = expiry
		} else {
			delete(m.mutes, port)
		}
	}
	return out
}
