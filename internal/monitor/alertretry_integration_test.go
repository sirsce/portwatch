package monitor

import (
	"sync/atomic"
	"testing"
	"time"
)

// atomicNotifier is safe for concurrent use and fails the first N calls.
type atomicNotifier struct {
	failFirst int32
	calls     int32
}

func (a *atomicNotifier) Notify(_, _ string) error {
	c := atomic.AddInt32(&a.calls, 1)
	if c <= atomic.LoadInt32(&a.failFirst) {
		return &retryableError{"transient"}
	}
	return nil
}

type retryableError struct{ msg string }

func (e *retryableError) Error() string { return e.msg }

// TestAlertRetry_IntegrationWithAlertFilter verifies that AlertRetry and
// AlertFilter compose correctly: the filter suppresses duplicates while
// the retry layer handles transient failures transparently.
func TestAlertRetry_IntegrationWithAlertFilter(t *testing.T) {
	an := &atomicNotifier{failFirst: 1}
	retrier := NewAlertRetry(an, 3, time.Millisecond)
	retrier.sleep = noSleep

	filter := NewAlertFilter(500*time.Millisecond, func(port int, state string) error {
		return retrier.Notify("port changed", state)
	})

	// First call: notifier fails once then succeeds on retry.
	if err := filter.Check(8080, "closed"); err != nil {
		t.Fatalf("first call should succeed after retry: %v", err)
	}

	// Second call within cooldown: filter suppresses it, notifier not called again.
	callsBefore := atomic.LoadInt32(&an.calls)
	if err := filter.Check(8080, "closed"); err == nil {
		t.Fatal("expected suppression error from filter")
	}
	if atomic.LoadInt32(&an.calls) != callsBefore {
		t.Fatal("notifier should not be called during cooldown")
	}
}
