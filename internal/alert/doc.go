// Package alert provides notification backends for portwatch.
//
// Supported notifiers:
//
//   - WebhookNotifier: sends JSON payloads to an HTTP/HTTPS endpoint.
//   - EmailNotifier: sends plain-text emails via SMTP.
//
// All notifiers share the same interface:
//
//	Notify(subject, message string) error
//
// Notifiers are configured through the top-level Config and are instantiated
// by the daemon at startup based on the loaded configuration.
package alert
