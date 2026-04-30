package monitor

import (
	"errors"
	"testing"
	"time"
)

// fakeNotifier records call count and can be told to fail N times.
type fakeNotifier struct {
	failFirst int
	calls     int
}

func (f *fakeNotifier) Notify(_, _ string) error {
	f.calls++
	if f.calls <= f.failFirst {
		return errors.New("transient error")
	}
	return nil
}

func noSleep(_ time.Duration) {}

func TestNewAlertRetry_PanicsOnZeroMaxTries(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewAlertRetry(&fakeNotifier{}, 0, time.Millisecond)
}

func TestNewAlertRetry_PanicsOnZeroDelay(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewAlertRetry(&fakeNotifier{}, 3, 0)
}

func TestAlertRetry_SucceedsOnFirstAttempt(t *testing.T) {
	n := &fakeNotifier{failFirst: 0}
	ar := NewAlertRetry(n, 3, time.Millisecond)
	ar.sleep = noSleep
	if err := ar.Notify("subj", "body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.calls != 1 {
		t.Fatalf("expected 1 call, got %d", n.calls)
	}
}

func TestAlertRetry_RetriesAndSucceeds(t *testing.T) {
	n := &fakeNotifier{failFirst: 2}
	ar := NewAlertRetry(n, 5, time.Millisecond)
	ar.sleep = noSleep
	if err := ar.Notify("subj", "body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", n.calls)
	}
}

func TestAlertRetry_ExhaustsAllAttempts(t *testing.T) {
	n := &fakeNotifier{failFirst: 10}
	ar := NewAlertRetry(n, 3, time.Millisecond)
	ar.sleep = noSleep
	if err := ar.Notify("subj", "body"); err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if n.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", n.calls)
	}
}

func TestAlertRetry_SleepCalledBetweenAttempts(t *testing.T) {
	n := &fakeNotifier{failFirst: 3}
	ar := NewAlertRetry(n, 4, 10*time.Millisecond)
	sleepCount := 0
	ar.sleep = func(_ time.Duration) { sleepCount++ }
	_ = ar.Notify("subj", "body")
	// 4 attempts → 3 sleeps between them
	if sleepCount != 3 {
		t.Fatalf("expected 3 sleeps, got %d", sleepCount)
	}
}
