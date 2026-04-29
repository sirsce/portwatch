package monitor

import (
	"strings"
	"testing"
	"time"
)

func TestNewAlertSummary_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero window")
		}
	}()
	log := NewStateChangeLog(10)
	NewAlertSummary(0, log)
}

func TestNewAlertSummary_PanicsOnNilLog(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil log")
		}
	}()
	NewAlertSummary(time.Minute, nil)
}

func TestAlertSummary_DigestEmpty(t *testing.T) {
	log := NewStateChangeLog(10)
	s := NewAlertSummary(time.Minute, log)
	entries := s.Digest(time.Now())
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestAlertSummary_DigestWithinWindow(t *testing.T) {
	log := NewStateChangeLog(10)
	now := time.Now()
	log.Record(StateChange{Port: 80, OldState: "open", NewState: "closed", At: now.Add(-30 * time.Second)})
	log.Record(StateChange{Port: 443, OldState: "closed", NewState: "open", At: now.Add(-10 * time.Second)})

	s := NewAlertSummary(time.Minute, log)
	entries := s.Digest(now)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Port != 80 || entries[1].Port != 443 {
		t.Errorf("unexpected ports: %v", entries)
	}
}

func TestAlertSummary_DigestExcludesOldEntries(t *testing.T) {
	log := NewStateChangeLog(10)
	now := time.Now()
	log.Record(StateChange{Port: 80, OldState: "open", NewState: "closed", At: now.Add(-2 * time.Minute)})
	log.Record(StateChange{Port: 443, OldState: "closed", NewState: "open", At: now.Add(-10 * time.Second)})

	s := NewAlertSummary(time.Minute, log)
	entries := s.Digest(now)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Port != 443 {
		t.Errorf("expected port 443, got %d", entries[0].Port)
	}
}

func TestAlertSummary_FormatNoEntries(t *testing.T) {
	log := NewStateChangeLog(10)
	s := NewAlertSummary(time.Minute, log)
	out := s.Format(time.Now())
	if !strings.Contains(out, "No state changes") {
		t.Errorf("unexpected format output: %q", out)
	}
}

func TestAlertSummary_FormatWithEntries(t *testing.T) {
	log := NewStateChangeLog(10)
	now := time.Now()
	log.Record(StateChange{Port: 8080, OldState: "open", NewState: "closed", At: now.Add(-5 * time.Second)})

	s := NewAlertSummary(time.Minute, log)
	out := s.Format(now)
	if !strings.Contains(out, "8080") {
		t.Errorf("expected port 8080 in output: %q", out)
	}
	if !strings.Contains(out, "open -> closed") {
		t.Errorf("expected transition in output: %q", out)
	}
	if !strings.Contains(out, "1 total") {
		t.Errorf("expected count in output: %q", out)
	}
}
