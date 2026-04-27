package monitor

import (
	"strings"
	"testing"
	"time"
)

func TestReport_OpenAndClosedCount(t *testing.T) {
	r := &Report{
		Host:      "localhost",
		Generated: time.Now(),
		Ports: []PortStatus{
			{Port: 80, Open: true},
			{Port: 443, Open: true},
			{Port: 8080, Open: false},
		},
	}

	if got := r.OpenCount(); got != 2 {
		t.Errorf("OpenCount() = %d, want 2", got)
	}
	if got := r.ClosedCount(); got != 1 {
		t.Errorf("ClosedCount() = %d, want 1", got)
	}
}

func TestReport_Summary(t *testing.T) {
	r := &Report{
		Host:      "myhost",
		Generated: time.Now(),
		Ports: []PortStatus{
			{Port: 22, Open: true},
			{Port: 9090, Open: false},
		},
	}

	summary := r.Summary()
	if !strings.Contains(summary, "host=myhost") {
		t.Errorf("Summary missing host: %s", summary)
	}
	if !strings.Contains(summary, "open=1") {
		t.Errorf("Summary missing open count: %s", summary)
	}
	if !strings.Contains(summary, "closed=1") {
		t.Errorf("Summary missing closed count: %s", summary)
	}
	if !strings.Contains(summary, "22=open") {
		t.Errorf("Summary missing port 22 state: %s", summary)
	}
	if !strings.Contains(summary, "9090=closed") {
		t.Errorf("Summary missing port 9090 state: %s", summary)
	}
}

func TestBuildReport_UsesHistory(t *testing.T) {
	h := NewHistory(10)
	h.Record("80", true)
	h.Record("443", false)

	report := BuildReport("testhost", []int{80, 443, 9999}, h)

	if report.Host != "testhost" {
		t.Errorf("Host = %q, want testhost", report.Host)
	}
	if len(report.Ports) != 3 {
		t.Fatalf("expected 3 port statuses, got %d", len(report.Ports))
	}

	byPort := map[int]bool{}
	for _, p := range report.Ports {
		byPort[p.Port] = p.Open
	}

	if !byPort[80] {
		t.Error("port 80 should be open")
	}
	if byPort[443] {
		t.Error("port 443 should be closed")
	}
	if byPort[9999] {
		t.Error("port 9999 should be closed (no history)")
	}
}

func TestBuildReport_EmptyPorts(t *testing.T) {
	h := NewHistory(10)
	report := BuildReport("empty", []int{}, h)

	if len(report.Ports) != 0 {
		t.Errorf("expected 0 ports, got %d", len(report.Ports))
	}
	if report.OpenCount() != 0 || report.ClosedCount() != 0 {
		t.Error("expected zero counts for empty report")
	}
}
