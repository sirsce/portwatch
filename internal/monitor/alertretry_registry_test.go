package monitor

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestAlertRetry_ConcurrentNotify ensures AlertRetry is safe under
// concurrent use (each goroutine owns its own call sequence).
func TestAlertRetry_ConcurrentNotify(t *testing.T) {
	const goroutines = 20
	var totalCalls int32
	n := &funcNotifier{fn: func() error {
		atomic.AddInt32(&totalCalls, 1)
		return nil
	}}
	ar := NewAlertRetry(n, 3, time.Millisecond)
	ar.sleep = noSleep

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ar.Notify("subject", "body")
		}()
	}
	wg.Wait()

	if int(atomic.LoadInt32(&totalCalls)) != goroutines {
		t.Fatalf("expected %d calls, got %d", goroutines, totalCalls)
	}
}

// TestAlertRetry_BackoffDoublesEachTime checks that sleep durations double.
func TestAlertRetry_BackoffDoublesEachTime(t *testing.T) {
	n := &fakeNotifier{failFirst: 4}
	ar := NewAlertRetry(n, 5, 10*time.Millisecond)
	var durations []time.Duration
	ar.sleep = func(d time.Duration) { durations = append(durations, d) }
	_ = ar.Notify("s", "b")

	expected := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		40 * time.Millisecond,
		80 * time.Millisecond,
	}
	if len(durations) != len(expected) {
		t.Fatalf("expected %d sleep calls, got %d", len(expected), len(durations))
	}
	for i, d := range durations {
		if d != expected[i] {
			t.Errorf("sleep[%d]: want %v, got %v", i, expected[i], d)
		}
	}
}

// funcNotifier lets tests supply an arbitrary Notify implementation.
type funcNotifier struct{ fn func() error }

func (f *funcNotifier) Notify(_, _ string) error { return f.fn() }
