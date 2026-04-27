package monitor

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewTicker_PanicsOnZeroInterval(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for zero interval, got none")
		}
	}()
	NewTicker(0)
}

func TestNewTicker_PanicsOnNegativeInterval(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for negative interval, got none")
		}
	}()
	NewTicker(-time.Second)
}

func TestTicker_Interval(t *testing.T) {
	tk := NewTicker(500 * time.Millisecond)
	defer tk.Stop()

	if tk.Interval() != 500*time.Millisecond {
		t.Errorf("expected 500ms, got %v", tk.Interval())
	}
}

func TestTicker_FiresAtLeastOnce(t *testing.T) {
	tk := NewTicker(20 * time.Millisecond)
	defer tk.Stop()

	select {
	case <-tk.C:
		// success
	case <-time.After(200 * time.Millisecond):
		t.Error("ticker did not fire within timeout")
	}
}

func TestRunEvery_CallsFnMultipleTimes(t *testing.T) {
	var count atomic.Int32

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	RunEvery(ctx, 30*time.Millisecond, func(_ context.Context) {
		count.Add(1)
	})

	got := count.Load()
	if got < 2 {
		t.Errorf("expected at least 2 calls, got %d", got)
	}
}

func TestRunEvery_StopsOnContextCancel(t *testing.T) {
	var count atomic.Int32

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		RunEvery(ctx, 20*time.Millisecond, func(_ context.Context) {
			count.Add(1)
		})
		close(done)
	}()

	time.Sleep(70 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// success
	case <-time.After(200 * time.Millisecond):
		t.Error("RunEvery did not stop after context cancellation")
	}
}
