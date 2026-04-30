package monitor

import (
	"testing"
	"time"
)

func TestNewAlertDedup_PanicsOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero window")
		}
	}()
	NewAlertDedup(0)
}

func TestNewAlertDedup_PanicsOnNegative(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on negative window")
		}
	}()
	NewAlertDedup(-time.Second)
}

func TestAlertDedup_FirstCallNotDuplicate(t *testing.T) {
	d := NewAlertDedup(time.Minute)
	if d.IsDuplicate(8080, "closed") {
		t.Fatal("first call should not be a duplicate")
	}
}

func TestAlertDedup_SamePortAndStateWithinWindow(t *testing.T) {
	d := NewAlertDedup(time.Minute)
	d.IsDuplicate(8080, "closed")
	if !d.IsDuplicate(8080, "closed") {
		t.Fatal("second call with same port/state within window should be duplicate")
	}
}

func TestAlertDedup_DifferentStateNotDuplicate(t *testing.T) {
	d := NewAlertDedup(time.Minute)
	d.IsDuplicate(8080, "closed")
	if d.IsDuplicate(8080, "open") {
		t.Fatal("different state should not be a duplicate")
	}
}

func TestAlertDedup_DifferentPortsAreIndependent(t *testing.T) {
	d := NewAlertDedup(time.Minute)
	d.IsDuplicate(8080, "closed")
	if d.IsDuplicate(9090, "closed") {
		t.Fatal("different port should not be a duplicate")
	}
}

func TestAlertDedup_ExpiredWindowNotDuplicate(t *testing.T) {
	now := time.Now()
	d := NewAlertDedup(time.Second)
	d.now = func() time.Time { return now }
	d.IsDuplicate(8080, "closed")

	d.now = func() time.Time { return now.Add(2 * time.Second) }
	if d.IsDuplicate(8080, "closed") {
		t.Fatal("call after window expiry should not be duplicate")
	}
}

func TestAlertDedup_Reset(t *testing.T) {
	d := NewAlertDedup(time.Minute)
	d.IsDuplicate(8080, "closed")
	d.Reset(8080)
	if d.IsDuplicate(8080, "closed") {
		t.Fatal("after reset, call should not be duplicate")
	}
}

func TestAlertDedup_Len(t *testing.T) {
	d := NewAlertDedup(time.Minute)
	if d.Len() != 0 {
		t.Fatalf("expected 0, got %d", d.Len())
	}
	d.IsDuplicate(8080, "closed")
	d.IsDuplicate(9090, "open")
	if d.Len() != 2 {
		t.Fatalf("expected 2, got %d", d.Len())
	}
}
