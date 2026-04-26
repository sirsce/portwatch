// Package alert provides mechanisms for sending notifications
// via webhook or email when port/service state changes are detected.
package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Event represents a port state change event to be alerted on.
type Event struct {
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Status    string    `json:"status"` // "open" or "closed"
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// Notifier defines the interface for sending alerts.
type Notifier interface {
	Notify(event Event) error
}

// WebhookNotifier sends alert events as JSON POST requests to a URL.
type WebhookNotifier struct {
	URL        string
	HTTPClient *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier with a default HTTP client.
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL: url,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Notify marshals the event to JSON and POSTs it to the configured URL.
func (w *WebhookNotifier) Notify(event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("alert: failed to marshal event: %w", err)
	}

	resp, err := w.HTTPClient.Post(w.URL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("alert: webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("alert: webhook returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
