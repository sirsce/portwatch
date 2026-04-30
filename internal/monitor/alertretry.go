package monitor

import (
	"errors"
	"time"
)

// AlertRetry wraps a Notifier and retries failed deliveries up to a
// configurable number of attempts with an exponential back-off.
type AlertRetry struct {
	notifier  Notifier
	maxTries  int
	baseDelay time.Duration
	sleep     func(time.Duration) // injectable for tests
}

// Notifier is the minimal interface required by AlertRetry.
type Notifier interface {
	Notify(subject, body string) error
}

// NewAlertRetry creates an AlertRetry.
// maxTries must be >= 1; baseDelay must be > 0.
func NewAlertRetry(n Notifier, maxTries int, baseDelay time.Duration) *AlertRetry {
	if maxTries < 1 {
		panic("alertretry: maxTries must be >= 1")
	}
	if baseDelay <= 0 {
		panic("alertretry: baseDelay must be > 0")
	}
	return &AlertRetry{
		notifier:  n,
		maxTries:  maxTries,
		baseDelay: baseDelay,
		sleep:     time.Sleep,
	}
}

// Notify attempts delivery, retrying with exponential back-off on failure.
// It returns the last error if all attempts are exhausted.
func (r *AlertRetry) Notify(subject, body string) error {
	var err error
	delay := r.baseDelay
	for attempt := 1; attempt <= r.maxTries; attempt++ {
		if err = r.notifier.Notify(subject, body); err == nil {
			return nil
		}
		if attempt < r.maxTries {
			r.sleep(delay)
			delay *= 2
		}
	}
	return errors.New("alertretry: all attempts failed: " + err.Error())
}
