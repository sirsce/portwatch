package monitor

import (
	"testing"
	"time"
)

func TestNewRateLimiter_PanicsOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero minGap")
		}
	}()
	NewRateLimiter(0)
}

func TestNewRateLimiter_PanicsOnNegative(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for negative minGap")
		}
	}()
	NewRateLimiter(-time.Second)
}

func TestRateLimiter_FirstCallAlwaysAllowed(t *testing.T) {
	rl := NewRateLimiter(time.Minute)
	now := time.Now()
	if !rl.Allow(8080, now) {
		t.Error("first call should be allowed")
	}
}

func TestRateLimiter_SecondCallWithinGapDenied(t *testing.T) {
	rl := NewRateLimiter(time.Minute)
	now := time.Now()
	rl.Allow(8080, now)
	if rl.Allow(8080, now.Add(30*time.Second)) {
		t.Error("second call within gap should be denied")
	}
}

func TestRateLimiter_CallAfterGapAllowed(t *testing.T) {
	rl := NewRateLimiter(time.Minute)
	now := time.Now()
	rl.Allow(8080, now)
	if !rl.Allow(8080, now.Add(61*time.Second)) {
		t.Error("call after gap should be allowed")
	}
}

func TestRateLimiter_IndependentPerPort(t *testing.T) {
	rl := NewRateLimiter(time.Minute)
	now := time.Now()
	rl.Allow(8080, now)
	// different port should be allowed immediately
	if !rl.Allow(9090, now.Add(time.Second)) {
		t.Error("different port should be independently allowed")
	}
}

func TestRateLimiter_ResetAllowsImmediately(t *testing.T) {
	rl := NewRateLimiter(time.Minute)
	now := time.Now()
	rl.Allow(8080, now)
	rl.Reset(8080)
	if !rl.Allow(8080, now.Add(time.Second)) {
		t.Error("after Reset, port should be allowed again")
	}
}

func TestRateLimiter_ResetAll(t *testing.T) {
	rl := NewRateLimiter(time.Minute)
	now := time.Now()
	rl.Allow(8080, now)
	rl.Allow(9090, now)
	rl.ResetAll()
	if !rl.Allow(8080, now.Add(time.Second)) {
		t.Error("after ResetAll, port 8080 should be allowed")
	}
	if !rl.Allow(9090, now.Add(time.Second)) {
		t.Error("after ResetAll, port 9090 should be allowed")
	}
}
