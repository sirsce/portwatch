package monitor

import (
	"fmt"
	"strings"
	"time"
)

// PortStatus represents the current status of a monitored port.
type PortStatus struct {
	Port    int
	Open    bool
	Checked time.Time
}

// Report summarises the current state of all monitored ports.
type Report struct {
	Host      string
	Generated time.Time
	Ports     []PortStatus
}

// OpenCount returns the number of ports currently open.
func (r *Report) OpenCount() int {
	count := 0
	for _, p := range r.Ports {
		if p.Open {
			count++
		}
	}
	return count
}

// ClosedCount returns the number of ports currently closed.
func (r *Report) ClosedCount() int {
	return len(r.Ports) - r.OpenCount()
}

// Summary returns a human-readable one-line summary of the report.
func (r *Report) Summary() string {
	var parts []string
	for _, p := range r.Ports {
		state := "closed"
		if p.Open {
			state = "open"
		}
		parts = append(parts, fmt.Sprintf("%d=%s", p.Port, state))
	}
	return fmt.Sprintf("host=%s open=%d closed=%d ports=[%s]",
		r.Host, r.OpenCount(), r.ClosedCount(), strings.Join(parts, ","))
}

// BuildReport constructs a Report from the monitor's current history snapshot.
func BuildReport(host string, ports []int, h *History) *Report {
	now := time.Now()
	statuses := make([]PortStatus, 0, len(ports))
	for _, port := range ports {
		key := fmt.Sprintf("%d", port)
		latest, ok := h.Latest(key)
		open := ok && latest
		statuses = append(statuses, PortStatus{
			Port:    port,
			Open:    open,
			Checked: now,
		})
	}
	return &Report{
		Host:      host,
		Generated: now,
		Ports:     statuses,
	}
}
