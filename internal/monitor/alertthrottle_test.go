package monitor

import (
	"testing"
	"time"
)

func TestNewAlertThrottle_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero window")
		}
	}()
	NewAlertThrottle(0, 3)
}

func TestNewAlertThrottle_PanicsOnZeroBurst(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero maxBurst")
		}
	}()
	NewAlertThrottle(time.Minute, 0)
}

func TestAlertThrottle_FirstCallAlwaysAllowed(t *testing.T) {
	th := NewAlertThrottle(time.Minute, 3)
	if !th.Allow(8080) {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAlertThrottle_BurstLimitEnforced(t *testing.T) {
	th := NewAlertThrottle(time.Minute, 2)
	if !th.Allow(9000) {
		t.Fatal("call 1 should be allowed")
	}
	if !th.Allow(9000) {
		t.Fatal("call 2 should be allowed")
	}
	if th.Allow(9000) {
		t.Fatal("call 3 should be suppressed (burst limit reached)")
	}
}

func TestAlertThrottle_DifferentPortsAreIndependent(t *testing.T) {
	th := NewAlertThrottle(time.Minute, 1)
	if !th.Allow(80) {
		t.Fatal("port 80 call 1 should be allowed")
	}
	if th.Allow(80) {
		t.Fatal("port 80 call 2 should be suppressed")
	}
	if !th.Allow(443) {
		t.Fatal("port 443 call 1 should be allowed independently")
	}
}

func TestAlertThrottle_AllowedAfterWindowExpires(t *testing.T) {
	var fakeNow time.Time
	th := NewAlertThrottle(50*time.Millisecond, 1)
	th.nowFn = func() time.Time { return fakeNow }

	fakeNow = time.Unix(1000, 0)
	if !th.Allow(22) {
		t.Fatal("first call should be allowed")
	}
	if th.Allow(22) {
		t.Fatal("second call within window should be suppressed")
	}

	// Advance time beyond the window.
	fakeNow = fakeNow.Add(100 * time.Millisecond)
	if !th.Allow(22) {
		t.Fatal("call after window expiry should be allowed")
	}
}

func TestAlertThrottle_Count(t *testing.T) {
	th := NewAlertThrottle(time.Minute, 5)
	th.Allow(3306)
	th.Allow(3306)
	th.Allow(3306)
	if got := th.Count(3306); got != 3 {
		t.Fatalf("expected count 3, got %d", got)
	}
}

func TestAlertThrottle_Reset(t *testing.T) {
	th := NewAlertThrottle(time.Minute, 1)
	th.Allow(5432)
	if th.Allow(5432) {
		t.Fatal("should be suppressed before reset")
	}
	th.Reset(5432)
	if !th.Allow(5432) {
		t.Fatal("should be allowed after reset")
	}
}
