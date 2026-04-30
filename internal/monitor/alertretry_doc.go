// Package monitor provides runtime monitoring primitives for portwatch.
//
// # AlertRetry
//
// AlertRetry wraps any Notifier and adds automatic retry logic with
// exponential back-off. It is useful when the underlying transport
// (webhook or SMTP) may experience transient failures.
//
// Usage:
//
//	base := alert.NewWebhookNotifier(webhookURL)
//	retrier := monitor.NewAlertRetry(base, 3, 500*time.Millisecond)
//	// retrier.Notify will try up to 3 times: 500 ms, 1 s between attempts.
//
// AlertRetry composes naturally with AlertFilter, AlertThrottle, and the
// rest of the alert pipeline — wrap the innermost Notifier before passing
// it to higher-level components.
package monitor
