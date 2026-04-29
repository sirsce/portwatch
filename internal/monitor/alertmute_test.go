package monitor

import (
	"testing"
	"time"
)

func TestAlertMute_NotMutedByDefault(t *testing.T) {
	m := NewAlertMute()
	if m.IsMuted(8080) {
		t.Error("expected port 8080 to not be muted by default")
	}
}

func TestAlertMute_MuteAndCheck(t *testing.T) {
	m := NewAlertMute()
	m.Mute(8080, 10*time.Minute)
	if !m.IsMuted(8080) {
		t.Error("expected port 8080 to be muted")
	}
}

func TestAlertMute_UnmuteRemovesMute(t *testing.T) {
	m := NewAlertMute()
	m.Mute(8080, 10*time.Minute)
	m.Unmute(8080)
	if m.IsMuted(8080) {
		t.Error("expected port 8080 to be unmuted after Unmute()")
	}
}

func TestAlertMute_ExpiredMuteReturnsFalse(t *testing.T) {
	now := time.Now()
	m := &AlertMute{
		mutes: make(map[int]time.Time),
		now:   func() time.Time { return now },
	}
	m.mutes[9090] = now.Add(-1 * time.Second) // already expired
	if m.IsMuted(9090) {
		t.Error("expected expired mute to return false")
	}
}

func TestAlertMute_ZeroDurationIsNoOp(t *testing.T) {
	m := NewAlertMute()
	m.Mute(8080, 0)
	if m.IsMuted(8080) {
		t.Error("expected zero-duration mute to have no effect")
	}
}

func TestAlertMute_DifferentPortsAreIndependent(t *testing.T) {
	m := NewAlertMute()
	m.Mute(8080, 10*time.Minute)
	if m.IsMuted(9090) {
		t.Error("expected port 9090 to not be muted when only 8080 is muted")
	}
}

func TestAlertMute_MutedPortsReturnsActive(t *testing.T) {
	now := time.Now()
	m := &AlertMute{
		mutes: make(map[int]time.Time),
		now:   func() time.Time { return now },
	}
	m.mutes[8080] = now.Add(5 * time.Minute)
	m.mutes[9090] = now.Add(-1 * time.Second) // expired

	active := m.MutedPorts()
	if _, ok := active[8080]; !ok {
		t.Error("expected port 8080 in MutedPorts result")
	}
	if _, ok := active[9090]; ok {
		t.Error("expected expired port 9090 to be excluded from MutedPorts result")
	}
}

func TestAlertMute_MuteExtendsDuration(t *testing.T) {
	now := time.Now()
	m := &AlertMute{
		mutes: make(map[int]time.Time),
		now:   func() time.Time { return now },
	}
	m.Mute(8080, 1*time.Minute)
	first := m.mutes[8080]
	m.Mute(8080, 10*time.Minute)
	second := m.mutes[8080]
	if !second.After(first) {
		t.Error("expected second Mute call to extend the expiry")
	}
}
