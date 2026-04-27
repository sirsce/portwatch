// Package monitor ties together the scanner and alert subsystems,
// periodically checking configured ports and dispatching notifications
// when their state changes.
package monitor

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/scanner"
)

// Notifier is the common interface satisfied by both WebhookNotifier and
// EmailNotifier.
type Notifier interface {
	Notify(subject, body string) error
}

// Monitor holds runtime state for the port-watching loop.
type Monitor struct {
	cfg       *config.Config
	scanner   *scanner.Scanner
	notifiers []Notifier
	// portState tracks the last known open/closed state per port.
	portState map[int]bool
}

// New constructs a Monitor from the supplied configuration.
func New(cfg *config.Config) *Monitor {
	var notifiers []Notifier

	if cfg.Webhook.URL != "" {
		notifiers = append(notifiers, alert.NewWebhookNotifier(cfg.Webhook.URL))
	}
	if cfg.Email.SMTPHost != "" {
		notifiers = append(notifiers, alert.NewEmailNotifier(
			cfg.Email.SMTPHost,
			cfg.Email.SMTPPort,
			cfg.Email.From,
			cfg.Email.To,
		))
	}

	return &Monitor{
		cfg:       cfg,
		scanner:   scanner.New(cfg.ScanTimeout),
		notifiers: notifiers,
		portState: make(map[int]bool),
	}
}

// Run starts the monitoring loop and blocks until ctx is cancelled.
func (m *Monitor) Run(ctx context.Context) {
	ticker := time.NewTicker(m.cfg.ScanInterval)
	defer ticker.Stop()

	// Run an immediate scan before waiting for the first tick.
	m.scan()

	for {
		select {
		case <-ticker.C:
			m.scan()
		case <-ctx.Done():
			log.Println("monitor: shutting down")
			return
		}
	}
}

// scan checks every configured port and fires alerts on state transitions.
func (m *Monitor) scan() {
	results := m.scanner.ScanPorts(m.cfg.Host, m.cfg.Ports)

	for _, r := range results {
		prev, seen := m.portState[r.Port]
		if seen && prev == r.Open {
			continue // no change
		}
		m.portState[r.Port] = r.Open
		if !seen {
			continue // skip notification on first scan
		}

		var subject, body string
		if r.Open {
			subject = "Port opened"
			body = "Port " + itoa(r.Port) + " on " + m.cfg.Host + " is now OPEN."
		} else {
			subject = "Port closed"
			body = "Port " + itoa(r.Port) + " on " + m.cfg.Host + " is now CLOSED."
		}

		for _, n := range m.notifiers {
			if err := n.Notify(subject, body); err != nil {
				log.Printf("monitor: notify error: %v", err)
			}
		}
	}
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
