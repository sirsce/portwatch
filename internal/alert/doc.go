// Package alert implements notification delivery for portwatch.
//
// It defines the Notifier interface and concrete implementations
// for sending alerts when monitored port states change.
//
// Supported notifiers:
//   - WebhookNotifier: sends a JSON POST request to a configured URL.
//
// Usage:
//
//	notifier := alert.NewWebhookNotifier("https://example.com/hook")
//	err := notifier.Notify(alert.Event{
//		Host:      "localhost",
//		Port:      8080,
//		Status:    "closed",
//		Timestamp: time.Now(),
//		Message:   "port 8080 is no longer reachable",
//	})
package alert
